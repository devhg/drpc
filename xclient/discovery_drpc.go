package xclient

import "time"

type DrpcRegistryDiscovery struct {
	*MultiServerDiscovery
	registry   string    // 服务中心
	timeout    string    // 服务列表过期时间
	lastUpdate time.Time // 最后从服务中心更新服务的时间
}

const defaultLastUpdate = time.Second * 2

// NewDrpcRegistryDiscovery
func NewDrpcRegistryDiscovery(registry string, timeout time.Duration) *DrpcRegistryDiscovery {
	return &DrpcRegistryDiscovery{}
}

// Update update server list
func (dr *DrpcRegistryDiscovery) Update(servers []string) error {
	return nil
}

// Refresh
func (dr *DrpcRegistryDiscovery) Refresh() error {
	return nil
}

// Get 和 GetAll 与 MultiServersDiscovery 相似，
// 唯一的不同在于，GeeRegistryDiscovery 需要先调用 Refresh 确保服务列表没有过期。
func (dr *DrpcRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := dr.Refresh(); err != nil {
		return "", nil
	}
	return dr.MultiServerDiscovery.Get(mode)
}

func (dr *DrpcRegistryDiscovery) GetAll() ([]string, error) {
	if err := dr.Refresh(); err != nil {
		return nil, err
	}
	return dr.MultiServerDiscovery.GetAll()
}
