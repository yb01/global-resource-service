package endpoints

// URL path
const (
	//RegionlessResourcePath is the default api service url
	RegionlessResourcePath = "/resource"
	ListResourcePath       = RegionlessResourcePath + "/{clientid}"
	WatchResourcePath      = RegionlessResourcePath + "/{clientid}?watch?{resourceVersion}"
	UpdateResourcePath     = RegionlessResourcePath + "/{clientid}" + "/addResource"
	ReduceResourcePath     = RegionlessResourcePath + "/{clientid}" + "/reduceResource"
	// InsecureServiceAPIPort is the default port for Service-api when running insecure mode.
	// TODO: Can be overridden by a flag at startup.
	InsecureServiceAPIPort = "8080"
	// SecureServiceAPIPort is the default port for Service-api when running secure mode.
	// TODO: Can be overridden by a flag at startup.
	SecureServiceAPIPort = 443

	WatchChannelSize   = 100
	WatchParameter     = "watch"
	WatchParameterTrue = "true"
)
