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
	regClient := registry.InitRegistryClient(context.Background(), "", log{})

	reg := registry.InitRegistration("Log Service", ":4000", "http://localhost:4000/log", []string{""})

	err := regClient.Register(reg)
	if err != nil {
		fmt.Println("服务注册失败")
	}

	srv := regClient.Start("/log", ":4000")

	fmt.Println(srv.ListenAndServe())

	// 关闭服务逻辑
	fmt.Println("log服务正在关闭...")
}

type log struct{}

func (l log) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("接收到log请求")
	w.WriteHeader(http.StatusOK)
}
