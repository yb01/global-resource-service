package main

import (
	"flag"
	"fmt"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"time"

	"global-resource-service/resource-management/pkg/clientSdk/rmsclient"
)

func main() {
	flag.Usage = printUsage

	cfg := rmsclient.Config{}
	var regions string

	flag.StringVar(&cfg.ServiceUrl, "service_url", "localhost:8080", "Service IP address, if not set, default to localhost")
	flag.DurationVar(&cfg.RequestTimeout, "request_timeout", 30*time.Minute, "Timeout for client requests and responses")
	flag.StringVar(&cfg.ClientFriendlyName, "friendly_name", "testclient", "Client friendly name other that the assigned Id")
	flag.StringVar(&cfg.ClientRegion, "client_region", "Beijing", "Client identify where it is located")
	flag.IntVar(&cfg.InitialRequestTotalMachines, "request_machines", 1000, "Initial request of number of machines")
	flag.StringVar(&regions, "request_regions", "Beijing", "list of regions, in comma separated string, to allocate the machines for the client")

	if !flag.Parsed() {
		klog.InitFlags(nil)
		flag.Parse()
	}
	klog.StartFlushDaemon(time.Second * 1)
	defer klog.Flush()

	cfg.InitialRequestRegions = strings.Split(regions, ",")

	client := rmsclient.NewRmsClient(cfg)

	klog.Infof("Register client to service  ...")
	registrationResp, err := client.Register()
	if err != nil {
		klog.Errorf("failed register client to service. error %v", err)
	}

	klog.Infof("Got client registration from service: %v", registrationResp)
	client.Id = registrationResp.ClientId

	klog.Infof("List resources from service ...")
	nodeList, crv, err := client.List(client.Id)
	if err != nil {
		klog.Errorf("failed list resource from service. error %v", err)
	}
	klog.Infof("Got [%v] nodes from service", len(nodeList))
	klog.Infof("Got [%v] resource versions service", crv)

	klog.Infof("Watch resources update from service ...")
	watcher, err := client.Watch(client.Id)
	if err != nil {
		klog.Errorf("failed list resource from service. error %v", err)
	}

	watchCh := watcher.ResultChan()
	// retrieve updates from watcher
	for {
		select {
		case record, ok := <-watchCh:
			if !ok {
				// End of results.
				klog.V(3).Infof("End of results")
				return
			}

			// TODO: write cache, for now just logout
			klog.V(3).Infof("Getting event from servicee: %v", record)
		}
	}
}

// function to print the usage info for the resource management api server
func printUsage() {
	fmt.Println("Usage: ")
	fmt.Println("--service_url=127.0.0.1:8080 --request_machines=10000 --request_regions=Beijing,Shanghai ")

	os.Exit(0)
}
