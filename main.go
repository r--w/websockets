package main

import (
	"fmt"
	"log"
	"time"

	"websockets/entity"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	upgrader = websocket.Upgrader{}
)

func hello(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	// ws.SetWriteDeadline(time.Now().Add(1 * time.Second))
	// ws.SetReadDeadline(time.Now().Add(1 * time.Second))
	ws.SetReadLimit(1_000_000)
	// ws.SetCompressionLevel(6)
	// ws.EnableWriteCompression()
	ws.SetPongHandler(func(string) error { c.Logger().Infof("PONG"); return nil })

	errCh := make(chan error)
	// Read
	go func() {
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()

		for {
			message := map[string]string{}
			err := ws.ReadJSON(&message)
			if err != nil {
				c.Logger().Error(err)
				errCh <- err
			}
			c.Logger().Print("Message: %v\n", message)
		}
	}()

	tPing := time.NewTicker(2 * time.Second)
	defer tPing.Stop()

	tSend := time.NewTicker(3 * time.Second)
	defer tSend.Stop()

	for i := 0; ; i++ {
		select {
		case e := <-errCh:
			ws.WriteJSON(map[string]string{
				"error": e.Error(),
			})
		case <-tPing.C:
			c.Logger().Print("PING")
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return err
			}
		case <-tSend.C:
			tick := &entity.Ticker{
				Symbol: "eth",
				Price:  fmt.Sprintf("%d", i),
			}

			if c.Param("format") == "binary" {
				t, _ := proto.Marshal(tick)
				err := ws.WriteMessage(websocket.BinaryMessage, t)
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						c.Logger().Error(err)
						return err
					}
				}
			} else {
				t, _ := protojson.Marshal(tick)
				err := ws.WriteMessage(websocket.TextMessage, t)
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						c.Logger().Error(err)
						return err
					}
				}
			}
			// spew.Dump(proto.Marshal(tick))

		}
	}
}

func main() {
	for i := 0; i < 3; i++ {
		if err := Subscribe(fmt.Sprintf("%d", i)); err != nil {
			log.Fatal(err)
		}
	}
	if err := SubscribeChanGoroutine("chan2"); err != nil {
		log.Fatal(err)
	}
	if err := SubscribeChan("chan"); err != nil {
		log.Fatal(err)
	}

	go Publish()
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	pprof.Register(e)

	e.GET("/ws/:format", hello)
	e.Logger.Fatal(e.Start(":1323"))
}
