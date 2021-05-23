package xclient

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type DrpcRegistryDiscovery struct {
	*MultiServerDiscovery
	registry   string        // 服务中心
	timeout    time.Duration // 服务列表过期时间
	lastUpdate time.Time     // 最后从服务中心更新服务的时间
}

const defaultLastUpdate = time.Second * 10

// NewDrpcRegistryDiscovery do
func NewDrpcRegistryDiscovery(registry string, timeout time.Duration) *DrpcRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultLastUpdate
	}
	return &DrpcRegistryDiscovery{
		MultiServerDiscovery: NewMultiServerDiscovery(make([]string, 0)), // ?
		registry:             registry,
		timeout:              timeout,
	}
}

// Update update server list
func (dr *DrpcRegistryDiscovery) Update(servers []string) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	dr.servers = servers
	dr.lastUpdate = time.Now()
	return nil
}

// Refresh do
func (dr *DrpcRegistryDiscovery) Refresh() error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	if dr.lastUpdate.Add(dr.timeout).After(time.Now()) {
		return nil
	}
	log.Println("rpc registry: refresh servers from registry:", dr.registry, len(dr.servers))

	resp, err := http.Get(dr.registry)
	if err != nil {
		log.Println("rpc registry refresh err:", err)
		return err
	}

	servers := strings.Split(resp.Header.Get("X-Drpc-Servers"), ",")
	log.Println(servers)
	dr.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		server = strings.TrimSpace(server)
		if server != "" {
			dr.servers = append(dr.servers, server)
		}
	}

	dr.lastUpdate = time.Now()
	return nil
}

// Get 和 GetAll 与 MultiServersDiscovery 相似，
// 唯一的不同在于，GeeRegistryDiscovery 需要先调用 Refresh 确保服务列表没有过期。
func (dr *DrpcRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := dr.Refresh(); err != nil {
		return "", err
	}
	return dr.MultiServerDiscovery.Get(mode)
}

func (dr *DrpcRegistryDiscovery) GetAll() ([]string, error) {
	if err := dr.Refresh(); err != nil {
		return nil, err
	}
	return dr.MultiServerDiscovery.GetAll()
}
