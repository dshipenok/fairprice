package pkg

import "sync"

type Multiplexor struct{}

// NewMultiplexor constructor
func NewMultiplexor() *Multiplexor {
	return &Multiplexor{}
}

func (m *Multiplexor) Subscribe(apis []IPriceStreamSubscriber) chan TickerPrice { // TODO: not sure we have to return channel here
	if len(apis) == 0 {
		result := make(chan TickerPrice, 1)
		close(result)
		return result
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(apis))
	output := make(chan TickerPrice, 1)
	for _, api := range apis {
		priceCh, errCh := api.SubscribePriceStream(BTCUSDTicker)
		// goroutine per channel, thanks it's lightweight
		go runStream(output, priceCh, errCh, wg)
	}
	go func() {
		// don't forget to close output channel if all input channels were closed
		wg.Wait()
		close(output)
	}()

	return output
}

func runStream(
	output chan<- TickerPrice,
	priceCh <-chan TickerPrice,
	errCh <-chan error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for {
		select {
		case price, opened := <-priceCh:
			if !opened {
				return
			}
			output <- price
		case <-errCh:
			return
		}
	}
}
