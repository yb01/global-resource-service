package app

import (
	"net/http"
	"time"

	"global-resource-service/resource-management/pkg/aggregrator"
	"global-resource-service/resource-management/pkg/distributor"
	"global-resource-service/resource-management/pkg/service-api/endpoints"

	"github.com/gorilla/mux"
	"k8s.io/klog/v2"
)

type Config struct {
	ResourceUrls []string
	MasterIp     string
	MasterPort   string
}

// Run and create new service-api.  This should never exit.
func Run(c *Config) error {
	klog.V(3).Infof("Starting the API server...")

	dist := distributor.GetResourceDistributor()
	installer := endpoints.NewInstaller(dist)

	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc(endpoints.ListResourcePath, installer.ResourceHandler).Methods("GET")
	r.HandleFunc(endpoints.WatchResourcePath, installer.ResourceHandler)
	r.HandleFunc(endpoints.UpdateResourcePath, installer.ResourceHandler)
	r.HandleFunc(endpoints.ReduceResourcePath, installer.ResourceHandler)

	var addr string
	var p string

	if c.MasterIp == "" {
		addr = "localhost"
	}

	if c.MasterPort == "" {
		p = endpoints.InsecureServiceAPIPort
	}

	server := &http.Server{
		Handler:      r,
		Addr:         addr + p,
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
	}

	err := server.ListenAndServe()

	if err != nil {
		return err
	}

	// start the aggregator instance
	klog.V(3).Infof("Starting the Aggregator ...")

	aggregator := aggregrator.NewAggregator(c.ResourceUrls, dist)
	err = aggregator.Run()
	if err != nil {
		return err
	}

	return nil
}
