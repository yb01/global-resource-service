package handlers

import (
	"encoding/json"
	"global-resource-service/resource-management/pkg/common-lib/metrics"
	"global-resource-service/resource-management/pkg/common-lib/types"
	"global-resource-service/resource-management/pkg/common-lib/types/event"
	ep "global-resource-service/resource-management/pkg/service-api/endpoints"
	apiTypes "global-resource-service/resource-management/pkg/service-api/types"
	"global-resource-service/resource-management/test/resourceRegionMgrSimulator/data"
	simulatorTypes "global-resource-service/resource-management/test/resourceRegionMgrSimulator/types"

	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
)

type WatchHandler struct{}

func NewWatchHander() *WatchHandler {
	return &WatchHandler{}
}

func (i *WatchHandler) ResourceHandler(resp http.ResponseWriter, req *http.Request) {
	klog.V(3).Infof("handle /resource. URL path: %s", req.URL.Path)

	switch req.Method {
	case http.MethodGet:
		i.list(resp, req)
		return

	// hack: currently crv is used for watch watermark, this is up to 200 RPs which cannot fit as parameters or headers
	// unfortunately GET will return 405 with request body.
	// It's unlikely we can change to other solutions for now. so use POST to test verify the watch logic and flows for now.
	// TODO: switch to logical record or other means to set the water mark as query parameter
	case http.MethodPost:
		if req.URL.Query().Get(ep.WatchParameter) == ep.WatchParameterTrue {
			i.serverWatch(resp, req, "foo")
			return
		}
		return
	case http.MethodPut:
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	default:
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}

func (i *WatchHandler) list(resp http.ResponseWriter, req *http.Request) {
	var nodeEvents simulatorTypes.RegionNodeEvents
	var count uint64

	nodeEvents, count = data.GetRegionNodeAddedEvents(0)

	if count == 0 {
		klog.V(6).Info("Pulling Region Node Events with batch is in the end")
	} else {
		klog.V(6).Infof("Pulling Region Node Event with final batch size (%v) for (%v) RPs", count, len(nodeEvents))
	}

	response := &simulatorTypes.ResponseFromRRM{
		RegionNodeEvents: nodeEvents,
		RvMap:            nil,
		Length:           uint64(count),
	}

	// Serialize region node events result to JSON
	err := response.ToJSON(resp)

	if err != nil {
		klog.Errorf("Error - Unable to marshal json : ", err)
	}

	// Process post CRV to discard all old region node modified event
	//
}

// simple watch routine
// TODO: add timeout support
// TODO: with serialization options
// TODO: error code and string definition
//
func (i *WatchHandler) serverWatch(resp http.ResponseWriter, req *http.Request, clientId string) {
	klog.V(3).Infof("Serving watch for client: %s", clientId)

	// For 630 distributor impl, watchChannel and stopChannel passed into the Watch routine from API layer
	watchCh := make(chan *event.NodeEvent, ep.WatchChannelSize)
	stopCh := make(chan struct{})

	// Signal the distributor to stop the watch for this client on exit
	defer stopWatch(stopCh)

	// TODO: change this to
	// read request body and get the crv
	crvMap, err := getResourceVersionsMap(req)
	if err != nil {
		klog.Errorf("unable to get the resource versions. Error %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	klog.V(9).Infof("Received CRV: %v", crvMap)

	// start the watcher
	klog.V(3).Infof("Start watching distributor for client: %v", clientId)

	// TODO: change this to simulator nod change watch interface
	//	err = i.dist.Watch(clientId, crvMap, watchCh, stopCh)
	if err != nil {
		klog.Errorf("unable to start the watch at store. Error %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	done := req.Context().Done()
	flusher, ok := resp.(http.Flusher)
	if !ok {
		klog.Errorf("unable to start watch - can't get http.Flusher")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// begin the stream
	resp.Header().Set("Content-Type", req.Header.Get("Content-Type"))
	resp.Header().Set("Transfer-Encoding", "chunked")
	resp.WriteHeader(http.StatusOK)
	flusher.Flush()

	klog.V(3).Infof("Start processing watch event for client: %v", clientId)
	for {
		select {
		case <-done:
			return
		case record, ok := <-watchCh:
			if !ok {
				// End of results.
				klog.Infof("End of results")
				return
			}

			klog.V(6).Infof("Getting event from distributor, node Id: %v", record.Node.Id)

			if err := json.NewEncoder(resp).Encode(*record); err != nil {
				klog.V(3).Infof("encoding record failed. error %v", err)
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			record.SetCheckpoint(metrics.Serializer_Encoded)
			if len(watchCh) == 0 {
				flusher.Flush()
			}
			record.SetCheckpoint(metrics.Serializer_Sent)
			event.AddLatencyMetricsAllCheckpoints(record)
		}
	}
}

// Helper functions
func stopWatch(stopCh chan struct{}) {
	stopCh <- struct{}{}
}

func getResourceVersionsMap(req *http.Request) (types.TransitResourceVersionMap, error) {
	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		klog.Errorf("Failed read request body, error %v", err)
		return nil, err
	}

	wr := apiTypes.WatchRequest{}

	err = json.Unmarshal(body, &wr)
	if err != nil {
		klog.Errorf("Failed unmarshal request body, error %v", err)
		return nil, err
	}

	return wr.ResourceVersions, nil
}
