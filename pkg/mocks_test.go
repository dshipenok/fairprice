package pkg

import (
	"errors"
	"time"
)

func valChannels(vals ...interface{}) (chan TickerPrice, chan error) {
	priceCh := make(chan TickerPrice)
	errCh := make(chan error)
	go func() {
		for _, val := range vals {
			switch typed := val.(type) {
			case *TickerPrice:
				priceCh <- *typed
			case error:
				errCh <- typed
			case string:
				errCh <- errors.New(typed)
			case time.Duration:
				<-time.After(typed)
			}
		}
	}()
	return priceCh, errCh
}

type mockStream struct {
	priceCh chan TickerPrice
	errCh   chan error
}

// newMockStream constructor
func newMockStream(vals ...interface{}) *mockStream {
	priceCh, errCh := valChannels(vals...)
	return &mockStream{
		priceCh: priceCh,
		errCh:   errCh,
	}
}

func (m *mockStream) SubscribePriceStream(Ticker) (chan TickerPrice, chan error) {
	return m.priceCh, m.errCh
}
