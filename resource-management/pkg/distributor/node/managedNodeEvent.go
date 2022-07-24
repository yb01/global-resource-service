package node

import (
	"k8s.io/klog/v2"
	"strconv"

	"global-resource-service/resource-management/pkg/common-lib/types"
)

// TODO - add more fields for minimal node record
type ManagedNodeEvent struct {
	nodeEvent *types.NodeEvent
	loc       *types.Location
}

func NewManagedNodeEvent(nodeEvent *types.NodeEvent, loc *types.Location) *ManagedNodeEvent {
	return &ManagedNodeEvent{
		nodeEvent: nodeEvent,
		loc:       loc,
	}
}

func (n *ManagedNodeEvent) GetId() string {
	return n.nodeEvent.Node.Id
}

func (n *ManagedNodeEvent) GetLocation() *types.Location {
	return n.loc
}

func (n *ManagedNodeEvent) GetRvLocation() *types.RvLocation {

	return &types.RvLocation{Region: n.loc.GetRegion(), Partition: n.loc.GetResourcePartition()}
}

func (n *ManagedNodeEvent) GetResourceVersion() uint64 {
	rv, err := strconv.ParseUint(n.nodeEvent.Node.ResourceVersion, 10, 64)
	if err != nil {
		klog.Errorf("Unable to convert resource version %s to uint64\n", n.nodeEvent.Node.ResourceVersion)
		return 0
	}
	return rv
}

func (n *ManagedNodeEvent) GetEventType() types.EventType {
	return n.nodeEvent.Type
}

func (n *ManagedNodeEvent) GetNodeEvent() *types.NodeEvent {
	return n.nodeEvent
}

func (n *ManagedNodeEvent) CopyNode() *types.LogicalNode {
	return n.nodeEvent.Node.Copy()
}
