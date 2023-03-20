package main

import (
	"context"
	"fmt"
	"net/http"
	registry "simple_registry"
)

func main() {
	// 注册服务
	// 初始化服务注册客户端
	regClient := registry.InitRegistryClient(context.Background(), "", "http://localhost:6000/heartbeat")

	reg := registry.InitRegistration("Function Servcie", ":6000", "http://localhost:6000/function", "http://localhost:6000/heartbeat", "http://localhost:6000/update")
	err := regClient.RegisterService(reg, &log{})
	if err != nil {
		fmt.Println("服务注册失败")
	}

	// 关闭服务逻辑
	fmt.Println("log服务正在关闭...")
}

type log struct{}

func (l log) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("接收到function请求")
	w.WriteHeader(http.StatusOK)
}
