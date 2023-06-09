package simple_registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

const DefaultServiceURL = "http://localhost:3000/services"

type Dependence struct {
	mut      sync.Mutex
	Services map[string][]Registration
}

var dependence *Dependence

type RegistryClient struct {
	RegistryUrl string

	HeartbeatService heartbeatService
	UpdateService    updateService

	Service http.Handler

	Ctx context.Context
}

// 初始化RegistryClient
func InitRegistryClient(ctx context.Context, registryUrl string, service http.Handler) *RegistryClient {
	client := &RegistryClient{
		RegistryUrl:      registryUrl,
		HeartbeatService: heartbeatService{},
		UpdateService:    updateService{},

		Service: service,

		Ctx: ctx,
	}
	if client.RegistryUrl == "" {
		client.RegistryUrl = "http://localhost:3000/services"
	}

	return client
}

// 启动服务
func (client *RegistryClient) Start(servicePath, addr string) http.Server {
	dependence = &Dependence{
		Services: make(map[string][]Registration),
	}
	// 注册依赖更新服务
	http.Handle("/update", client.UpdateService)
	// 注册心跳服务
	http.Handle("/heartbeat", client.HeartbeatService)

	// 注册服务
	http.Handle(servicePath, client.Service)

	var srv http.Server
	srv.Addr = addr

	return srv
}

// 注册服务
func (client *RegistryClient) Register(reg Registration) error {
	// 向注册中心发送服务注册请求
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(reg); err != nil {
		fmt.Printf("编码错误，err: %v\n", err)
		return err
	}

	// 发送请求
	resp, err := http.Post(client.RegistryUrl, "application/json", buf)
	if err != nil {
		fmt.Println("服务注册请求发送失败")
		return err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("服务注册失败")
		return errors.New("服务注册失败")
	}

	return nil
}

// 注销服务
func (client *RegistryClient) Deregister(reg Registration) error {

	return nil
}

// 默认心跳检测服务
type heartbeatService struct{}

func (h heartbeatService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Println("接受到心跳检测请求")
		w.WriteHeader(http.StatusOK)
		return
	default:
		fmt.Println("ssss")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// 定义默认依赖更新服务
type updateService struct{}

func (u updateService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	var update UpdateInfo
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&update)
	if err != nil {
		fmt.Println("编码失败")
		w.WriteHeader(http.StatusBadRequest)
	}

	dependence.mut.Lock()
	defer dependence.mut.Unlock()
	if len(update.Add) > 0 {
		fmt.Println("Update!!!!")
		for _, add := range update.Add {
			if services, ok := dependence.Services[add.ServiceName]; ok {
				success := true
				for _, service := range services {
					if add.ServiceUrl == service.ServiceUrl {
						success = false
						break
					}
				}
				if success {
					dependence.Services[add.ServiceName] = append(dependence.Services[add.ServiceName], add)
				}
			} else {
				dependence.Services[add.ServiceName] = []Registration{add}
			}
		}
	}

	if len(update.Remove) > 0 {
		fmt.Println("Remove!!!!")
		for _, remove := range update.Remove {
			if services, ok := dependence.Services[remove.ServiceName]; ok {
				for index, service := range services {
					if service.ServiceUrl == remove.ServiceUrl {
						dependence.Services[remove.ServiceName] = append(services[:index], services[index+1:]...)
						break
					}
				}
			}
		}
	}

	fmt.Println(dependence.Services)

	w.WriteHeader(http.StatusOK)
}
