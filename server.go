package simple_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	DefaultPort        = ":3000"
	DefaultServiceName = "/services"
)

// 启动服务
func StartService(opts *Options) {
	http.Handle(opts.ServiceName, &Registry{})

	var srv http.Server
	srv.Addr = opts.Port

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// 启动端口监听
		fmt.Println(srv.ListenAndServe())
		// 当监听端口down掉后，退出程序
		cancel()
	}()

	go func() {
		// 手动获取输入，退出程序，可以进一步晚上，通过监听请求，退出程序
		fmt.Println("Registry service started. Press any key to stop.")
		var s string
		fmt.Scanln(&s)
		srv.Shutdown(ctx)
		cancel()
	}()

	<-ctx.Done()
	fmt.Println("Shutting down registry service")
}

// 服务注册服务端
type Registry struct{}

func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Println("received request")

	switch req.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(req.Body)
		var registration Registration
		err := decoder.Decode(&registration)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		fmt.Println(registration)
		w.WriteHeader(http.StatusOK)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
