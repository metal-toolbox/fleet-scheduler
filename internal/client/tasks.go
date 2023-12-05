package client

import (
	"github.com/metal-toolbox/fleet-scheduler/internal/model"
	// "github.com/sirupsen/logrus"
)

func (c* Client) CollectServers() error {
	// First make sure the endpoints we need are established
	err := c.initServerServiceClient()
	if err != nil {
		return err
	}
	// err = c.initConditionsClient()
	// if err != nil {
	// 	return err
	// }

	// Gather servers asynchronously
	serverCh, concLimiter, err := c.AsyncGatherServers(model.ConcurrencyDefault) // TODO; Swap out conc default with actual
	if err != nil {
		return err
	}

	// Loop through servers and create conditions
	for server := range(serverCh) {
		c.logger.Logger.Info("Server UUID: ", server.UUID)

		// err := c.CreateCondition(server.UUID)
		// if err != nil {
		// 	c.logger.WithFields(logrus.Fields{
		// 			"server": server.UUID,
		// 	}).Logger.Error("Failed to create condition")
		// }

		concLimiter.Release(1)
	}

	return nil
}