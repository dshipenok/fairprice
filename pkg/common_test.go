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
	testDuration = time.Millisecond
	pause        = time.Microsecond * 100

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
			expect: []string{},
		},
		{
			desc: "no values, several periods",
			streams: [][]interface{}{{
				testDuration,
				testDuration + time.Microsecond,
				testDuration + time.Microsecond,
				"Disconnected",
			}},
			expect: []string{"no value", "no value", "no value"},
		},
		{
			desc: "single value - correct result",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
				pause,
				"Disconnected",
			}},
			expect: []string{"1.000"},
		},
		{
			desc: "two values - got average",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "2.0"},
				pause,
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
				pause,
				"Disconnected",
			}},
			expect: []string{"1.200"},
		},
		{
			desc: "two periods - got two values",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
				&TickerPrice{Time: fixedTimeNow().Add(1 * time.Hour), Price: "1.2"},
				testDuration + 1,
				&TickerPrice{Time: fixedTimeNow().Add(2 * time.Hour), Price: "1.2"},
				&TickerPrice{Time: fixedTimeNow().Add(3 * time.Hour), Price: "1.4"},
				pause,
				"Disconnected",
			}},
			expect: []string{"1.100", "1.300"},
		},
		{
			desc: "two sources - got two values",
			streams: [][]interface{}{
				{
					&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
					testDuration + time.Microsecond,
					&TickerPrice{Time: fixedTimeNow().Add(1 * time.Hour), Price: "1.2"},
					pause,
					"Disconnected 1",
				},
				{
					&TickerPrice{Time: fixedTimeNow().Add(2 * time.Hour), Price: "1.2"},
					testDuration + time.Microsecond,
					&TickerPrice{Time: fixedTimeNow().Add(3 * time.Hour), Price: "1.4"},
					pause,
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
			result := []string{}
			output := func(val string) {
				result = append(result, val)
			}

			p.Start(context.Background(), resultCh, testDuration, output)

			require.Len(t, result, len(tst.expect))
			assert.EqualValues(t, tst.expect, result, "got values %+v", tst.expect)
		})
	}
}

func Test_TableLatestPrice(t *testing.T) {
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
			expect: []string{},
		},
		{
			desc: "no values, several periods",
			streams: [][]interface{}{{
				testDuration + time.Microsecond,
				testDuration + time.Microsecond,
				testDuration + time.Microsecond,
				"Disconnected",
			}},
			expect: []string{"no value", "no value", "no value"},
		},
		{
			desc: "single value - correct result",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
				pause,
				"Disconnected",
			}},
			expect: []string{"1.000"},
		},
		{
			desc: "two values - got latest",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour + 1), Price: "1.0"},
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "2.0"},
				pause,
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
				pause,
				"Disconnected",
			}},
			expect: []string{"1.400"},
		},
		{
			desc: "two periods - got two values",
			streams: [][]interface{}{{
				&TickerPrice{Time: fixedTimeNow().Add(3 * time.Hour), Price: "1.4"},
				&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
				testDuration + time.Microsecond,
				&TickerPrice{Time: fixedTimeNow().Add(1 * time.Hour), Price: "1.2"},
				&TickerPrice{Time: fixedTimeNow().Add(2 * time.Hour), Price: "1.3"},
				pause,
				"Disconnected",
			}},
			expect: []string{"1.400", "1.300"},
		},
		{
			desc: "two sources - got two values",
			streams: [][]interface{}{
				{
					&TickerPrice{Time: fixedTimeNow().Add(time.Hour), Price: "1.0"},
					testDuration + time.Microsecond,
					&TickerPrice{Time: fixedTimeNow().Add(1 * time.Hour), Price: "1.2"},
					pause,
					"Disconnected 1",
				},
				{
					&TickerPrice{Time: fixedTimeNow().Add(2 * time.Hour), Price: "1.2"},
					testDuration + time.Microsecond,
					&TickerPrice{Time: fixedTimeNow().Add(3 * time.Hour), Price: "1.4"},
					pause,
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
			result := []string{}
			output := func(val string) {
				result = append(result, val)
			}

			p.Start(context.Background(), resultCh, testDuration, output)

			require.Len(t, result, len(tst.expect))
			assert.EqualValues(t, tst.expect, result, "got values %+v", tst.expect)
		})
	}
}

func valuesToStreams(blocks [][]interface{}) (result []IPriceStreamSubscriber) {
	for _, block := range blocks {
		result = append(result, newMockStream(block...))
	}
	return result
}
