package collector

import (
	"strconv"
	"sync"
	"time"
)

type Latest struct {
	m sync.Mutex

	hasPrice    bool
	latestTime  time.Time
	latestPrice float64
	precision   int
}

// NewLatest constructor
func NewLatest(precision int) *Latest {
	return &Latest{
		precision: precision,
	}
}

func (c *Latest) Collect(price string, t time.Time) error {
	c.m.Lock()
	defer c.m.Unlock()

	if c.hasPrice && t.Before(c.latestTime) {
		return nil
	}
	f, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return err
	}
	c.latestTime = t
	c.latestPrice = f
	c.hasPrice = true
	return nil
}

func (c *Latest) GetFairPriceAndReset() string {
	if !c.hasPrice {
		return "no value"
	}
	c.hasPrice = false
	return strconv.FormatFloat(c.latestPrice, 'f', c.precision, 64)
}
