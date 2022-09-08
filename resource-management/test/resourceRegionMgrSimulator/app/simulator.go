/*
Copyright 2022 Authors of Global Resource Service.

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

package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/klog/v2"

	"global-resource-service/resource-management/test/resourceRegionMgrSimulator/handlers"
)

type RegionConfig struct {
	RegionName            string
	RpNum                 int
	NodesPerRP            int
	MasterPort            string
	DataPattern           string
	WaitTimeForMakeRpDown int
}

func Run(c *RegionConfig) error {

	// Create the handlers
	rh := handlers.NewRegionNodeEventsHander()

	wh := handlers.NewWatchHander()

	// create a new serve mux and register the handlers
	sm := mux.NewRouter().StrictSlash(true)

	// handlers for GET API
	//
	getRouter := sm.Methods(http.MethodGet).Subrouter()

	// For initial pull all mini node added events in all RPs of one specified region
	getRouter.HandleFunc(handlers.InitPullPath, rh.SimulatorHandler)

	// For subsequent pull all mini node modified events in all RPs of one specified region
	getRouter.HandleFunc(handlers.SubsequentPullPath, rh.SimulatorHandler)

	// List resources
	getRouter.HandleFunc(handlers.RegionlessResourcePath, wh.ResourceHandler)

	// handlers for POST API
	//
	postRouter := sm.Methods(http.MethodPost).Subrouter()

	// with CRV, discard all mini node modified events
	// which resource version is older than CRV in all RPs of specified region
	postRouter.HandleFunc(handlers.PostCRVPath, rh.SimulatorHandler)

	// watch for node changes
	postRouter.HandleFunc(handlers.RegionlessResourcePath, wh.ResourceHandler)

	var bindAddress = ":" + c.MasterPort

	// Define HTTP Server
	s := &http.Server{
		Addr:         bindAddress,
		Handler:      sm,
		ReadTimeout:  120 * time.Second, // hack: large time out for listing large amount of nodes without pagenition
		WriteTimeout: 120 * time.Second,
	}

	go func() {
		klog.V(3).Infof("\nStarting resource region manager simulator server on port (%v)", c.MasterPort)

		err := s.ListenAndServe()
		if err != nil {
			klog.Errorf("The HTTP server is not gracefully shutdown : ", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	sig := <-sigChan
	klog.V(3).Infof("\n\nReceived terminate, graceful shutdown\n\n", sig)

	tc, err := context.WithTimeout(context.Background(), 30*time.Second)
	if err != nil {
		klog.Errorf("The HTTP server is not gracefully shutdown : ", err)
	}
	s.Shutdown(tc)
	return nil
}
