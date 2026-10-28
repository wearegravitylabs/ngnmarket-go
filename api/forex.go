package api

import (
	"context"

	"github.com/wearegravitylabs/ngnmarket-go/model"
)

// GetForexRates returns the latest NGN exchange rates (Free plan) - https://docs.ngnmarket.com/api-reference/forex/current
func (c *Call) GetForexRates(ctx context.Context) (*model.ForexRates, error) {
	var out model.ForexRates
	meta, err := c.makeRequest(ctx, "/forex/current", nil, &out)
	if err != nil {
		return nil, err
	}
	out.Meta = meta
	return &out, nil
}
