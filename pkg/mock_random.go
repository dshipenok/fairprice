package pkg

import (
	"math/rand"
	"strconv"
	"time"
)

const priceRange = 1000

type MockRandomStream struct {
	priceCh chan TickerPrice
	errCh   chan error
}

// NewMockRandomStream constructor
func NewMockRandomStream() *MockRandomStream {
	return &MockRandomStream{
		priceCh: make(chan TickerPrice, 1),
		errCh:   make(chan error),
	}
}

func (m *MockRandomStream) SubscribePriceStream(Ticker) (chan TickerPrice, chan error) {
	go m.generate()
	return m.priceCh, m.errCh
}

func (m *MockRandomStream) generate() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			value := 40000. + rand.Float64()*priceRange - (priceRange * 0.5)
			m.priceCh <- TickerPrice{
				Price: strconv.FormatFloat(value, 'f', 3, 64),
				Time:  time.Now(),
			}
		}
	}
}
