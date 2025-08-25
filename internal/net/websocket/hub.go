package websocket

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pmoieni/rmx/internal/net/msg"
)

func read(conn *TransportHandler, cli *Hub) {
	defer func() {
		cli.unregister <- conn
		err := conn.rwc.Close()
		if err != nil {
			slog.Error("conn close: %v", err)
		}
		slog.Debug("read: conn closed")
	}()

	if err := conn.setReadDeadLine(pongWait); err != nil {
		slog.Info("setReadDeadLine: %v\n", err)
		return
	}

	for {
		wsMsg, err := conn.read()
		if err != nil {
			// TODO: handle error
			slog.Error("read err: %v\n", err)
			break
		}

		// TODO: add a way use custom read validation here unsure how yet
		var envelope msg.Envelope
		slog.Info("read msg: OpCode: %v\n\n", wsMsg.OpCode)
		if err := json.Unmarshal(wsMsg.Payload, &envelope); err != nil {
			slog.Error("wsMsg unmarshal: %v", err)
		} else {
			slog.Error("read msg:\nType: %d\nID: %s\nUserID: %s\n\n", envelope.Typ, envelope.ID, envelope.UserID)
		}

		cli.broadcast <- wsMsg
	}
}

func write(conn *TransportHandler) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		slog.Debug("write: conn closed")
		if err := conn.rwc.Close(); err != nil {
			slog.Error("error closing connection: %v", err)
		}
	}()

	for {
		select {
		case msg, ok := <-conn.send:
			_ = conn.setWriteDeadLine(writeWait)
			if !ok {
				slog.Error("<-conn.send not ok")
				_ = conn.write(&wsutil.Message{OpCode: ws.OpClose, Payload: []byte{}})
				return
			}

			if err := conn.write(msg); err != nil {
				slog.Error("msg err: %v\n", err)
				return
			}
		case <-ticker.C:
			_ = conn.setWriteDeadLine(writeWait)
			if err := conn.write(&wsutil.Message{OpCode: ws.OpPing, Payload: nil}); err != nil {
				slog.Error("ticker err: %v\n", err)
				return
			}
		}
	}
}

type Hub struct {
	register, unregister chan *TransportHandler
	broadcast            chan *wsutil.Message
	lock                 *sync.Mutex
	connections          map[*TransportHandler]bool
	upgrader             *ws.HTTPUpgrader

	// Capacity of the send channel.
	// If capacity is 0, the send channel is unbuffered.
	Capacity uint
}

// Len returns the number of connections.
func (cli *Hub) Len() int {
	cli.lock.Lock()
	defer cli.lock.Unlock()
	return len(cli.connections)
}

// TODO -- should be able to close all connections via their own channels
func (cli *Hub) Close() error {
	defer func() {
		slog.Info("cli.Close()")
		// close channels
		close(cli.register)
		close(cli.unregister)
		close(cli.broadcast)
	}()

	cli.broadcast <- &wsutil.Message{OpCode: ws.OpClose, Payload: []byte{}} // broadcast close
	return nil
}

/*
NewClient instantiates a new websocket client.

NOTE: these may be useful to set: Capacity, ReadBufferSize, ReadTimeout, WriteTimeout
*/
func NewHub(cap uint) *Hub {
	cli := &Hub{
		register:    make(chan *TransportHandler),
		unregister:  make(chan *TransportHandler),
		broadcast:   make(chan *wsutil.Message),
		lock:        &sync.Mutex{},
		connections: make(map[*TransportHandler]bool),
		upgrader:    &ws.HTTPUpgrader{
			// TODO: may be fields here that worth setting
		},
		Capacity: cap,
	}

	go cli.listen()
	return cli
}

func (cli *Hub) listen() {
	for {
		select {
		case conn := <-cli.register:
			cli.lock.Lock()
			cli.connections[conn] = true
			cli.lock.Unlock()
		case conn := <-cli.unregister:
			slog.Debug("unregister channel handler")
			delete(cli.connections, conn)
			close(conn.send)
		case msg := <-cli.broadcast:
			for conn := range cli.connections {
				select {
				case conn.send <- msg:
				default:
					// From Gorilla WS
					// https://github.com/gorilla/websocket/tree/master/examples/chat#hub
					// If the client’s send buffer is full, then the hub assumes that the client is dead or stuck. In this case, the hub unregisters the client and closes the websocket
					slog.Debug("conn.send channel buffer possible full\n")
					slog.Debug("broadcast channel handler: default case:\nopCode: %d\npayload: %+v\n", msg.OpCode, msg.Payload)
					close(conn.send)
					delete(cli.connections, conn)
				}
			}
		}
	}
}

func (cli *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO check capacity
	if cli.Capacity > 0 && cli.Len() >= int(cli.Capacity) {
		http.Error(w, "too many connections", http.StatusServiceUnavailable)
		return
	}

	rwc, _, _, err := cli.upgrader.Upgrade(r, w)
	if err != nil {
		// TODO log that there was an error
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn := &TransportHandler{
		rwc:  rwc,
		send: make(chan *wsutil.Message, 256),
	}

	cli.register <- conn

	go read(conn, cli)
	go write(conn)
}
