package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type client struct {
	conn      *websocket.Conn
	logStream io.ReadCloser

	writeWait      time.Duration
	pongWait       time.Duration
	pingPeriod     time.Duration
	maxMessageSize int64
}

func (c *client) readPump(ctx context.Context, stopCh chan struct{}) {
	defer func() {
		_ = c.conn.Close()
		stopCh <- struct{}{}
	}()
	c.conn.SetReadLimit(c.maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(c.pongWait)); return nil })
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, _, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
					log.Printf("error: %v", err)
				}
				return
			}
		}
	}
}

func (c *client) writePump(ctx context.Context, stopCh chan struct{}) {
	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
		stopCh <- struct{}{}
	}()

	reader := bufio.NewReader(c.logStream)
	for {
		select {
		case <-ctx.Done():
		case <-stopCh:
			return
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		default:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			line, err := readLongLine(reader)
			if err != nil {
				log.Println("writePump: cannot read line", err)
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, line); err != nil {
				log.Println("writePump: cannot write message", err)
				continue
			}
		}
	}
}

func readLongLine(r *bufio.Reader) (line []byte, err error) {
	var buffer []byte
	var isPrefix bool

	for {
		buffer, isPrefix, err = r.ReadLine()
		line = append(line, buffer...)
		if err != nil {
			break
		}

		if !isPrefix {
			break
		}
	}

	return line, err
}
