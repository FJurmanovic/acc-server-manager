package network

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"

	"gorm.io/gorm"
)

const (
	defaultPortPoolMin = 9600
	defaultPortPoolMax = 9700
)

type PortPoolManager struct {
	mu    sync.Mutex
	min   int
	max   int
	inUse map[int]struct{}
}

func NewPortPoolManager(db *gorm.DB) (*PortPoolManager, error) {
	min := portEnvInt("PORT_POOL_MIN", defaultPortPoolMin)
	max := portEnvInt("PORT_POOL_MAX", defaultPortPoolMax)

	var ports []int
	if err := db.Raw("SELECT port FROM servers WHERE platform = 'docker' AND port > 0").Scan(&ports).Error; err != nil {
		return nil, fmt.Errorf("failed to seed port pool from database: %v", err)
	}

	inUse := make(map[int]struct{}, len(ports))
	for _, p := range ports {
		inUse[p] = struct{}{}
	}

	return &PortPoolManager{
		min:   min,
		max:   max,
		inUse: inUse,
	}, nil
}

func (p *PortPoolManager) Acquire(_ context.Context) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for port := p.min; port <= p.max; port++ {
		if _, used := p.inUse[port]; !used {
			p.inUse[port] = struct{}{}
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in pool [%d-%d]", p.min, p.max)
}

func (p *PortPoolManager) Release(_ context.Context, port int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.inUse, port)
}

func portEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}
