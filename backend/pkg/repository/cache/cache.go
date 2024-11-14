package cache

import (
	"github.com/CloudDetail/apo/backend/pkg/util"
	"sync"
	"time"
)

type Repo interface {
	AddToken(token string) error
	IsInBlackList(token string) (bool, error)
}

type cache struct {
	blackList sync.Map
}

func (c *cache) IsInBlackList(token string) (bool, error) {
	_, ok := c.blackList.Load(token)
	return ok, nil
}

func (c *cache) AddToken(token string) error {
	c.blackList.Store(token, struct{}{})
	return nil
}

func (c *cache) refreshLoop() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.refreshCache()
		}
	}
}

func (c *cache) refreshCache() {
	c.blackList.Range(func(token, _ any) bool {
		if util.IsExpire(token.(string)) {
			c.blackList.Delete(token)
		}
		return true
	})
}

func New() (Repo, error) {
	c := cache{
		blackList: sync.Map{},
	}
	go c.refreshLoop()
	return &c, nil
}