package server

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"github.com/tystuyfzand/gosf-socketio"
	"github.com/tystuyfzand/gosf-socketio/transport"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	cacheSince = time.Now().Format(http.TimeFormat)
)

type Option func(*Server)

// Signal is a WebRTC Signal, serialized with a to/from field.
type Signal struct {
	Data interface{} `json:"data"`
	To   string      `json:"to,omitempty"`
	From string      `json:"from,omitempty"`
}

// Server represents a CrewLink server instance.
type Server struct {
	server        *gosocketio.Server
	connected     int64
	connections   *ConnectionMap
	playerIds     map[string]uint64
	playerIdMutex *sync.RWMutex
	startTime     time.Time

	Name              string
	supportedVersions []string
	peerConfig        *PeerConfig
}

// NewServer constructs a new server with the given options.
func NewServer(opts ...Option) *Server {
	server := &Server{
		connections:   &ConnectionMap{m: make(map[string]*Connection), l: &sync.RWMutex{}},
		playerIds:     make(map[string]uint64),
		playerIdMutex: &sync.RWMutex{},
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

// WithName sets the server's Name variable
func WithName(name string) Option {
	return func(s *Server) {
		s.Name = name
	}
}

// WithVersions sets the server's supported versions
func WithVersions(versions []string) Option {
	return func(s *Server) {
		s.supportedVersions = versions
	}
}

// WithPeerConfig sets the server's peer config
func WithPeerConfig(config *PeerConfig) Option {
	return func(s *Server) {
		s.peerConfig = config
	}
}

// Setup initializes the socket.io server
func (s *Server) Setup() {
	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())

	server.On(gosocketio.OnConnection, s.onConnection)
	server.On(gosocketio.OnDisconnection, s.onDisconnection)
	server.On("join", s.onJoin)
	server.On("leave", s.onLeave)
	server.On("id", s.onId)
	server.On("signal", s.onSignal)

	s.server = server
}

// Start initializes the HTTP routes, and starts the server
func (s *Server) Start(addr string) error {
	if s.server == nil {
		s.Setup()
	}

	s.startTime = time.Now()

	offsetBox := packr.New("Offsets", "./offsets")
	assetBox := packr.New("Assets", "./assets")

	str, err := assetBox.FindString("views/index.gohtml")

	if err != nil {
		return err
	}

	mainTemplate := template.Must(template.New("").Parse(str))

	if s.supportedVersions == nil {
		supportedVersions := offsetBox.List()

		for i, version := range supportedVersions {
			supportedVersions[i] = strings.TrimSuffix(version, ".yml")
		}

		s.supportedVersions = supportedVersions
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(schemeSetter)
	r.Use(urlCleaner)
	r.Handle("/socket.io/", s.server)

	offsetHandler := http.FileServer(offsetBox)

	for _, name := range offsetBox.List() {
		r.Handle("/"+name, offsetHandler)
	}

	logoBytes, _ := assetBox.Find("images/logo.png")

	// TODO better way to do this?
	r.Get("/logo.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(logoBytes)))
		w.Header().Set("Cache-Control", "max-age:290304000, public")
		w.Header().Set("Last-Modified", cacheSince)
		w.Header().Set("Expires", time.Now().AddDate(0, 0, 30).Format(http.TimeFormat))

		w.Write(logoBytes)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		mainTemplate.Execute(w, stats{
			Address:   r.URL.Scheme + "://" + r.Host,
			Connected: atomic.LoadInt64(&s.connected),
		})
	})

	r.Get("/health", s.httpHealth)

	log.WithField("address", addr).Info("Listening")

	return http.ListenAndServe(addr, r)
}

// onConnection handles new socket.io connections
func (s *Server) onConnection(c *gosocketio.Channel) {
	log.WithField("id", c.Id()).Debug("New client connected")

	atomic.AddInt64(&s.connected, 1)

	s.connections.Set(c.Id(), &Connection{channel: c})

	if s.peerConfig != nil {
		// Upcoming support for peer config
		// See: https://github.com/ottomated/CrewLink-server/pull/28
		// AND https://github.com/ottomated/CrewLink/pull/149
		c.Emit("peerConfig", s.peerConfig)
	}
}

// onDisconnection handles client disconnects, removing their playerId from the list
func (s *Server) onDisconnection(c *gosocketio.Channel) {
	log.WithField("id", c.Id()).Debug("Client disconnected")

	atomic.AddInt64(&s.connected, -1)

	s.playerIdMutex.Lock()
	defer s.playerIdMutex.Unlock()

	delete(s.playerIds, c.Id())
}

// onJoin handles lobby joins, setting their room and sending current players
func (s *Server) onJoin(c *gosocketio.Channel, code string, id uint64) {
	conn := s.connections.Get(c.Id())

	if conn != nil {
		log.WithFields(log.Fields{
			"id":   c.Id(),
			"code": code,
		}).Debug("Client joined room")

		conn.code = code
		c.Join(conn.code)

		c.BroadcastTo(conn.code, "join", []interface{}{c.Id(), id})

		idMap := make(map[string]uint64)

		s.playerIdMutex.RLock()

		for _, ch := range c.List(conn.code) {
			if ch.Id() == c.Id() {
				continue
			}

			idMap[ch.Id()] = s.playerIds[ch.Id()]
		}

		s.playerIdMutex.RUnlock()

		c.Emit("setIds", idMap)
	}
}

// onLeave handles clients leaving a lobby/room
func (s *Server) onLeave(c *gosocketio.Channel) {
	conn := s.connections.Get(c.Id())

	if conn != nil && conn.code != "" {
		log.WithFields(log.Fields{
			"id":   c.Id(),
			"code": conn.code,
		}).Debug("Client left room")

		c.Leave(conn.code)
		conn.code = ""
	}
}

// onId is used when a user's id is set
func (s *Server) onId(c *gosocketio.Channel, id uint64) {
	s.playerIdMutex.Lock()
	s.playerIds[c.Id()] = id
	s.playerIdMutex.Unlock()

	conn := s.connections.Get(c.Id())

	if conn != nil {
		log.WithFields(log.Fields{
			"id":     c.Id(),
			"gameId": id,
		}).Debug("Client set id")
		c.BroadcastTo(conn.code, "setId", []interface{}{c.Id(), id})
	}
}

// onSignal handles the WebRTC signal event
func (s *Server) onSignal(c *gosocketio.Channel, signal Signal) {
	ch, err := s.server.GetChannel(signal.To)

	if err == nil {
		ch.Emit("signal", Signal{Data: signal.Data, From: c.Id()})
	}
}
