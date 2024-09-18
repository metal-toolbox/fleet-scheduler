package client

import (
	fleetdbapi "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	fleetDBRivets "github.com/metal-toolbox/rivets/fleetdb"
)

func (c *Client) getServerPage(pageSize, page int) ([]fleetdbapi.Server, *fleetdbapi.ServerResponse, error) {
	params := &fleetdbapi.ServerListParams{
		FacilityCode: c.cfg.FacilityCode,
		AttributeListParams: []fleetdbapi.AttributeListParams{
			{
				Namespace: fleetDBRivets.ServerAttributeNSBmcAddress,
			},
		},
		PaginationParams: &fleetdbapi.PaginationParams{
			Limit: pageSize,
			Page:  page,
		},
	}

	return c.fdbClient.List(c.ctx, params)
}
