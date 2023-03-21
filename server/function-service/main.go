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
	regClient := registry.InitRegistryClient(context.Background(), "", function{})

	reg := registry.InitRegistration("Function Servcie", ":6000", "http://localhost:6000/function", []string{"Log Service"})

	err := regClient.Register(reg)
	if err != nil {
		fmt.Println("服务注册失败")
	}

	srv := regClient.Start("/function", ":6000")

	fmt.Println(srv.ListenAndServe())

	// 关闭服务逻辑
	fmt.Println("function服务正在关闭...")
}

type function struct{}

func (l function) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("接收到function请求")
	w.WriteHeader(http.StatusOK)
}
