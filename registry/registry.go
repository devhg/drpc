package registry

import (
	"log"
	"net/http"
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

	return nil
}

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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *DrpcRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path: ", registryPath)
}

func HandleHTTP() {
	DefaultDrpcRegister.HandleHTTP(defaultPath)
}

func Heartbeat() {

}

func sendHeartbeat() {

}
