package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

var (
	errUnexpectedStatusCode = errors.New("unexpected status code")
	courses                 = big.NewInt(5)
)

func w(ctx context.Context, c *websocket.Conn, m string, cid int) error {
	log.Printf("%d <- %s", cid, m)
	err := c.Write(ctx, websocket.MessageText, []byte(m))
	if err != nil {
		return fmt.Errorf("error writing to connection: %w", err)
	}
	return nil
}

func connect(cid int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, r, err := websocket.Dial(ctx, "wss://localhost.runxiyu.org:8080/ws", nil)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := c.CloseNow()
		if err != nil {
			panic(err)
		}
	}()
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	if r.StatusCode != http.StatusSwitchingProtocols {
		panic(errUnexpectedStatusCode)
	}

	err = w(ctx, c, "HELLO", cid)
	if err != nil {
		panic(err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				cancel()
				log.Printf("%d !R %v", cid, r)
			}
		}()
		for {
			_, msg, err := c.Read(ctx)
			if err != nil {
				panic(err)
			}
			log.Printf("%d -> %s", cid, string(msg))
		}
	}()

	courseID, err := rand.Int(rand.Reader, courses)
	if err != nil {
		panic(err)
	}
	err = w(ctx, c, fmt.Sprintf("Y %d", courseID.Int64()+1), cid)
	if err != nil {
		panic(err)
	}

	time.Sleep(30 * time.Second)

	err = c.Close(websocket.StatusNormalClosure, "")
	if err != nil {
		panic(err)
	}
}

func main() {
	var wg sync.WaitGroup
	for i := range 800 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("%d !M %v", i, r)
				}
			}()
			connect(i)
		}()
		time.Sleep(2 * time.Millisecond)
	}
	wg.Wait()
}
