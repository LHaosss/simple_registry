package simple_registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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
								// 把依赖添加回来
								success = true
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
					if !success {
						// 通知所有依赖该服务的服务，移除该依赖
						r.sendRemove(*reg)
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
			return
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

		fmt.Println(registry.services)

		// 发送可用依赖信息
		go registry.sendAdd(registration)

		// 向所有需要依赖该注册服务的服务发送依赖更新

		go func() {
			update := UpdateInfo{
				Add: []Registration{registration},
			}

			p, err := json.Marshal(update)
			if err != nil {
				fmt.Println("依赖更新内容编码失败")
			}

			for _, services := range registry.services {
				for _, service := range services {
					for _, name := range service.DependedServicesName {
						if name == registration.ServiceName {
							go func() {
								resp, err := http.Post(service.UpdateUrl, "application/json", bytes.NewBuffer(p))
								if err != nil {
									fmt.Printf("移除依赖失败，service: %s\n", name)
								}
								if resp.StatusCode != http.StatusOK {
									fmt.Println("依赖更新处理失败")
								}
							}()
						}
					}
				}
			}
		}()

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

func (registry *registryStr) sendAdd(reg Registration) error {
	fmt.Println("开始更新依赖")
	fmt.Println(reg.UpdateUrl)
	// 发送可用依赖信息
	update := UpdateInfo{
		Add: make([]Registration, 0),
	}

	for serviceName, services := range registry.services {
		for _, name := range reg.DependedServicesName {
			if name == serviceName {
				for _, service := range services {
					update.Add = append(update.Add, *service)
					fmt.Println(name)
				}
			}
		}
	}

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(update)

	if err != nil {
		fmt.Printf("编码失败, err: %v", err)
		return err
	}

	resp, err := http.Post(reg.UpdateUrl, "application/json", buf)
	if err != nil {
		fmt.Printf("发送更新依赖失败, %v\n", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("依赖更新错误")
		return errors.New("依赖更新出错, statusCode:" + strconv.Itoa(resp.StatusCode))
	}

	return nil
}

func (registry *registryStr) sendRemove(reg Registration) {
	var wg sync.WaitGroup

	update := UpdateInfo{
		Remove: make([]Registration, 0),
	}
	update.Remove = append(update.Remove, reg)

	p, err := json.Marshal(update)
	if err != nil {
		fmt.Println("更新移除的依赖失败，json序列化失败")
		return
	}

	for _, services := range registry.services {
		for _, service := range services {
			for _, name := range service.DependedServicesName {
				if name == reg.ServiceName {
					wg.Add(1)
					go func() {
						resp, err := http.Post(service.UpdateUrl, "application/json", bytes.NewBuffer(p))
						if err != nil {
							fmt.Printf("移除依赖失败，service: %s\n", name)
						}
						if resp.StatusCode != http.StatusOK {
							fmt.Println("依赖更新处理失败")
						}
						wg.Done()
					}()
				}
			}
		}
	}

	wg.Wait()
}

type UpdateInfo struct {
	Add    []Registration
	Remove []Registration
}
