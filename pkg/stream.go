package pkg

type stream struct {
	priceCh chan TickerPrice
	errCh   chan error
}

// newStream constructor
func newStream(priceCh chan TickerPrice, errCh chan error) *stream {
	return &stream{
		priceCh: priceCh,
		errCh:   errCh,
	}
}

func (s *stream) shutdown() {
	// close(s.priceCh)
	// close(s.errCh)
}
