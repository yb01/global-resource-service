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
type ResourceVersionMap map[RvLocation]uint64

func (rvs *ResourceVersionMap) Copy() ResourceVersionMap {
	dupRVs := make(ResourceVersionMap, len(*rvs))
	for loc, rv := range *rvs {
		dupRVs[loc] = rv
	}

	return dupRVs
}
