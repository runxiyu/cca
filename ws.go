package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5"
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

	session_cookie, err := req.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		c.Write(req.Context(), websocket.MessageText, []byte("U"))
		return
	} else if err != nil {
		c.Write(req.Context(), websocket.MessageText, []byte("E :Error fetching cookie"))
		return
	}

	err = handleConn(req.Context(), c, session_cookie.Value)
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func handleConn(ctx context.Context, c *websocket.Conn, session string) error {
	var userid string
	var expr int
	err := db.QueryRow(context.Background(), "SELECT userid, expr FROM sessions WHERE cookie = $1", session).Scan(&userid, &expr)
	if errors.Is(err, pgx.ErrNoRows) {
		c.Write(ctx, websocket.MessageText, []byte("U"))
		return err
	} else if err != nil {
		c.Write(ctx, websocket.MessageText, []byte("E :Database error"))
		return err
	}

	for {
		typ, b, err := c.Read(ctx)
		if err != nil {
			return err
		}

		fmt.Println(string(b))
		_ = typ
	}

	// err = c.Write(ctx, typ, b)
	// if err != nil {
	// 	return err
	// }

	return nil
}
