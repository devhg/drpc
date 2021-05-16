package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

// 服务发现接口

type SelectMode int

const (
	RandomSelect     SelectMode = iota // select randomly
	RoundRobinSelect                   // select using Robbin algorithm
)

type Discovery interface {
	// 从注册中心更新服务列表
	Refresh() error
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

type MultiServerDiscovery struct {
	r       *rand.Rand
	mu      sync.RWMutex
	servers []string
	index   int
}

func NewMultiServerDiscovery(servers []string) *MultiServerDiscovery {
	d := &MultiServerDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

var _ Discovery = (*MultiServerDiscovery)(nil)

// Refresh do what???
func (d *MultiServerDiscovery) Refresh() error {
	return nil
}

// Update updates the servers of discovery dynamically if necessary
func (d *MultiServerDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

// Get gets a server by mode
func (d *MultiServerDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelect:
		server := d.servers[d.index%n]
		d.index = (d.index + 1) % n
		return server, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

// GetAll return all the servers in discovery
func (d *MultiServerDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	servers := make([]string, len(d.servers), cap(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
