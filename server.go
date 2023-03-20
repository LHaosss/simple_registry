package simple_registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	DefaultPort        = ":3000"
	DefaultServiceName = "/services"
)

type registryStr struct {
	services map[string][]*Registration
	mut      sync.Mutex
}

var registry registryStr

// 启动服务
func StartService(opts *Options) {
	// 初始化服务存储信息
	registry = registryStr{
		services: make(map[string][]*Registration),
	}

	// 启动心跳检测
	SetupRegistryService()

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

// 心跳检测
var once sync.Once

func SetupRegistryService() {
	once.Do(func() {
		go registry.heartbeat(3 * time.Second)
	})
}

func (r *registryStr) heartbeat(freq time.Duration) {
	for {
		var wg sync.WaitGroup

		for _, services := range r.services {
			for _, service := range services {
				wg.Add(1)
				go func(reg *Registration) {
					defer wg.Done()
					success := true
					for attampts := 0; attampts < 3; attampts++ {
						client := http.Client{
							Timeout: 2 * time.Second,
						}
						res, err := client.Get(reg.HeartbeatDetectionUrl)
						if err != nil {
							fmt.Println(err)
						} else if res.StatusCode == http.StatusOK {
							fmt.Println("心跳检测成功")
							if !success {
								r.services[reg.ServiceName] = append(r.services[reg.ServiceName], reg)
							}
							break
						}
						fmt.Println("心跳检测失败")
						if success {
							success = false
							r.remove(reg)
						}
						time.Sleep(1 * time.Second)
					}
				}(service)
			}
		}
		wg.Wait()
		time.Sleep(freq)
	}
}

func (r *registryStr) remove(reg *Registration) {
	r.mut.Lock()
	defer r.mut.Unlock()

	// 查找服务
	_, ok := r.services[reg.ServiceName]
	if !ok {
		return
	}

	for i, service := range r.services[reg.ServiceName] {
		if service.ServiceUrl == reg.ServiceUrl {
			r.services[reg.ServiceName] = append(r.services[reg.ServiceName][:i], r.services[reg.ServiceName][i+1:]...)
		}
	}
}

func sendUpdate(addNames, removeNames []string, updateUrl string) {
	fmt.Println("开始执行依赖处理")
	update := Update{
		Add:    make([]*Patch, 0),
		Remove: make([]*Patch, 0),
	}

	for _, name := range addNames {
		if services, ok := registry.services[name]; ok {
			patchs := make([]*Patch, 0)
			for _, service := range services {
				patchs = append(patchs, &Patch{
					ServiceName: service.ServiceName,
					ServiceUrl:  service.ServiceUrl,
				})

			}
			update.Add = append(update.Add, patchs...)
		}
	}

	for _, name := range removeNames {
		if services, ok := registry.services[name]; ok {
			patchs := make([]*Patch, 0)
			for _, service := range services {
				patchs = append(patchs, &Patch{
					ServiceName: service.ServiceName,
					ServiceUrl:  service.ServiceUrl,
				})

			}
			update.Remove = append(update.Add, patchs...)
		}
	}

	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(update)
	if err != nil {
		fmt.Println("更新服务编码失败")
		return
	}

	resp, err := http.Post(updateUrl, "application/json", buf)
	if err != nil {
		fmt.Printf("请求出错 %v\n", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("服务更新请求出错")
		return
	}
	fmt.Println("服务更新请求发送成功")

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
		// 存储服务信息
		registry.mut.Lock()
		if _, ok := registry.services[registration.ServiceName]; !ok {
			registry.services[registration.ServiceName] = make([]*Registration, 0)
		}
		for i, reg := range registry.services[registration.ServiceName] {
			if reg.ServiceUrl == registration.ServiceUrl {
				registry.services[registration.ServiceName] = append(registry.services[registration.ServiceName][:i], registry.services[registration.ServiceName][i+1:]...)
			}
		}
		registry.services[registration.ServiceName] = append(registry.services[registration.ServiceName], &registration)
		registry.mut.Unlock()

		// 发送依赖信息
		add := make([]string, 0)
		for _, patch := range registration.DependedServices {
			add = append(add, patch.ServiceName)
		}

		sendUpdate(add, []string{}, registration.UpdateUrl)

		w.WriteHeader(http.StatusOK)
		return
	case http.MethodGet:
		registry.mut.Lock()
		defer registry.mut.Unlock()

		// 返回registry.services中所有数据
		buf := new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		err := enc.Encode(registry.services)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write(buf.Bytes())
		w.WriteHeader(http.StatusOK)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
