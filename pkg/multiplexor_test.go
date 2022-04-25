package pkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ErrorBeforeResult_ExpectNoResult(t *testing.T) {
	m := NewMultiplexor()
	resultCh := m.Subscribe([]IPriceStreamSubscriber{
		newMockStream(
			"ERROR1",
			&TickerPrice{Time: time.Now(), Price: "1.0"},
		),
	})

	select {
	case price, opened := <-resultCh:
		if opened {
			assert.Fail(t, "Got price after error", price)
		}
	case <-time.After(10 * time.Millisecond):
	}
}

func Test_ErrorsAfterResult_ExpectCorrectValue(t *testing.T) {
	tn := time.Now()
	m := NewMultiplexor()
	resultCh := m.Subscribe([]IPriceStreamSubscriber{
		newMockStream(
			&TickerPrice{Time: tn, Price: "1.0"},
			"ERROR1",
		),
	})

	select {
	case price, opened := <-resultCh:
		if opened {
			assert.EqualValues(t, TickerPrice{Time: tn, Price: "1.0"}, price)
		}
	case <-time.After(10 * time.Millisecond):
	}
}
