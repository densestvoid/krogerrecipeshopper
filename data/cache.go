package data

import (
	"context"
	"encoding/json"
	"log/slog"
	"maps"
	"slices"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client     *redis.Client
	expiration time.Duration
}

func NewCache(client *redis.Client, expiration time.Duration) *Cache {
	return &Cache{
		client:     client,
		expiration: expiration,
	}
}

type CacheProduct struct {
	ProductID   string `json:"productID"`
	Brand       string `json:"brand"`
	Description string `json:"description"`
	Size        string `json:"size"`
}

func (c *Cache) StoreKrogerProduct(ctx context.Context, products ...CacheProduct) error {
	var productsJSON = map[string]any{}
	for _, product := range products {
		productJSON, err := json.Marshal(product)
		if err != nil {
			return err
		}
		productsJSON[product.ProductID] = string(productJSON)
	}

	_, err := c.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		if err := p.HSet(ctx, "products", productsJSON).Err(); err != nil {
			slog.Error("caching products", "error", err)
			return err
		}

		if err := p.HExpire(ctx, "products", c.expiration, slices.Collect(maps.Keys(productsJSON))...).Err(); err != nil {
			slog.Error("setting product cache expiration", "error", err)
			return err
		}

		return nil
	})
	return err
}

func (c *Cache) RetrieveKrogerProduct(ctx context.Context, productIDs ...string) ([]CacheProduct, []string, error) {
	values, err := c.client.HMGet(ctx, "products", productIDs...).Result()
	if err != nil {
		return nil, productIDs, err
	}

	var products []CacheProduct
	var productIDMisses []string
	for i, value := range values {
		if value == nil {
			productIDMisses = append(productIDMisses, productIDs[i])
			continue
		}

		var product CacheProduct
		if err := json.Unmarshal([]byte(value.(string)), &product); err != nil {
			productIDMisses = append(productIDMisses, productIDs[i])
			continue
		}

		products = append(products, product)
	}

	slog.Info("product cache", "misses", productIDMisses)
	return products, productIDMisses, nil
}

type CacheLocation struct {
	LocationID string `json:"locationID"`
	Name       string `json:"name"`
	Address    string `json:"address"`
}

func (c *Cache) StoreKrogerLocation(ctx context.Context, location CacheLocation) error {
	locationJSON, err := json.Marshal(location)
	if err != nil {
		return err
	}

	_, err = c.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		if err := p.HSet(ctx, "locations", location.LocationID, locationJSON).Err(); err != nil {
			slog.Error("caching location", "error", err)
			return err
		}

		if err := p.HExpire(ctx, "locations", c.expiration, location.LocationID).Err(); err != nil {
			slog.Error("setting location cache expiration", "error", err)
			return err
		}

		return nil
	})
	return err
}

func (c *Cache) RetrieveKrogerLocation(ctx context.Context, locationID string) (*CacheLocation, error) {
	values, err := c.client.HGet(ctx, "locations", locationID).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var location CacheLocation
	return &location, json.Unmarshal([]byte(values), &location)
}
