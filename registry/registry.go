package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type DrpcRegistry struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr      string
	startTime time.Time
}

const (
	defaultPath    = "/_drpc_/registry"
	defaultTimeOut = time.Minute * 2
)

func New(timeout time.Duration) *DrpcRegistry {
	return &DrpcRegistry{
		timeout: timeout,
		servers: make(map[string]*ServerItem),
	}
}

var DefaultDrpcRegister = New(defaultTimeOut)

// putServers 添加服务实例，
func (r *DrpcRegistry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	server := r.servers[addr]
	if server == nil {
		r.servers[addr] = &ServerItem{Addr: addr, startTime: time.Now()}
	} else {
		server.startTime = time.Now()
	}
}

// aliveServers
func (r *DrpcRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		if r.timeout == 0 || s.startTime.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

// run at /_drpc_/registry
func (r *DrpcRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Set("X-Drpc-Servers", strings.Join(r.aliveServers(), ","))
	case http.MethodPost:
		addr := req.Header.Get("X-Drpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
		log.Printf("rpc registry: putServer=%s, Numbers=%d\n", addr, len(r.servers))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// HandleHTTP registers a HTTP handler for DrpcRegistry messages on registryPath
func (r *DrpcRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path: ", registryPath)
}

func HandleHTTP() {
	DefaultDrpcRegister.HandleHTTP(defaultPath)
}

// Heartbeat 提供 HeartBeat 方法，便于服务启动时定时向注册中心发送心跳，默认周期比注册中心设置过期的时间少1min
func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		// make sure there is enough time to send heart beat
		// before it's removed from registry
		duration = defaultTimeOut - time.Minute*time.Duration(1)
	}
	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		ticker := time.NewTicker(duration)
		for err == nil {
			<-ticker.C
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodPost, registry, nil)
	req.Header.Set("X-Drpc-Server", addr)
	if _, err := client.Do(req); err != nil {
		log.Println("rpc server: heart beat err: ", err)
		return err
	}
	return nil
}
