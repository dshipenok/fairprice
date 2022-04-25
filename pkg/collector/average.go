package collector

import (
	"strconv"
	"sync"
	"time"
)

type Average struct {
	m sync.Mutex

	prices    []float64
	precision int
}

// NewAverage constructor
func NewAverage(precision int) *Average {
	return &Average{
		precision: precision,
	}
}

func (c *Average) Collect(price string, t time.Time) error {
	f, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return err
	}
	c.m.Lock()
	c.prices = append(c.prices, f)
	c.m.Unlock()
	return nil
}

func (c *Average) GetFairPriceAndReset() string {
	c.m.Lock()
	defer c.m.Unlock()
	count := len(c.prices)
	if count == 0 {
		return "no value"
	}
	sum := 0.
	for _, v := range c.prices {
		sum += v
	}
	c.prices = nil
	final := sum / float64(count)
	return strconv.FormatFloat(final, 'f', c.precision, 64)
}
