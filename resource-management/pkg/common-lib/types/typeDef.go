package types

import "time"

// for now, simply define those as string
// RegionName and ResourcePartitionName are updated to int per initial performance test of distributor ProcessEvents
// Later the data type might be changed back to string due to further performance evaluation result
type RegionName int
type ResourcePartitionName int
type DataCenterName string
type AvailabilityZoneName string
type FaultDomainName string
type NodeMachineType string

// TODO: from the Node definition in resource cluster, to the logicalNode struct, to the scheduler node_info structure
//       the ResourceName need to be set and aligned
type ResourceName string

// EventType defines the possible types of events.
type EventType string

type NodeGeoInfo struct {
	// Region and RsourcePartition are required
	Region            RegionName            `json:"region" protobuf:"bytes,1,opt,name=region"`
	ResourcePartition ResourcePartitionName `json:"rp" protobuf:"bytes,2,opt,name=rp"`

	// Optional fields for fine-tuned resource management and application placements
	DataCenter       DataCenterName       `json:"dc" protobuf:"bytes,3,opt,name=dc"`
	AvailabilityZone AvailabilityZoneName `json:"az" protobuf:"bytes,4,opt,name=az"`
	FaultDomain      FaultDomainName      `json:"fd" protobuf:"bytes,5,opt,name=fd"`
}

type NodeTaints struct {
	// Do not allow new pods to schedule onto the node unless they tolerate the taint,
	// Enforced by the scheduler.
	NoSchedule bool `json:"no_schedule" protobuf:"varint,1,opt,name=no_schedule"`
	// Evict any already-running pods that do not tolerate the taint
	NoExecute bool `json:"no_execute" protobuf:"varint,2,opt,name=no_execute"`
}

// TODO: consider refine for GPU types, such as NVIDIA and AMD etc.
type NodeSpecialHardWareTypeInfo struct {
	HasGpu  bool `json:"hasgpu" protobuf:"varint,1,opt,name=hasgpu"`
	HasFPGA bool `json:"hasfpga" protobuf:"varint,2,opt,name=hasfpga"`
}

// struct definition from Arktos node_info.go
type NodeResource struct {
	MilliCPU         int64 `json:"milli_cpu" protobuf:"varint,1,opt,name=milli_cpu"`
	Memory           int64 `json:"memory" protobuf:"varint,2,opt,name=memory"`
	EphemeralStorage int64 `json:"ephemeral_storage" protobuf:"varint,3,opt,name=ephemeral_storage"`
	// We store allowedPodNumber (which is Node.Status.Allocatable.Pods().Value())
	// explicitly as int, to avoid conversions and improve performance.
	AllowedPodNumber int32 `json:"allowed_pod_number" protobuf:"varint,4,opt,name=allowed_pod_number"`
	// ScalarResources such as GPU or FPGA etc.
	ScalarResources map[ResourceName]int64 `json:"scalar_resources" protobuf:"bytes,5,opt,name=scalar_resources"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LogicalNode is the abstraction of the node definition in the resource clusters
// LogicalNode is a minimum set of information the scheduler needs to place the workload to a node in the region-less platform
//
// Initial set of fields as shown below.
//
// TODO: add the annotation for serialization
//
type LogicalNode struct {
	// Node UUID from each resource partition cluster
	Id string `json:"id" protobuf:"bytes,1,opt,name=id"`

	// ResourceVersion is the RV from each resource partition cluster
	ResourceVersion string `json:"resource_version" protobuf:"bytes,2,opt,name=resource_version"`

	// GeoInfo defines the node location info such as region, DC, RP cluster etc. for application placement
	GeoInfo NodeGeoInfo `json:"geo_info" protobuf:"bytes,3,opt,name=geo_info"`

	// Taints defines scheduling or other control action for a node
	Taints NodeTaints `json:"taints" protobuf:"bytes,4,opt,name=taints"`

	// SpecialHardwareTypes defines if the node has special hardware such as GPU or FPGA etc
	SpecialHardwareTypes NodeSpecialHardWareTypeInfo `json:"special_hardware_types" protobuf:"bytes,5,opt,name=special_hardware_types"`

	// AllocatableReesource defines the resources on the node that can be used by schedulers
	AllocatableResource NodeResource `json:"allocatable_resource" protobuf:"bytes,6,opt,name=allocatable_resource"`

	// Conditions is a short version of the node condition array from Arktos, each bits in the byte defines one node condition
	Conditions int32 `json:"conditions" protobuf:"varint,7,opt,name=conditions"`

	// Reserved defines if the node is reserved at the resource partition cluster level
	// TBD Node reservation model for post 630
	Reserved bool `json:"reserved" protobuf:"varint,8,opt,name=reserved"`

	// MachineType defines the type of category of the node, such as # of CPUs of the node, where the category can be
	// defined as highend, lowend, medium as an example
	// TBD for post 630
	MachineType NodeMachineType `json:"machine_type" protobuf:"bytes,9,opt,name=machine_type"`

	// LastUpdatedTime defines the time when node status was updated in resource partition
	LastUpdatedTime Time `json:"last_updated_time" protobuf:"bytes,10,opt,name=last_updated_time"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RpNodeEvents is a struct for node events from each RP
type RpNodeEvents struct {
	NodeEvents []*NodeEvent `json:"node_events" protobuf:"bytes,1,rep,name=node_events"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RRM: Resource Region Manager
type ResponseFromRRM struct {
	RegionNodeEvents []RpNodeEvents `json:"region_node_events" protobuf:"bytes,1,rep,name=region_node_events"`
	Length           uint64         `json:"length" protobuf:"varint,3,opt,name=length"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeEvent is a event of nodes
type NodeEvent struct {
	Type EventType    `json:"id" protobuf:"bytes,1,opt,name=id"`
	Node *LogicalNode `json:"node" protobuf:"bytes,2,opt,name=node"`
	// +optional
	checkpoints []time.Time `protobuf:"-"` //`json:"checkpoints" protobuf:"bytes,3,rep,name=checkpoints"`
}
