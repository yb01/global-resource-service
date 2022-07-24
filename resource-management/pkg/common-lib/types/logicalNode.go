package types

import (
	"fmt"
	"k8s.io/klog/v2"
	"strconv"
)

const (
	PreserveNode_KeyPrefix = "MinNode"
)

func (n *LogicalNode) Copy() *LogicalNode {
	return &LogicalNode{
		Id:                   n.Id,
		ResourceVersion:      n.ResourceVersion,
		GeoInfo:              n.GeoInfo,
		Taints:               n.Taints,
		SpecialHardwareTypes: n.SpecialHardwareTypes,
		AllocatableResource:  n.AllocatableResource,
		Conditions:           n.Conditions,
		Reserved:             n.Reserved,
		MachineType:          n.MachineType,
		LastUpdatedTime:      n.LastUpdatedTime,
	}
}

func (n *LogicalNode) GetResourceVersionInt64() uint64 {
	rv, err := strconv.ParseUint(n.ResourceVersion, 10, 64)
	if err != nil {
		klog.Errorf("Unable to convert resource version %s to uint64\n", n.ResourceVersion)
		return 0
	}
	return rv
}

func (n *LogicalNode) GetKey() string {
	if n != nil {
		return fmt.Sprintf("%s.%s.%v.%v", PreserveNode_KeyPrefix, n.Id, n.GeoInfo.Region, n.GeoInfo.ResourcePartition)
	}
	return ""
}
