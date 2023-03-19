package simple_registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const DefaultServiceURL = "http://localhost:3000/services"

// 服务注册客户端

type RegistryClient struct {
	ServiceURL string
}

func InitRegistryClient(serviceURL string) *RegistryClient {
	regClient := &RegistryClient{
		ServiceURL: DefaultServiceURL,
	}
	if serviceURL != "" {
		regClient.ServiceURL = serviceURL
	}
	return regClient
}

// 注册服务
func (client *RegistryClient) RegisterService(reg Registration) error {
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
	return nil
}
