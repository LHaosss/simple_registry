package simple_registry

type Registration struct {
	ServiceName string
	ServicePort string

	ServiceUrl string

	HeartbeatDetectionUrl string
	UpdateUrl             string
}

func InitRegistration(serviceName, servicePort, serviceUrl, heartbeatDetectionUrl, updateUrl string) Registration {
	return Registration{
		ServiceName: serviceName,
		ServicePort: servicePort,

		ServiceUrl:            serviceUrl,
		HeartbeatDetectionUrl: heartbeatDetectionUrl,
		UpdateUrl:             updateUrl,
	}
}
