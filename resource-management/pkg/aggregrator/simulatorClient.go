/*
Copyright The Kubernetes Authors.
Copyright 2022 Authors of Global Resource Service - file modified.
Copyright 2020 Authors of Arktos - file modified.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aggregrator

import (
	"encoding/json"
	"global-resource-service/resource-management/pkg/common-lib/types/event"
	simulatorTypes "global-resource-service/resource-management/test/resourceRegionMgrSimulator/types"
	"net/http"
	"strconv"
	"time"

	"k8s.io/klog/v2"

	"global-resource-service/resource-management/pkg/clientSdk/rest"
	"global-resource-service/resource-management/pkg/clientSdk/watch"
	"global-resource-service/resource-management/pkg/common-lib/types"
	apiTypes "global-resource-service/resource-management/pkg/service-api/types"
)

// client config that the client can setup for
type Config struct {
	ServiceUrl     string
	RequestTimeout time.Duration
}

// ListOptions contains optional settings for List nodes
type ListOptions struct {
	// Limit is equailent to URL query parameter ?limit=500
	Limit int
}

// RmsInterface has methods to work with Resource management service resources.
// below are just 630 related interface definitions
type SimInterface interface {
	List(ListOptions) ([][]*event.NodeEvent, types.TransitResourceVersionMap, uint64, error)
	Watch(types.TransitResourceVersionMap) (watch.Interface, error)
}

// rmsClient implements RmsInterface
type simClient struct {
	config Config
	// REST client to RMS service
	restClient rest.Interface
}

// NewSimClient returns a refence to the rsmClient object
func NewSimClient(cfg Config) *simClient {
	httpclient := http.Client{Timeout: cfg.RequestTimeout}
	url, err := rest.DefaultServerURL(cfg.ServiceUrl, "", false)

	if err != nil {
		klog.Errorf("failed to get the default URL. error %v", err)
		return nil
	}

	c, err := rest.NewRESTClient(url, rest.ClientContentConfig{}, nil, &httpclient)
	if err != nil {
		klog.Errorf("failed to get the RESTClient. error %v", err)
		return nil
	}

	return &simClient{
		config:     cfg,
		restClient: c,
	}
}

// List takes label and field selectors, and returns the list of Nodes that match those selectors.
func (c *simClient) List(opts ListOptions) ([][]*event.NodeEvent, types.TransitResourceVersionMap, uint64, error) {
	req := c.restClient.Get()
	req = req.Resource("resource")
	req = req.Timeout(c.config.RequestTimeout)
	req = req.Param("limit", strconv.Itoa(opts.Limit))

	respRet, err := req.DoRaw()
	if err != nil {
		return nil, nil, 0, err
	}

	resp := simulatorTypes.ResponseFromRRM{}

	err = json.Unmarshal(respRet, &resp)

	actualCrv := resp.RvMap

	return resp.RegionNodeEvents, actualCrv, resp.Length, nil

}

// Watch returns a watch.Interface that watches the requested rmsClient.
func (c *simClient) Watch(versionMap types.TransitResourceVersionMap) (watch.Interface, error) {
	req := c.restClient.Post()
	req = req.Resource("resource")
	req = req.Timeout(c.config.RequestTimeout)
	req = req.Param("watch", "true")

	crv := apiTypes.WatchRequest{ResourceVersions: versionMap}

	body, err := json.Marshal(crv)
	if err != nil {
		return nil, err
	}
	req = req.Body(body)

	watcher, err := req.Watch()
	if err != nil {
		return nil, err
	}

	return watcher, nil
}
