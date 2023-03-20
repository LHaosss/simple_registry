package simple_registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

const DefaultServiceURL = "http://localhost:3000/services"

type providerStr struct {
	mut      sync.Mutex
	services map[string][]*Patch
}

var provider providerStr = providerStr{
	services: make(map[string][]*Patch, 0),
}

// 服务注册客户端

type RegistryClient struct {
	ServiceURL            string
	HeartbeatDetectionApi string
	UpdateApi             string
	HeartbeatDetection    heartbeatDetection
	Update                update

	Ctx context.Context
}

func InitRegistryClient(ctx context.Context, serviceURL string, heartbeatDetectionApi, updateApi string) *RegistryClient {
	regClient := &RegistryClient{
		ServiceURL:            DefaultServiceURL,
		HeartbeatDetectionApi: heartbeatDetectionApi,
		HeartbeatDetection:    heartbeatDetection{},

		UpdateApi: updateApi,
		Update:    update{},

		Ctx: ctx,
	}
	if serviceURL == "" {
		regClient.ServiceURL = DefaultServiceURL
	}
	return regClient
}

// 注册服务
func (client *RegistryClient) RegisterService(reg Registration, handler http.Handler) error {
	ctx, cancel := context.WithCancel(client.Ctx)
	defer cancel()

	// 注册服务
	http.Handle(strings.Split(reg.ServiceUrl, "/")[len(strings.Split(reg.ServiceUrl, "/"))-1], handler)

	// 提供心跳检测向服务端服务
	http.Handle("/heartbeat", client.HeartbeatDetection)

	// 提供依赖更新接口
	http.Handle("/update", client.Update)

	// 启动服务
	var srv http.Server
	srv.Addr = reg.ServicePort

	go func() {
		fmt.Println(srv.ListenAndServe())
		cancel()
	}()

	// 向服务端发送注册服务注册请求
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(reg)
	if err != nil {
		return err
	}
	// 向服务注册服务端发送注册请求
	resp, err := http.Post(client.ServiceURL, "application/json", buf)
	if err != nil {
		fmt.Printf("服务注册请求发送失败, %v\n", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("服务请求出错")
		return err
	}

	fmt.Println("服务注册成功")
	<-ctx.Done()

	return nil
}

// 注销服务
func (client *RegistryClient) DeregisterService(reg Registration) error {

	return nil
}

// 获取依赖信息
func (client *RegistryClient) GetDependedServicesByName(serviceName string) ([]*Patch, error) {
	services := make([]*Patch, 0)
	if srvs, ok := provider.services[serviceName]; ok {
		services = append(services, srvs...)
		return services, nil
	} else {
		return nil, errors.New("未找到可用依赖服务")
	}

}

// 心跳检测
type heartbeatDetection struct{}

func (h heartbeatDetection) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Println("接受到心跳检测请求")
		w.WriteHeader(http.StatusOK)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

type update struct{}

func (u update) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("接收到依赖更新请求")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	provider.mut.Lock()
	defer provider.mut.Unlock()
	// 处理依赖更新
	decoder := json.NewDecoder(r.Body)
	var update Update
	err := decoder.Decode(&update)
	if err != nil {
		fmt.Println("更新依赖失败")
		return
	}
	// 增添依赖
	for _, srv := range update.Add {
		services, ok := provider.services[srv.ServiceName]
		if !ok {
			provider.services[srv.ServiceName] = make([]*Patch, 0)
		}
		for index, service := range services {
			if service.ServiceUrl == srv.ServiceUrl {
				provider.services[srv.ServiceName] = append(provider.services[srv.ServiceName][:index], provider.services[srv.ServiceName][index+1:]...)
			}
		}
		provider.services[srv.ServiceName] = append(provider.services[srv.ServiceName], srv)
	}

	// 删除依赖
	for _, srv := range update.Remove {
		if services, ok := provider.services[srv.ServiceName]; !ok {
			continue
		} else {
			for index, service := range services {
				if srv.ServiceUrl == service.ServiceName {
					provider.services[srv.ServiceName] = append(provider.services[srv.ServiceName][:index], provider.services[srv.ServiceName][index+1:]...)
					break
				}
			}
		}

	}
	for key, value := range provider.services {
		fmt.Println("lllll", key, value)
	}

	w.WriteHeader(http.StatusOK)
}
