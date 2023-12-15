package client

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/metal-toolbox/fleet-scheduler/internal/model"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	// "github.com/sirupsen/logrus"
	fleetDBapi "go.hollow.sh/serverservice/pkg/api/v1"
)

func (c* Client) CreateNewDummyServers(count uint) error {
	bmcUser := "bar"
	bmcPass := "foo"
	bmcAddress := "127.0.0.1"

	if !c.cfg.FdbCfg.DisableOAuth || !c.cfg.CoCfg.DisableOAuth {
		return errors.New("CreateNewDummyServer is only to be run in the sandbox!")
	}

	for i := uint(0); i < count; i++ {
		newServerUUID := uuid.New()
		newServer := fleetDBapi.Server{UUID: newServerUUID, Name: newServerUUID.String(), FacilityCode: c.cfg.FacilityCode}

		_, _, err := c.fdbClient.Create(c.ctx, newServer)
		if err != nil {
			return err
		}

		_, err = c.fdbClient.SetCredential(c.ctx, newServerUUID, "bmc", bmcUser, bmcPass)
		if err != nil {
			return err
		}

		addrAttr := fmt.Sprintf(`{"address": "%s"}`, bmcAddress)
		bmcIPAttr := fleetDBapi.Attributes{Namespace: "sh.hollow.bmc_info", Data: []byte(addrAttr)}
		_, err = c.fdbClient.CreateAttributes(c.ctx, newServerUUID, bmcIPAttr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c* Client) CreateConditionInventoryForAllServers() error {
	// Start thread to start collecting servers
	serverCh, concLimiter, err := c.GatherServersNonBlocking(model.ConcurrencyDefault) // TODO; Swap out conc default with actual
	if err != nil {
		return err
	}

	// Loop through servers and create conditions
	for server := range(serverCh) {
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
	concLimiter := semaphore.NewWeighted(int64(page_size*page_size))

	go func() {
		c.gatherServers(page_size, serverCh, concLimiter)
	}()

	return serverCh, concLimiter, nil
}