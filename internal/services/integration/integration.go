package integration

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/services"
)

type Client struct {
	Service services.DeviceDefinitionService
}

func (c *Client) ByTokenID(ctx context.Context, tokenID int) (Integration, error) {
	i, err := c.Service.GetIntegrationByTokenID(ctx, uint64(tokenID))
	if err != nil {
		return Integration{}, err
	}

	return Integration{ID: i.Id, Vendor: i.Vendor}, nil
}

type Integration struct {
	ID     string
	Vendor string
}
