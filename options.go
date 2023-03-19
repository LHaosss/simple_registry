package simple_registry

type Option func(*Options)

type Options struct {
	Port        string
	ServiceName string
}

func InitOptions(opts ...Option) *Options {
	o := &Options{
		Port:        DefaultPort,
		ServiceName: DefaultServiceName,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func Port(port string) Option {
	return func(o *Options) {
		o.Port = port
	}
}
