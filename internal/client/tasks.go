package client

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	// "github.com/sirupsen/logrus"
	fleetdbapi "github.com/metal-toolbox/fleetdb/pkg/api/v1"
)

func (c *Client) CreateConditionInventoryForAllServers() error {
	// Start thread to start collecting servers
	serverCh, concLimiter, err := c.GatherServersNonBlocking(c.cfg.Concurrency)
	if err != nil {
		return err
	}

	// Loop through servers and create conditions
	for server := range serverCh {
		err := c.CreateConditionInventory(server.UUID)
		if err != nil {
			c.log.WithFields(logrus.Fields{
				"server": server.UUID,
			}).Logger.Error("Failed to create condition")
		}

		concLimiter.Release(1)
	}

	return nil
}

func (c *Client) GatherServersNonBlocking(pageSize int) (chan *fleetdbapi.Server, *semaphore.Weighted, error) {
	serverCh := make(chan *fleetdbapi.Server)
	concLimiter := semaphore.NewWeighted(int64(pageSize * pageSize))

	go func() {
		c.gatherServers(pageSize, serverCh, concLimiter)
	}()

	return serverCh, concLimiter, nil
}
