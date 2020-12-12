package server

import (
	gosocketio "github.com/tystuyfzand/gosf-socketio"
	"sync"
)

type Connection struct {
	channel *gosocketio.Channel
	code    string
}

type ConnectionMap struct {
	m map[string]*Connection
	l *sync.RWMutex
}

func (c *ConnectionMap) Set(id string, conn *Connection) {
	c.l.Lock()
	defer c.l.Unlock()

	c.m[id] = conn
}

func (c *ConnectionMap) Get(id string) *Connection {
	c.l.RLock()
	defer c.l.RUnlock()

	return c.m[id]
}

func (c *ConnectionMap) Remove(id string) {
	c.l.Lock()
	defer c.l.Unlock()

	delete(c.m, id)
}
