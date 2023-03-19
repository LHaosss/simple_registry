package simple_registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const DefaultServiceURL = "http://localhost:3000/services"

// 服务注册客户端

type RegistryClient struct {
	ServiceURL            string
	HeartbeatDetectionApi string
	HeartbeatDetection    heartbeatDetection

	Ctx context.Context
}

func InitRegistryClient(ctx context.Context, serviceURL string, heartbeatDetectionApi string) *RegistryClient {
	regClient := &RegistryClient{
		ServiceURL:            DefaultServiceURL,
		HeartbeatDetectionApi: heartbeatDetectionApi,
		HeartbeatDetection:    heartbeatDetection{},
		Ctx:                   ctx,
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
func (client *Registration) DeregisterService(reg Registration) error {

	return nil
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
