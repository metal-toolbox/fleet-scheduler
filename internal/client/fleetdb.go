package client

import (
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	fleetDBrivets "github.com/metal-toolbox/rivets/serverservice"
	fleetDBapi "go.hollow.sh/serverservice/pkg/api/v1"
)

func (c *Client) gatherServers(pageSize int, serverCh chan fleetDBapi.Server, concLimiter *semaphore.Weighted) {
	// signal to receiver that we are done
	defer close(serverCh)

	// First page, use the response from it to figure out how many pages we have to loop through
	// Dont change page size
	servers, response, err := c.getServerPage(pageSize, 1)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"pageSize":  pageSize,
			"pageIndex": 1,
		}).Logger.Error("Failed to get list of servers")
		return
	}
	totalPages := response.TotalPages

	if !concLimiter.TryAcquire(int64(response.PageSize)) {
		c.logger.Error("Failed to acquire semaphore! Going to attempt to continue.")
	}

	// send first page of servers to the channel
	for i := range servers {
		serverCh <- servers[i]
	}

	c.logger.WithFields(logrus.Fields{
		"index":      1,
		"iterations": totalPages,
		"got":        len(servers),
	}).Trace("Got server page")

	// Start the second page, and loop through rest the pages
	for i := 2; i <= totalPages; i++ {
		servers, response, err = c.getServerPage(pageSize, i)
		if err != nil {
			c.logger.WithFields(logrus.Fields{
				"pageSize":  pageSize,
				"pageIndex": i,
			}).Logger.Error("Failed to get page of servers")

			continue
		}

		c.logger.WithFields(logrus.Fields{
			"index":      i,
			"iterations": totalPages,
			"got":        len(servers),
		}).Trace("Got server page")

		// throttle this loop
		// Doing a spinlock to prevent a permanent lock if the ctx gets canceled
		// TODO; Kill thread if context is canceled?
		for !concLimiter.TryAcquire(int64(response.PageSize)) && c.ctx.Err() == nil {
			time.Sleep(time.Second)
		}

		for i := range servers {
			serverCh <- servers[i]
		}
	}
}

func (c *Client) getServerPage(pageSize, page int) ([]fleetDBapi.Server, *fleetDBapi.ServerResponse, error) {
	params := &fleetDBapi.ServerListParams{
		FacilityCode: c.cfg.FacilityCode,
		AttributeListParams: []fleetDBapi.AttributeListParams{
			{
				Namespace: fleetDBrivets.ServerAttributeNSBmcAddress,
			},
		},
		PaginationParams: &fleetDBapi.PaginationParams{
			Limit: pageSize,
			Page:  page,
		},
	}

	return c.ssClient.List(c.ctx, params)
}
