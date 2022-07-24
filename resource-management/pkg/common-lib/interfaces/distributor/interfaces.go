package distributor

import (
	"global-resource-service/resource-management/pkg/common-lib/types"
)

type Interface interface {
	RegisterClient(*types.Client) error

	ListNodesForClient(clientId string) ([]*types.LogicalNode, types.TransitResourceVersionMap, error)
	Watch(clientId string, rvs types.TransitResourceVersionMap, watchChan chan *types.NodeEvent, stopCh chan struct{}) error
	ProcessEvents(events []*types.NodeEvent) (bool, types.TransitResourceVersionMap)
}
