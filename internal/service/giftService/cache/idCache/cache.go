package idCache

import (
	"errors"
	"sync"

	"github.com/gotd/td/tg"
)

type IDCacheImpl struct {
	users    map[int64]*tg.User
	channels map[int64]*tg.Channel
	mu       sync.RWMutex
}

func NewIDCache() *IDCacheImpl {
	return &IDCacheImpl{
		users:    make(map[int64]*tg.User),
		channels: make(map[int64]*tg.Channel),
	}
}

func (c *IDCacheImpl) SetUser(user *tg.User) {
	if user == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.users[user.ID] = user
}

func (c *IDCacheImpl) GetUser(id int64) (*tg.User, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	user, ok := c.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (c *IDCacheImpl) SetChannel(channel *tg.Channel) {
	if channel == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.channels[channel.ID] = channel
}

func (c *IDCacheImpl) GetChannel(id int64) (*tg.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	channel, ok := c.channels[id]
	if !ok {
		return nil, errors.New("channel not found")
	}
	return channel, nil
}
