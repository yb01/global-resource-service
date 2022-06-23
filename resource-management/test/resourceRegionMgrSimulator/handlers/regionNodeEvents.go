package handlers

import (
	"net/http"

	"k8s.io/klog/v2"

	"global-resource-service/resource-management/test/resourceRegionMgrSimulator/data"
	simulatorTypes "global-resource-service/resource-management/test/resourceRegionMgrSimulator/types"
)

type RegionNodeEventHandler struct{}

// NewRegionNodeEvents creates a Region Node Events handler with the given logger
//
func NewRegionNodeEventsHander() *RegionNodeEventHandler {
	return &RegionNodeEventHandler{}
}

func (re *RegionNodeEventHandler) SimulatorHandler(rw http.ResponseWriter, r *http.Request) {

	klog.Infof("Handle /resources. URL path: %s", r.URL.Path)

	// Check URL Path received from aggregator client side
	if r.URL.Path == InitPullPath {
		klog.Info("Handle GET all region node added events via initPull")
	} else if r.URL.Path == SubsequentPullPath {
		klog.Info("Handle GET all region node modified events via SubsequentPull")
	} else if r.URL.Path == PostCRVPath {
		klog.Info("Handle POST CRV to discard all old region node modified events")
	} else {
		klog.Errorf("Error: The current URL (%v) is not supported!", r.URL.Path)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// Deserialize aggregator client request
	aggregatorClientReq := &simulatorTypes.PullDataFromRRM{}

	err := simulatorTypes.FromJSON(aggregatorClientReq, r.Body)

	if err != nil {
		klog.Errorf("Deserializing aggregator client request error : (%v)", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check CRV data received from aggregator client side
	if r.URL.Path == InitPullPath {
		if len(aggregatorClientReq.CRV) != 0 {
			klog.Error("Error: CRV should blank, but right now it is not blank")
			return
		}
	} else if r.URL.Path == SubsequentPullPath || r.URL.Path == PostCRVPath {
		if len(aggregatorClientReq.CRV) == 0 {
			klog.Error("Error: CRVs should not blank, but right now it is blank")
			return
		}
	} else {
		klog.Errorf("Error: The current URL (%v) is not supported!", r.URL.Path)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	//
	// Process aggregator client request based on three different URL PATH
	//
	// Process initpull or subsequentpull request
	//
	if r.URL.Path == InitPullPath || r.URL.Path == SubsequentPullPath {
		var nodeEvents simulatorTypes.RegionNodeEvents
		var count uint64

		if r.URL.Path == InitPullPath {
			nodeEvents, count = data.GetRegionNodeAddedEvents(aggregatorClientReq.BatchLength)
		} else if r.URL.Path == SubsequentPullPath {
			nodeEvents, count = data.GetRegionNodeModifiedEventsCRV(aggregatorClientReq.CRV)
		}

		if len(nodeEvents) == 0 {
			klog.Info("Pulling Region Node Events with batch is in the end")
		} else {
			klog.Infof("Pulling Region Node Event with final batch size (%v) for (%v) regions", count, len(nodeEvents))

			response := &simulatorTypes.ResponseFromRRM{
				RegionNodeEvents: nodeEvents,
				RvMap:            aggregatorClientReq.CRV,
				Length:           uint64(count),
			}

			// Serialize region node events result to JSON
			err = response.ToJSON(rw)

			if err != nil {
				klog.Errorf("Error - Unable to marshal json : ", err)
			}

			if r.URL.Path == InitPullPath {
				klog.Info("Handle GET all region node added events via initPull successfully")
			} else {
				klog.Info("Handle GET all region node modified events via SubsequentPull succesfully")
			}
		}

		// Process post CRV to discard all old region node modified event
		//
	} else if r.URL.Path == PostCRVPath {
		var postCRV simulatorTypes.PostCRVstatus = true

		klog.V(9).Infof("debug: CRV is %v", aggregatorClientReq.CRV)

		for k, v := range aggregatorClientReq.CRV {
			klog.V(9).Infof("debug: RV in map, key: %v-%v, value: %v", k.GetRegion(), k.GetResourcePartition(), v)
		}

		// Serialize boolean result to JSON
		err = postCRV.ToJSON(rw)

		if err != nil {
			klog.Errorf("Error - Unable to marshal json : ", err)
		}

		klog.Info("Handle POST CRV to discard all old region node modified events successfully")
	}
}
