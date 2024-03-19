package client

import (
	"github.com/metal-toolbox/fleet-scheduler/internal/metrics"
	fleetdbapi "github.com/metal-toolbox/fleetdb/pkg/api/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func (c *Client) CreateConditionInventoryForAllServers(pageSize int) error {
	// First page, use the response from it to figure out how many pages we have to loop through
	response, err := c.getServerPageAndCreateInventory(1, pageSize, 0)
	if err != nil {
		return err
	}
	totalPages := response.TotalPages

	// Now that we know how many pages to expect, lets loop through the rest of the pages
	for i := 2; i <= totalPages; i++ {
		_, err := c.getServerPageAndCreateInventory(i, pageSize, totalPages)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) getServerPageAndCreateInventory(pageIndex, pageSize, totalPages int) (*fleetdbapi.ServerResponse, error) {
	servers, response, err := c.getServerPage(pageSize, pageIndex)
	if err != nil {
		c.log.WithFields(logrus.Fields{
			"pageIndex":  pageIndex,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		}).Logger.Errorf("Failed to get page of servers, attempting to continue: %s", err.Error())

		metrics.FleetdbErrorCounter.With(
			prometheus.Labels{"errors": err.Error()},
		).Inc()

		return response, err
	}

	c.log.WithFields(logrus.Fields{
		"pageIndex":  pageIndex,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	}).Trace("Got server page")

	for i := range servers {
		err = c.CreateConditionInventory(servers[i].UUID)
		if err != nil {
			metrics.ConditionorcErrorCounter.With(
				prometheus.Labels{"errors": err.Error()},
			).Inc()
			return response, err
		}

		metrics.InventoryCounter.With(
			prometheus.Labels{},
		).Inc()
	}

	return response, nil
}
