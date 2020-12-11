package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gobuffalo/packr/v2"
	"github.com/tystuyfzand/gosf-socketio"
	"github.com/tystuyfzand/gosf-socketio/transport"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
)

var (
	playerIds     = make(map[string]uint64)
	playerIdMutex = &sync.RWMutex{}
)

type stats struct {
	Address   string
	Connected int64
}

type Signal struct {
	Data interface{} `json:"data"`
	To   string      `json:"to,omitempty"`
	From string      `json:"from,omitempty"`
}

func main() {
	offsetBox := packr.New("Offsets", "./offsets")
	assetBox := packr.New("Assets", "./assets")

	str, err := assetBox.FindString("views/index.gohtml")

	if err != nil {
		log.Fatalln("Unable to load index template:", err)
	}

	var connected int64

	m := &ConnectionMap{m: make(map[string]*Connection), l: &sync.RWMutex{}}

	mainTemplate := template.Must(template.New("").Parse(str))

	log.Println("Files:", offsetBox.List())

	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())

	server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		log.Println("New client connected", c.Id())

		atomic.AddInt64(&connected, 1)

		m.Set(c.Id(), &Connection{channel: c})
	})

	server.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		log.Println("Client disconnected", c.Id())

		atomic.AddInt64(&connected, -1)

		playerIdMutex.Lock()
		defer playerIdMutex.Unlock()

		delete(playerIds, c.Id())
	})

	server.On("join", func(c *gosocketio.Channel, code string, id uint64) {
		log.Println("Join", code, id)

		conn := m.Get(c.Id())

		if conn != nil {
			conn.code = code
			c.Join(conn.code)

			c.BroadcastTo(conn.code, "join", []interface{}{c.Id(), id})

			idMap := make(map[string]uint64)

			playerIdMutex.RLock()

			for _, ch := range c.List(conn.code) {
				idMap[ch.Id()] = playerIds[c.Id()]
			}

			playerIdMutex.RUnlock()

			c.Emit("setIds", idMap)
		}
	})

	server.On("leave", func(c *gosocketio.Channel) {
		conn := m.Get(c.Id())

		if conn != nil && conn.code != "" {
			c.Leave(conn.code)
			conn.code = ""
		}
	})

	server.On("id", func(c *gosocketio.Channel, id uint64) {
		log.Println("Id", id)
		playerIdMutex.Lock()
		defer playerIdMutex.Unlock()

		playerIds[c.Id()] = id

		conn := m.Get(c.Id())

		if conn != nil {
			c.BroadcastTo(conn.code, "setId", []interface{}{c.Id(), id})
		}
	})

	server.On("signal", func(c *gosocketio.Channel, signal Signal) {
		log.Println("Signal", signal.To, signal.Data)

		ch, err := server.GetChannel(signal.To)

		if err == nil {
			ch.Emit("signal", Signal{Data: signal.Data, From: c.Id()})
		}
	})

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Handle("/socket.io/", server)

	offsetHandler := http.FileServer(offsetBox)

	for _, name := range offsetBox.List() {
		r.Handle("/"+name, offsetHandler)
	}

	logoBytes, _ := assetBox.Find("images/logo.png")

	// TODO better way to do this?
	r.Get("/logo.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(logoBytes)))

		w.Write(logoBytes)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		mainTemplate.Execute(w, stats{
			Address:   r.Proto + "://" + r.Host,
			Connected: atomic.LoadInt64(&connected),
		})
	})

	address := ":9736"

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		address = envAddress
	}

	log.Fatalln(http.ListenAndServe(address, r))
}
