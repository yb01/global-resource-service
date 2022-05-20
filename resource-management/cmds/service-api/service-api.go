package main

import (
	"os"

	"k8s.io/klog/v2"

	"global-resource-service/resource-management/cmds/service-api/app"
)

func main() {

        klog.InitFlags(nil)
	klog.Infof("Starting resource management service")
	if err := app.Run(); err != nil {
		klog.Errorf("error: %v\n", err)
		os.Exit(1)
	}
}
