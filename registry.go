package simple_registry

type Registration struct {
	ServiceName string
	ServicePort string

	ServiceUrl string

	HeartbeatDetectionUrl string
	UpdateUrl             string

	DependedServices []*Patch
}

func InitRegistration(serviceName, servicePort, serviceUrl, heartbeatDetectionUrl, updateUrl string, dependedServices []*Patch) Registration {
	return Registration{
		ServiceName: serviceName,
		ServicePort: servicePort,

		ServiceUrl:            serviceUrl,
		HeartbeatDetectionUrl: heartbeatDetectionUrl,
		UpdateUrl:             updateUrl,
		DependedServices:      dependedServices,
	}
}

type Patch struct {
	ServiceName string
	ServiceUrl  string
}

func InitDependedService(serviceNames []string) []*Patch {
	services := make([]*Patch, 0)
	for _, name := range serviceNames {
		services = append(services, &Patch{
			ServiceName: name,
		})
	}

	return services
}

type Update struct {
	Add    []*Patch
	Remove []*Patch
}
