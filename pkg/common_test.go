package pkg

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/dshipenok/tickers/pkg/collector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	periodDuration = time.Millisecond

	timeLayout = "2006-01-02T15:04:05.000Z"
	strTimeNow = "2020-01-01T10:05:00.000Z"
)

var (
	tn, _ = time.Parse(timeLayout, strTimeNow)
)

func fixedTimeNow() time.Time {
	return tn
}

func Test_TableAveragePrices(t *testing.T) {
	waitForNextPeriod := make(chan struct{}, 1000)  // for first stream
	waitForNextPeriod2 := make(chan struct{}, 1000) // for second stream
	tsts := []struct {
		desc    string
		streams [][]interface{}
		expect  []string
	}{
		{
			desc: "no values input - no values output",
			streams: [][]interface{}{{
				"Disconnected",
			}},
			expect: nil,
		},
		{
			desc: "no values, several periods",
			streams: [][]interface{}{{
				waitForNextPeriod,
				waitForNextPeriod,
				waitForNextPeriod,
				"Disconnected",
			}},
			expect: []string{"no value", "no value", "no value"},
		},
		{
			desc: "single value - correct result",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
				waitForNextPeriod,
				"Disconnected",
			}},
			expect: []string{"1.000"},
		},
		{
			desc: "two values - got average",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "2.0"},
				waitForNextPeriod,
				"Disconnected",
			}},
			expect: []string{"1.500"},
		},
		{
			desc: "three values - got average",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
				&TickerPrice{Time: fixedTimeNow().Add(2 * time.Hour), Price: "1.2"},
				&TickerPrice{Time: fixedTimeNow().Add(3 * time.Hour), Price: "1.4"},
				waitForNextPeriod,
				"Disconnected",
			}},
			expect: []string{"1.200"},
		},
		{
			desc: "two periods - got two values",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
				&TickerPrice{Time: fixedTimeNow().Add(1 * time.Hour), Price: "1.2"},
				waitForNextPeriod,
				&TickerPrice{Time: fixedTimeNow().Add(2 * time.Hour), Price: "1.2"},
				&TickerPrice{Time: fixedTimeNow().Add(3 * time.Hour), Price: "1.4"},
				waitForNextPeriod,
				"Disconnected",
			}},
			expect: []string{"1.100", "1.300"},
		},
		{
			desc: "two sources - got two values",
			streams: [][]interface{}{
				{
					&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
					waitForNextPeriod,
					&TickerPrice{Time: fixedTimeNow().Add(1 * time.Hour), Price: "1.2"},
					waitForNextPeriod,
					"Disconnected 1",
				},
				{
					&TickerPrice{Time: fixedTimeNow().Add(2 * time.Hour), Price: "1.2"},
					waitForNextPeriod2,
					&TickerPrice{Time: fixedTimeNow().Add(3 * time.Hour), Price: "1.4"},
					waitForNextPeriod2,
					"Disconnected 2",
				}},
			expect: []string{"1.100", "1.300"},
		},
	}
	for index, tst := range tsts {
		t.Run(tst.desc+"/"+strconv.Itoa(index), func(t *testing.T) {
			m := NewMultiplexor()
			resultCh := m.Subscribe(valuesToStreams(tst.streams))
			c := collector.NewAverage(3)
			p := NewFairPrice(c, fixedTimeNow)
			result := []TickerPrice{}
			output, wait := make(chan TickerPrice, 1), make(chan struct{})
			go func() {
				for p := range output {
					result = append(result, p)
					waitForNextPeriod <- struct{}{}
					waitForNextPeriod2 <- struct{}{}
				}
				close(wait)
			}()

			p.Start(context.Background(), resultCh, periodDuration, output)
			close(output)
			<-wait

			pricesResult := toPrices(result)
			require.Len(t, pricesResult, len(tst.expect))
			assert.EqualValues(t, tst.expect, pricesResult, "got values %+v", tst.expect)

		loop:
			for {
				select {
				case <-waitForNextPeriod:
				case <-waitForNextPeriod2:
				default:
					break loop
				}
			}
		})
	}
}

func Test_TableLatestPrice(t *testing.T) {
	waitForNextPeriod := make(chan struct{}, 1000)  // for first stream
	waitForNextPeriod2 := make(chan struct{}, 1000) // for second stream
	tsts := []struct {
		desc    string
		streams [][]interface{}
		expect  []string
	}{
		{
			desc: "no values input - no values output",
			streams: [][]interface{}{{
				"Disconnected",
			}},
			expect: nil,
		},
		{
			desc: "no values, several periods",
			streams: [][]interface{}{{
				waitForNextPeriod,
				waitForNextPeriod,
				waitForNextPeriod,
				"Disconnected",
			}},
			expect: []string{"no value", "no value", "no value"},
		},
		{
			desc: "single value - correct result",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
				waitForNextPeriod,
				"Disconnected",
			}},
			expect: []string{"1.000"},
		},
		{
			desc: "two values - got latest",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour + 1), Price: "1.0"},
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "2.0"},
				waitForNextPeriod,
				"Disconnected",
			}},
			expect: []string{"1.000"},
		},
		{
			desc: "three values - got latest",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
				&TickerPrice{Time: fixedTimeNow().Add(3 * time.Hour), Price: "1.4"},
				&TickerPrice{Time: fixedTimeNow().Add(2 * time.Hour), Price: "1.2"},
				waitForNextPeriod,
				"Disconnected",
			}},
			expect: []string{"1.400"},
		},
		{
			desc: "two periods - got two values",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(3 * time.Hour), Price: "1.4"},
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
				waitForNextPeriod,
				&TickerPrice{Time: fixedTimeNow().Add(1 * time.Hour), Price: "1.2"},
				&TickerPrice{Time: fixedTimeNow().Add(2 * time.Hour), Price: "1.3"},
				waitForNextPeriod,
				"Disconnected",
			}},
			expect: []string{"1.400", "1.300"},
		},
		{
			desc: "two sources - got two values",
			streams: [][]interface{}{
				{
					&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
					waitForNextPeriod,
					&TickerPrice{Time: fixedTimeNow().Add(1 * time.Hour), Price: "1.2"},
					waitForNextPeriod,
					"Disconnected 1",
				},
				{
					&TickerPrice{Time: fixedTimeNow().Add(2 * time.Hour), Price: "1.2"},
					waitForNextPeriod2,
					&TickerPrice{Time: fixedTimeNow().Add(3 * time.Hour), Price: "1.4"},
					waitForNextPeriod2,
					"Disconnected 2",
				}},
			expect: []string{"1.200", "1.400"},
		},
	}
	for index, tst := range tsts {
		t.Run(tst.desc+"/"+strconv.Itoa(index), func(t *testing.T) {
			m := NewMultiplexor()
			resultCh := m.Subscribe(valuesToStreams(tst.streams))
			c := collector.NewLatest(3)
			p := NewFairPrice(c, fixedTimeNow)
			result := []TickerPrice{}
			output, wait := make(chan TickerPrice, 1), make(chan struct{})
			go func() {
				for p := range output {
					result = append(result, p)
					waitForNextPeriod <- struct{}{}
					waitForNextPeriod2 <- struct{}{}
				}
				close(wait)
			}()

			p.Start(context.Background(), resultCh, periodDuration, output)
			close(output)
			<-wait

			pricesResult := toPrices(result)
			require.Len(t, pricesResult, len(tst.expect))
			assert.EqualValues(t, tst.expect, pricesResult, "got values %+v", tst.expect)

		loop:
			for {
				select {
				case <-waitForNextPeriod:
				case <-waitForNextPeriod2:
				default:
					break loop
				}
			}
		})
	}
}

func valuesToStreams(blocks [][]interface{}) (result []IPriceStreamSubscriber) {
	for _, block := range blocks {
		result = append(result, newMockStream(block...))
	}
	return result
}

func toPrices(values []TickerPrice) (result []string) {
	for _, val := range values {
		result = append(result, val.Price)
	}
	return result
}
