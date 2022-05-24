package endpoints

import (
	"encoding/json"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"

	di "global-resource-service/resource-management/pkg/common-lib/interfaces/distributor"
	"global-resource-service/resource-management/pkg/common-lib/types"
	"global-resource-service/resource-management/pkg/common-lib/types/event"
	apiTypes "global-resource-service/resource-management/pkg/service-api/types"
)

type Installer struct {
	dist di.Interface
}

func NewInstaller(d di.Interface) *Installer {
	return &Installer{d}
}

func (i *Installer) ResourceHandler(resp http.ResponseWriter, req *http.Request) {
	klog.V(3).Infof("handle /resource. URL path: %s", req.URL.Path)

	switch req.Method {
	case http.MethodGet:
		ctx := req.Context()
		clientId := ctx.Value("clientid").(string)

		if req.URL.Query().Get(WatchParameter) == WatchParameterTrue {
			i.serverWatch(resp, req, clientId)
			return
		}

		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "text/plain")

		nodes, _, err := i.dist.ListNodesForClient(clientId)

		ret, err := json.Marshal(nodes)
		klog.V(3).Infof("node ret: %s", ret)
		if err != nil {
			klog.V(3).Infof("error read get node list. error %v", err)
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp.Write(ret)
	case http.MethodPut:
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	default:
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}

// simple watch routine
// TODO: add timeout support
// TODO: with serialization options
// TODO: error code and string definition
//
func (i *Installer) serverWatch(resp http.ResponseWriter, req *http.Request, clientId string) {
	klog.V(3).Infof("Serving watch for client: %s", clientId)

	// For 630 distributor impl, watchChannel and stopChannel passed into the Watch routine from API layer
	watchCh := make(chan *event.NodeEvent, WatchChannelSize)
	stopCh := make(chan struct{})

	// Signal the distributor to stop the watch for this client on exit
	defer stopWatch(stopCh)

	// read request body and get the crv
	crvMap, err := getResourceVersionsMap(req)
	if err != nil {
		klog.Errorf("uUable to get the resource versions. Error %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// start the watcher
	err = i.dist.Watch(clientId, crvMap, watchCh, stopCh)
	if err != nil {
		klog.Errorf("uUable to start the watch at store. Error %v", err)
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

	for {
		select {
		case <-done:
			return
		case record, ok := <-watchCh:
			if !ok {
				// End of results.
				return
			}

			if err := json.NewEncoder(resp).Encode(*record.Node); err != nil {
				klog.V(3).Infof("encoding record failed. error %v", err)
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			if len(watchCh) == 0 {
				flusher.Flush()
			}
		}
	}
}

// Helper functions
func stopWatch(stopCh chan struct{}) {
	stopCh <- struct{}{}
}

func getResourceVersionsMap(req *http.Request) (types.ResourceVersionMap, error) {
	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}

	wr := apiTypes.WatchRequest{}

	err = json.Unmarshal(body, wr)
	if err != nil {
		return nil, err
	}

	return wr.ResourceVersions, nil
}
