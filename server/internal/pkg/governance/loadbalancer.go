package governance

import (
	"hash/crc32"
	"sync"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	Select(endpoints []string) string
}

// RoundRobinLoadBalancer 轮询负载均衡器
type RoundRobinLoadBalancer struct {
	mu        sync.Mutex
	current   int
	endpoints []string
}

// NewRoundRobinLoadBalancer 创建轮询负载均衡器
func NewRoundRobinLoadBalancer(endpoints []string) *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{
		endpoints: endpoints,
		current:   0,
	}
}

// Select 选择端点（轮询）
func (rr *RoundRobinLoadBalancer) Select(endpoints []string) string {
	if len(endpoints) == 0 {
		return ""
	}

	rr.mu.Lock()
	defer rr.mu.Unlock()

	endpoint := endpoints[rr.current]
	rr.current = (rr.current + 1) % len(endpoints)
	return endpoint
}

// RandomLoadBalancer 随机负载均衡器
type RandomLoadBalancer struct {
	endpoints []string
}

// NewRandomLoadBalancer 创建随机负载均衡器
func NewRandomLoadBalancer(endpoints []string) *RandomLoadBalancer {
	return &RandomLoadBalancer{
		endpoints: endpoints,
	}
}

// Select 选择端点（随机）
func (r *RandomLoadBalancer) Select(endpoints []string) string {
	if len(endpoints) == 0 {
		return ""
	}

	// 使用简单的哈希算法模拟随机
	hash := crc32.ChecksumIEEE([]byte(endpoints[0]))
	return endpoints[int(hash)%len(endpoints)]
}

// ConsistentHashLoadBalancer 一致性哈希负载均衡器
type ConsistentHashLoadBalancer struct {
	endpoints []string
	replicas  int // 虚拟节点数量
}

// NewConsistentHashLoadBalancer 创建一致性哈希负载均衡器
func NewConsistentHashLoadBalancer(endpoints []string, replicas int) *ConsistentHashLoadBalancer {
	if replicas <= 0 {
		replicas = 100 // 默认100个虚拟节点
	}
	return &ConsistentHashLoadBalancer{
		endpoints: endpoints,
		replicas:  replicas,
	}
}

// Select 选择端点（一致性哈希）
func (ch *ConsistentHashLoadBalancer) Select(endpoints []string) string {
	if len(endpoints) == 0 {
		return ""
	}

	// 使用key进行一致性哈希
	// 这里简化实现，实际应该使用完整的哈希环
	key := endpoints[0] // 实际应该传入业务key（如userID、orderID）
	hash := crc32.ChecksumIEEE([]byte(key))
	return endpoints[int(hash)%len(endpoints)]
}

// SelectByKey 根据key选择端点（一致性哈希）
func (ch *ConsistentHashLoadBalancer) SelectByKey(key string) string {
	if len(ch.endpoints) == 0 {
		return ""
	}

	hash := crc32.ChecksumIEEE([]byte(key))
	return ch.endpoints[int(hash)%len(ch.endpoints)]
}

// WeightedRoundRobinLoadBalancer 加权轮询负载均衡器
type WeightedRoundRobinLoadBalancer struct {
	mu        sync.Mutex
	current   int
	endpoints []WeightedEndpoint
}

// WeightedEndpoint 加权端点
type WeightedEndpoint struct {
	Endpoint string
	Weight   int
	Current  int
}

// NewWeightedRoundRobinLoadBalancer 创建加权轮询负载均衡器
func NewWeightedRoundRobinLoadBalancer(endpoints []WeightedEndpoint) *WeightedRoundRobinLoadBalancer {
	return &WeightedRoundRobinLoadBalancer{
		endpoints: endpoints,
		current:   0,
	}
}

// Select 选择端点（加权轮询）
func (wrr *WeightedRoundRobinLoadBalancer) Select(endpoints []string) string {
	if len(wrr.endpoints) == 0 {
		return ""
	}

	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	// 找到当前权重最大的端点
	maxWeight := -1
	selectedIndex := 0

	for i := range wrr.endpoints {
		wrr.endpoints[i].Current += wrr.endpoints[i].Weight
		if wrr.endpoints[i].Current > maxWeight {
			maxWeight = wrr.endpoints[i].Current
			selectedIndex = i
		}
	}

	// 减少选中端点的当前权重
	wrr.endpoints[selectedIndex].Current -= wrr.endpoints[selectedIndex].Weight * len(wrr.endpoints)

	return wrr.endpoints[selectedIndex].Endpoint
}

// LoadBalancerFactory 负载均衡器工厂
type LoadBalancerFactory struct{}

// NewLoadBalancerFactory 创建负载均衡器工厂
func NewLoadBalancerFactory() *LoadBalancerFactory {
	return &LoadBalancerFactory{}
}

// CreateLoadBalancer 创建负载均衡器
func (f *LoadBalancerFactory) CreateLoadBalancer(strategy string, endpoints []string) LoadBalancer {
	switch strategy {
	case "roundRobin":
		return NewRoundRobinLoadBalancer(endpoints)
	case "random":
		return NewRandomLoadBalancer(endpoints)
	case "consistentHash":
		return NewConsistentHashLoadBalancer(endpoints, 100)
	default:
		return NewRoundRobinLoadBalancer(endpoints)
	}
}
