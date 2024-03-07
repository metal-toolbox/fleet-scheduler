package client

import (
	"github.com/sirupsen/logrus"
)

func (c *Client) CreateConditionInventoryForAllServers(pageSize int) error {
	// First page, use the response from it to figure out how many pages we have to loop through
	// Dont change page size
	servers, response, err := c.getServerPage(pageSize, 1)
	if err != nil {
		c.log.WithFields(logrus.Fields{
			"pageSize":  pageSize,
			"pageIndex": 1,
		}).Logger.Errorf("Failed to get list of servers: %s", err.Error())
		return err
	}
	totalPages := response.TotalPages

	// send first page of servers to the channel
	for i := range servers {
		err = c.CreateConditionInventory(servers[i].UUID)
		if err != nil {
			return err
		}
	}

	c.log.WithFields(logrus.Fields{
		"index":      1,
		"iterations": totalPages,
		"got":        len(servers),
	}).Trace("Got server page")

	// Start the second page, and loop through rest the pages
	for i := 2; i <= totalPages; i++ {
		servers, _, err = c.getServerPage(pageSize, i)
		if err != nil {
			c.log.WithFields(logrus.Fields{
				"pageSize":  pageSize,
				"pageIndex": i,
			}).Logger.Errorf("Failed to get page of servers, attempting to continue: %s", err.Error())

			continue
		}

		c.log.WithFields(logrus.Fields{
			"index":      i,
			"iterations": totalPages,
			"got":        len(servers),
		}).Trace("Got server page")

		for i := range servers {
			err = c.CreateConditionInventory(servers[i].UUID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
