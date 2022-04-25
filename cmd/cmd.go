package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/dshipenok/tickers/pkg"
	"github.com/dshipenok/tickers/pkg/collector"
)

const preiod = time.Second * 5

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	apis := []pkg.IPriceStreamSubscriber{
		pkg.NewMockRandomStream(),
		pkg.NewMockRandomStream(),
		pkg.NewMockRandomStream(),
		pkg.NewMockRandomStream(),
		pkg.NewMockRandomStream(),
	}

	m := pkg.NewMultiplexor()
	output := m.Subscribe(apis)

	c := collector.NewLatest(3)
	p := pkg.NewFairPrice(c, time.Now)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-interrupt
		cancel()
	}()
	outputCh := make(chan pkg.TickerPrice, 1)
	defer close(outputCh)
	go func() {
		for p := range outputCh {
			outputValue(p.Price)
		}
	}()
	p.Start(ctx, output, preiod, outputCh)
}

func outputValue(value string) {
	tm := time.Now()
	timeStr := tm.Format("02/01 15:04:05")
	fmt.Println(timeStr+",", value)
}
