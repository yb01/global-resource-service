package types

import (
	"encoding/json"
	"global-resource-service/resource-management/pkg/common-lib/types/location"
)

type CompositeResourceVersion struct {
	RegionId            string
	ResourcePartitionId string
	ResourceVersion     uint64
}

type RvLocation struct {
	Region    location.Region
	Partition location.ResourcePartition
}

func (loc RvLocation) MarshalText() (text []byte, err error) {
	type l RvLocation
	return json.Marshal(l(loc))
}

func (loc *RvLocation) UnmarshalText(text []byte) error {
	type l RvLocation
	return json.Unmarshal(text, (*l)(loc))
}

// Map from (regionId, ResourcePartitionId) to resourceVersion
// used in REST API calls
type ResourceVersionMap map[RvLocation]uint64

// internally used in the eventqueue
type InternalResourceVersionMap map[location.Location]uint64

func ConvertToInternalResourceVersionMap(rvs ResourceVersionMap) InternalResourceVersionMap {
	internalMap := make(InternalResourceVersionMap)

	for k, v := range rvs {
		internalMap[*location.NewLocation(k.Region, k.Partition)] = v
	}

	return internalMap
}

func (rvs *ResourceVersionMap) Copy() ResourceVersionMap {
	dupRVs := make(ResourceVersionMap, len(*rvs))
	for loc, rv := range *rvs {
		dupRVs[loc] = rv
	}

	return dupRVs
}
