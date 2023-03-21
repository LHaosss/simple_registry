package simple_registry

type Registration struct {
	ServiceName string
	ServicePort string

	ServiceUrl string

	HeartbeatDetectionUrl string
	UpdateUrl             string

	DependedServicesName []string
}

func InitRegistration(serviceName, servicePort, serviceUrl string, dependences []string) Registration {
	return Registration{
		ServiceName: serviceName,
		ServicePort: servicePort,

		ServiceUrl:            serviceUrl,
		HeartbeatDetectionUrl: "http://localhost" + servicePort + "/heartbeat",
		UpdateUrl:             "http://localhost" + servicePort + "/update",

		DependedServicesName: dependences,
	}
}
