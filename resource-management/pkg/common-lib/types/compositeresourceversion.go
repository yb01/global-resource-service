package types

import (
	"encoding/json"
)

type CompositeResourceVersion struct {
	RegionId            string
	ResourcePartitionId string
	ResourceVersion     uint64
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//RvLocation is used in rv map for rest apis
type RvLocation struct {
	Region    Region
	Partition ResourcePartition
}

func (loc RvLocation) MarshalText() (text []byte, err error) {
	type l RvLocation
	return json.Marshal(l(loc))
}

func (loc *RvLocation) UnmarshalText(text []byte) error {
	type l RvLocation
	return json.Unmarshal(text, (*l)(loc))
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Map from (regionId, ResourcePartitionId) to resourceVersion
// used in REST API calls
type TransitResourceVersionMap map[RvLocation]uint64

// internally used in the eventqueue used in WATCH of nodes
type InternalResourceVersionMap map[Location]uint64

func ConvertToInternalResourceVersionMap(rvs TransitResourceVersionMap) InternalResourceVersionMap {
	internalMap := make(InternalResourceVersionMap)

	for k, v := range rvs {
		internalMap[*NewLocation(k.Region, k.Partition)] = v
	}

	return internalMap
}

func (rvs *TransitResourceVersionMap) Copy() TransitResourceVersionMap {
	dupRVs := make(TransitResourceVersionMap, len(*rvs))
	for loc, rv := range *rvs {
		dupRVs[loc] = rv
	}

	return dupRVs
}
