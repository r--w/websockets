package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/nats-io/nats.go"
)

var (
	nc *nats.Conn
	ec *nats.EncodedConn
)

type ticker struct {
	Symbol string
	Price  int32
}

func init() {
	var err error

	nc, err = nats.Connect("0.0.0.0:4222",
		nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
			if s != nil {
				log.Printf("Async error in %q/%q: %v", s.Subject, s.Queue, err)
			} else {
				log.Printf("Async error outside subscription: %v", err)
			}
		}))

	ec, err = nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		panic(err)
	}
}

func Subscribe(name string) error {
	var err error
	i := 0
	_, err = ec.Subscribe("tickers.*", func(t *ticker) {
		i += 1
		fmt.Printf("JSON name: %s, message: %v, id: %d \n", name, t, i)
	})
	if err != nil {
		return err
	}
	nc.Flush()

	return nc.LastError()
}

func SubscribeChanGoroutine(name string) error {
	var err error
	ch := make(chan *nats.Msg)
	_, err = nc.Subscribe("tickers.*", func(msg *nats.Msg) {
		ch <- msg
	})
	go func() {
		i := 0
		for msg := range ch {
			i += 1
			var t ticker
			if err := json.Unmarshal(msg.Data, &t); err != nil {
				log.Error(err)
			} else {
				fmt.Printf("name: %s, message: %v, id: %d \n", name, t, i)
			}
		}
	}()

	if err != nil {
		return err
	}
	nc.Flush()

	return nc.LastError()
}

func SubscribeChan(name string) error {
	var err error
	// for unbuffered channel: slow consumer, messages dropped on connection [38] for subscription on "tickers.*"
	// https://github.com/nats-io/nats.go/issues/412
	ch := make(chan *nats.Msg, 1024) // <- needs to be buffered
	_, err = nc.ChanSubscribe("tickers.*", ch)
	if err != nil {
		return err
	}

	go func() {
		i := 0
		for msg := range ch {
			i += 1
			var t = ticker{}
			if err := json.Unmarshal(msg.Data, &t); err != nil {
				log.Error(err)
			} else {
				fmt.Printf("name: %s, message: %v, id: %d \n", name, t, i)
			}
		}
	}()

	nc.Flush()

	return nc.LastError()
}

func Publish() {
	for {
		for _, t := range []string{"btc", "eth", "ada"} {
			err := ec.Publish("tickers."+t, ticker{
				Symbol: t,
				Price:  rand.Int31n(100),
			})

			if err != nil {
				log.Error(err)
				continue

			}
		}
		time.Sleep(1 * time.Second)
	}
}
