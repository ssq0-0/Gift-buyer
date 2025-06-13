package balanceCache

import (
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"sync/atomic"
)

type BalanceCacheImpl struct {
	balance atomic.Int64
}

func NewBalanceCache() giftInterfaces.BalanceCache {
	return &BalanceCacheImpl{}
}

func (c *BalanceCacheImpl) SetBalance(balance int64) {
	c.balance.Store(balance)
}

func (c *BalanceCacheImpl) GetBalance() int64 {
	return c.balance.Load()
}

func (c *BalanceCacheImpl) TrimBalance(deduction int64) {
	c.balance.Add(-deduction)
}
