package client

import (
	"github.com/metal-toolbox/fleet-scheduler/internal/model"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	// "github.com/sirupsen/logrus"
	fleetDBapi "go.hollow.sh/serverservice/pkg/api/v1"
)

func (c *Client) CreateConditionInventoryForAllServers() error {
	// Start thread to start collecting servers
	serverCh, concLimiter, err := c.GatherServersNonBlocking(model.ConcurrencyDefault) // TODO; Swap out conc default with actual
	if err != nil {
		return err
	}

	// Loop through servers and create conditions
	for server := range serverCh {
		c.logger.Logger.Info("Server UUID: ", server.UUID)

		err := c.CreateConditionInventory(server.UUID)
		if err != nil {
			c.logger.WithFields(logrus.Fields{
				"server": server.UUID,
			}).Logger.Error("Failed to create condition")
		}

		concLimiter.Release(1)
	}

	return nil
}

func (c *Client) GatherServersNonBlocking(page_size int) (chan fleetDBapi.Server, *semaphore.Weighted, error) {
	if c.fdbClient == nil {
		return nil, nil, ErrSsClientIsNil
	}

	serverCh := make(chan fleetDBapi.Server)
	concLimiter := semaphore.NewWeighted(int64(page_size * page_size))

	go func() {
		c.gatherServers(page_size, serverCh, concLimiter)
	}()

	return serverCh, concLimiter, nil
}
