package endpoints

import (
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"

	di "global-resource-service/resource-management/pkg/common-lib/interfaces/distributor"
	"global-resource-service/resource-management/pkg/common-lib/interfaces/store"
	"global-resource-service/resource-management/pkg/common-lib/serializer"
	"global-resource-service/resource-management/pkg/common-lib/serializer/json"
	"global-resource-service/resource-management/pkg/common-lib/types"
	"global-resource-service/resource-management/pkg/common-lib/types/event"
	apiTypes "global-resource-service/resource-management/pkg/service-api/types"
)

type Installer struct {
	dist di.Interface
	js   serializer.Serializer
	ps   serializer.Serializer
}

func NewInstaller(d di.Interface) *Installer {
	s := json.NewSerializer("placeHolder", false)
	return &Installer{d, s, s}
}

func (i *Installer) ClientAdministrationHandler(resp http.ResponseWriter, req *http.Request) {
	klog.V(3).Infof("handle /client. URL path: %s", req.URL.Path)

	// TODO: consider to add multiple handler for different serializer request, could avoid a bit perf impact for this set
	var desiredSerializer serializer.Serializer
	content := req.Header.Get("Content-Type")
	if content == "application/json" {
		desiredSerializer = i.js
	} else {
		desiredSerializer = i.ps
	}

	switch req.Method {
	case http.MethodPost:
		i.handleClientRegistration(resp, req, desiredSerializer)
		return
	case http.MethodDelete:
		i.handleClientUnRegistration(resp, req, desiredSerializer)
		return
	default:
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}

// TODO: error handling function
func (i *Installer) handleClientRegistration(resp http.ResponseWriter, req *http.Request, desiredSerializer serializer.Serializer) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		klog.V(3).Infof("error read request. error %v", err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	clientReq := apiTypes.ClientRegistrationRequest{}

	r, err := desiredSerializer.Decode(body, clientReq)
	if err != nil {
		klog.V(3).Infof("error unmarshal request body. error %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	clientReq = r.(apiTypes.ClientRegistrationRequest)

	requestedMachines := clientReq.InitialRequestedResource.TotalMachines
	if requestedMachines > types.MaxTotalMachinesPerRequest || requestedMachines < types.MinTotalMachinesPerRequest {
		klog.V(3).Infof("Invalid request of resources. requested total machines: %v", requestedMachines)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: need to design to avoid client to register itself
	c_id := fmt.Sprintf("%s-%s", store.Preserve_Client_KeyPrefix, uuid.New().String())
	client := types.Client{ClientId: c_id, Resource: clientReq.InitialRequestedResource, ClientInfo: clientReq.ClientInfo}

	err = i.dist.RegisterClient(&client)

	if err != nil {
		klog.V(3).Infof("error register client. error %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// for 630, request of initial resource request with client registration is either denied or granted in full
	ret := apiTypes.ClientRegistrationResponse{ClientId: client.ClientId, GrantedResource: client.Resource}

	err = desiredSerializer.Encode(ret, resp)
	if err != nil {
		klog.V(3).Infof("error write response. error %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}

func (i *Installer) handleClientUnRegistration(resp http.ResponseWriter, req *http.Request, desiredSerializer serializer.Serializer) {
	klog.V(3).Infof("not implemented")
	resp.WriteHeader(http.StatusNotImplemented)
	return
}

func (i *Installer) ResourceHandler(resp http.ResponseWriter, req *http.Request) {
	klog.V(3).Infof("handle /resource. URL path: %s", req.URL.Path)

	// TODO: consider to add multiple handler for different serializer request, could avoid a bit perf impact for this set
	var desiredSerializer serializer.Serializer
	content := req.Header.Get("Content-Type")
	if content == "application/json" {
		desiredSerializer = i.js
	} else {
		desiredSerializer = i.ps
	}

	switch req.Method {
	case http.MethodGet:
		ctx := req.Context()
		clientId := ctx.Value("clientid").(string)

		if req.URL.Query().Get(WatchParameter) == WatchParameterTrue {
			i.serverWatch(resp, req, clientId, desiredSerializer)
			return
		}
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "text/plain")

		nodes, _, err := i.dist.ListNodesForClient(clientId)
		if err != nil {
			klog.V(3).Infof("error to get node list from distributor. error %v", err)
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
		i.handleResponseTrunked(resp, nodes, desiredSerializer)
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
func (i *Installer) serverWatch(resp http.ResponseWriter, req *http.Request, clientId string, desiredSerializer serializer.Serializer) {
	klog.V(3).Infof("Serving watch for client: %s", clientId)

	// For 630 distributor impl, watchChannel and stopChannel passed into the Watch routine from API layer
	watchCh := make(chan *event.NodeEvent, WatchChannelSize)
	stopCh := make(chan struct{})

	// Signal the distributor to stop the watch for this client on exit
	defer stopWatch(stopCh)

	// read request body and get the crv
	crvMap, err := getResourceVersionsMap(req, desiredSerializer)
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

			if err := desiredSerializer.Encode(*record.Node, resp); err != nil {
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

func getResourceVersionsMap(req *http.Request, s serializer.Serializer) (types.ResourceVersionMap, error) {
	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}

	wr := apiTypes.WatchRequest{}

	r, err := s.Decode(body, wr)
	if err != nil {
		return nil, err
	}
	wr = r.(apiTypes.WatchRequest)

	return wr.ResourceVersions, nil
}

func (i *Installer) handleResponseTrunked(resp http.ResponseWriter, nodes []*types.LogicalNode, desiredSerializer serializer.Serializer) {
	var nodesLen = len(nodes)
	if nodesLen <= ResponseTrunkSize {
		err := desiredSerializer.Encode(nodes, resp)
		if err != nil {
			klog.Errorf("error read get node list. error %v", err)
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		flusher, ok := resp.(http.Flusher)
		if !ok {
			klog.Errorf("expected http.ResponseWriter to be an http.Flusher")
		}
		resp.Header().Set("Connection", "Keep-Alive")
		resp.Header().Set("X-Content-Type-Options", "nosniff")
		//TODO: handle network disconnect or similar cases.
		var chunkedNodes []*types.LogicalNode
		start := 0
		for start < nodesLen {
			end := start + ResponseTrunkSize
			if end < nodesLen {
				chunkedNodes = nodes[start:end]
			} else {
				chunkedNodes = nodes[start:nodesLen]
			}
			err := desiredSerializer.Encode(chunkedNodes, resp)
			if err != nil {
				klog.Errorf("error read get node list. error %v", err)
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			flusher.Flush()
			start = end
		}
	}
	return
}
