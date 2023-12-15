package client

import (
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	fleetDBrivets "github.com/metal-toolbox/rivets/serverservice"
	fleetDBapi "go.hollow.sh/serverservice/pkg/api/v1"
)

func (c* Client) gatherServers(page_size int, serverCh chan fleetDBapi.Server, concLimiter *semaphore.Weighted) {
	// signal to reciever that we are done
	defer close(serverCh)

	// First page, use the response from it to figure out how many pages we have to loop through
	// Dont change page size
	servers, response, err := c.getServerPage(page_size, 1)
	if err != nil {
		c.logger.WithFields(logrus.Fields {
				"page_size": page_size,
				"page_index": 1,
		}).Logger.Error("Failed to get list of servers")
		return
	}
	total_pages := response.TotalPages

	if !concLimiter.TryAcquire(int64(response.PageSize)) {
		c.logger.Error("Failed to acquire semaphore! Going to attempt to continue.")
	}

	// send first page of servers to the channel
	for _, server := range(servers) {
		serverCh <- server
	}

	c.logger.WithFields(logrus.Fields{
		"index":      1,
		"iterations": total_pages,
		"got"       : len(servers),
	}).Trace("Got server page")

	// Start the second page, and loop through rest the pages
	for i := 2; i <= total_pages; i++ {
		servers, response, err = c.getServerPage(page_size, i)
		if err != nil {
			c.logger.WithFields(logrus.Fields {
				"page_size": page_size,
				"page_index": i,
			}).Logger.Error("Failed to get page of servers")

			continue
		}

		c.logger.WithFields(logrus.Fields{
			"index":      i,
			"iterations": total_pages,
			"got"       : len(servers),
		}).Trace("Got server page")

		// throttle this loop
		// Doing a spinlock to prevent a permanent lock if the ctx gets canceled
		// TODO; Kill thread if context is canceled?
		for !concLimiter.TryAcquire(int64(response.PageSize)) && c.ctx.Err() == nil {
			time.Sleep(time.Second)
		}

		for _, server := range(servers) {
			serverCh <- server
		}
	}
}

func (c* Client) getServerPage(page_size int, page int) ([]fleetDBapi.Server, *fleetDBapi.ServerResponse, error) {
	params := &fleetDBapi.ServerListParams{
		FacilityCode: c.cfg.FacilityCode,
		AttributeListParams: []fleetDBapi.AttributeListParams{
			{
				Namespace: fleetDBrivets.ServerAttributeNSBmcAddress,
			},
		},
		PaginationParams: &fleetDBapi.PaginationParams{
			Limit: page_size,
			Page:  page,
		},
	}

	return c.fdbClient.List(c.ctx, params)
}