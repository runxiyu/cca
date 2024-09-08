package main

import (
	"context"
	"log"
	"net/http"

	"github.com/coder/websocket"
)

func handleWs(w http.ResponseWriter, req *http.Request) {
	c, err := websocket.Accept(w, req, &websocket.AcceptOptions{
		Subprotocols: []string{"cca1"},
	})
	if err != nil {
		w.Write([]byte("This endpoint only supports valid WebSocket connections."))
		return
	}
	defer c.CloseNow()

	err = handleConn(req.Context(), c)
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func handleConn(ctx context.Context, c *websocket.Conn) error {
	typ, b, err := c.Read(ctx)
	if err != nil {
		return err
	}

	err = c.Write(ctx, typ, b)
	if err != nil {
		return err
	}

	return nil
}
