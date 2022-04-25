package pkg

import (
	"context"
	"time"
)

type IFairPriceCollector interface {
	Collect(price string, t time.Time) error
	GetFairPriceAndReset() string
}

type timeNow func() time.Time

type FairPrice struct {
	collector IFairPriceCollector
	timeNow   timeNow
}

// NewFairPrice constructor
func NewFairPrice(collector IFairPriceCollector, tn timeNow) *FairPrice {
	return &FairPrice{
		collector: collector,
		timeNow:   tn,
	}
}

func (p *FairPrice) Start(
	ctx context.Context,
	stream <-chan TickerPrice,
	d time.Duration,
	output chan<- TickerPrice,
) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	startedTime := p.timeNow()

	for {
		//
		select {
		case <-ctx.Done():
			return
		case price, opened := <-stream:
			if !opened {
				return
			}
			// check price is valid
			if price.Time.Before(startedTime) {
				continue
			}
			_ = p.collector.Collect(price.Price, price.Time) // it's safe not to process an error, but it could be logged if required
		case <-ticker.C:
			fairPrice := p.collector.GetFairPriceAndReset()
			select {
			case output <- TickerPrice{Price: fairPrice, Time: p.timeNow()}:
			default:
				// non-blocking operation
			}
			// output <- TickerPrice{Price: fairPrice, Time: p.timeNow()}
			startedTime = p.timeNow()
		}
	}
}
