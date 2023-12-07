package client

import (
	"encoding/json"

	"github.com/google/uuid"
	conditionOrcapi "github.com/metal-toolbox/conditionorc/pkg/api/v1/types"
	conditionrivets "github.com/metal-toolbox/rivets/condition"
)

func (c *Client) CreateConditionInventory(serverUUID uuid.UUID) error {
	params, err := json.Marshal(conditionrivets.InventoryTaskParameters{
		AssetID:               serverUUID,
		CollectBiosCfg:        false,
		CollectFirwmareStatus: false,
		Method:                conditionrivets.OutofbandInventory,
	})
	if err != nil {
		return err
	}

	conditionCreate := conditionOrcapi.ConditionCreate {
		Exclusive: false,
		Parameters: params,
	}

	_, err = c.coClient.ServerConditionCreate(c.ctx, serverUUID, conditionrivets.Inventory, conditionCreate)
	if err != nil {
		return err
	}

	return nil
}