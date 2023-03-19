package main

import (
	"context"
	"fmt"
	"net/http"
	registry "simple_registry"
)

func main() {
	// 启动服务
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("接收到log请求")
		w.WriteHeader(http.StatusOK)
	})
	srv := http.Server{}
	srv.Addr = ":4000"

	go func() {
		fmt.Println(srv.ListenAndServe())
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 注册服务
	// 初始化服务注册客户端
	regClient := registry.InitRegistryClient("")

	reg := registry.Registration{
		ServiceName: "Log Service",
		ServiceUrl:  "http://localhost:5000/log",
	}
	err := regClient.RegisterService(reg)
	if err != nil {
		fmt.Println("服务注册失败，关闭服务")
		cancel()
	}

	<-ctx.Done()
	// 关闭服务逻辑
	fmt.Println("log服务正在关闭...")
}
