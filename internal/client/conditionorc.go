package client

import (
	"encoding/json"

	"github.com/google/uuid"
	coapi "github.com/metal-toolbox/conditionorc/pkg/api/v1/types"
	rivetsCondition "github.com/metal-toolbox/rivets/condition"
)

func (c *Client) CreateCondition(serverUUID uuid.UUID) error {
	params, err := json.Marshal(rivetsCondition.InventoryTaskParameters{
		AssetID:               serverUUID,
		CollectBiosCfg:        false,
		CollectFirwmareStatus: false,
		Method:                rivetsCondition.OutofbandInventory,
	})
	if err != nil {
		return err
	}

	conditionCreate := coapi.ConditionCreate {
		Exclusive: false,
		Parameters: params,
	}

	_, err = c.coClient.ServerConditionCreate(c.ctx, serverUUID, rivetsCondition.Inventory, conditionCreate)
	if err != nil {
		return err
	}

	return nil
}