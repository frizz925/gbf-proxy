package websocket

import (
	"sync"
	"time"

	"github.com/Frizz925/gbf-proxy/golang/lib/logging"

	httpHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/http"
	"github.com/gorilla/websocket"
)

type Request struct {
	ID      string
	Payload httpHelpers.Request
}

type Response struct {
	ID      string
	Payload httpHelpers.Response
}

func CreatePingHandler(ws *websocket.Conn, writePeriod time.Duration) func(string) error {
	return func(string) error {
		err := ws.SetWriteDeadline(time.Now().Add(writePeriod))
		if err != nil {
			return err
		}
		return ws.WriteMessage(websocket.PongMessage, nil)
	}
}

func CreatePongHandler(ws *websocket.Conn, readPeriod time.Duration) func(string) error {
	return func(string) error {
		return ws.SetReadDeadline(time.Now().Add(readPeriod))
	}
}

func HandlePing(logger *logging.Logger, ws *websocket.Conn, period time.Duration, isRunning func() bool) {
	mutex := &sync.Mutex{}
	running := true
	tick := time.NewTicker(period)

	ws.SetCloseHandler(func(code int, text string) error {
		defer mutex.Unlock()
		mutex.Lock()
		running = false
		return nil
	})

	defer func() {
		tick.Stop()
	}()

	for isRunning() {
		mutex.Lock()
		if !running {
			break
		}
		mutex.Unlock()

		select {
		case <-tick.C:
			err := ws.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				if _, ok := err.(*websocket.CloseError); ok {
					mutex.Lock()
					running = false
					mutex.Unlock()
				}
				logger.Error(err)
			}
		}
	}
}
