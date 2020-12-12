package server

import (
	"crypto/tls"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"github.com/tystuyfzand/gosf-socketio"
	"github.com/tystuyfzand/gosf-socketio/transport"
	"html/template"
	"net"
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
	mux           *chi.Mux
	connected     int64
	connections   *ConnectionMap
	playerIds     map[string]uint64
	playerIdMutex *sync.RWMutex
	startTime     time.Time

	Name              string
	supportedVersions []string
	peerConfig        *PeerConfig
	certificatePath   string
	turnServer        *TURNServer
}

// NewServer constructs a new server with the given options.
func NewServer(opts ...Option) *Server {
	server := &Server{
		connections:   &ConnectionMap{m: make(map[string]*Connection), l: &sync.RWMutex{}},
		playerIds:     make(map[string]uint64),
		playerIdMutex: &sync.RWMutex{},
		mux:           chi.NewRouter(),
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

// WithTURNServer sets the server's TURN Server instance
func WithTURNServer(server *TURNServer) Option {
	return func(s *Server) {
		s.turnServer = server
	}
}

// WithMiddleware passes middleware into chi's mux
func WithMiddleware(middlewares ...func(http.Handler) http.Handler) Option {
	return func(s *Server) {
		s.mux.Use(middlewares...)
	}
}

// WithCertificates enables TLS using the specified directory
func WithCertificates(certificatePath string) Option {
	return func(s *Server) {
		s.certificatePath = certificatePath
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

	s.mux.Use(urlCleaner, schemeSetter)

	s.mux.Handle("/socket.io/", s.server)

	offsetHandler := http.FileServer(offsetBox)

	for _, name := range offsetBox.List() {
		s.mux.Handle("/"+name, offsetHandler)
	}

	logoBytes, _ := assetBox.Find("images/logo.png")

	// TODO better way to do this?
	s.mux.Get("/logo.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(logoBytes)))
		w.Header().Set("Cache-Control", "max-age:290304000, public")
		w.Header().Set("Last-Modified", cacheSince)
		w.Header().Set("Expires", time.Now().AddDate(0, 0, 30).Format(http.TimeFormat))

		w.Write(logoBytes)
	})

	s.mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		mainTemplate.Execute(w, stats{
			Address:   r.URL.Scheme + "://" + r.Host,
			Connected: atomic.LoadInt64(&s.connected),
		})
	})

	s.mux.Get("/health", s.httpHealth)

	log.WithField("address", addr).Info("Listening")

	if s.certificatePath != "" {
		s.mux.Use(middleware.SetHeader("Strict-Transport-Security", "max-age=63072000; includeSubDomains"))

		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}

		srv := &http.Server{
			Addr:         addr,
			Handler:      s.mux,
			TLSConfig:    cfg,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}

		certFile, keyFile, err := findCertificates(s.certificatePath)

		if err != nil {
			return err
		}

		return srv.ListenAndServeTLS(certFile, keyFile)
	}

	return http.ListenAndServe(addr, s.mux)
}

// onConnection handles new socket.io connections
func (s *Server) onConnection(c *gosocketio.Channel) {
	log.WithField("id", c.Id()).Debug("New client connected")

	atomic.AddInt64(&s.connected, 1)

	s.connections.Set(c.Id(), &Connection{channel: c})

	// Upcoming support for peer config
	// See: https://github.com/ottomated/CrewLink-server/pull/28
	// AND https://github.com/ottomated/CrewLink/pull/149
	if s.peerConfig != nil {
		if s.turnServer != nil {
			// Add user for the server and configure it
			password, err := s.turnServer.AddUser(c.Id())

			if err != nil {
				return
			}

			config := *s.peerConfig

			// TODO: c.Request().Host might not be perfect, but should get us a host they can use for now.
			host := c.Request().Host

			if splitHost, _, err := net.SplitHostPort(host); err == nil {
				host = splitHost
			}

			config.TurnServers = []ICEServer{
				{URL: "turn:" + host, Username: c.Id(), Credential: password},
			}

			c.Emit("peerConfig", config)
		} else {
			// Static TURN config
			c.Emit("peerConfig", s.peerConfig)
		}
	}
}

// onDisconnection handles client disconnects, removing their playerId from the list
func (s *Server) onDisconnection(c *gosocketio.Channel) {
	log.WithField("id", c.Id()).Debug("Client disconnected")

	atomic.AddInt64(&s.connected, -1)

	s.connections.Remove(c.Id())

	if s.turnServer != nil {
		s.turnServer.RemoveUser(c.Id())
	}

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

		s.playerIdMutex.Lock()
		s.playerIds[c.Id()] = id
		s.playerIdMutex.Unlock()

		idMap := make(map[string]uint64)

		s.playerIdMutex.RLock()

		var chId string

		for _, ch := range c.List(conn.code) {
			chId = ch.Id()

			if chId == c.Id() {
				continue
			}

			if playerId, ok := s.playerIds[chId]; ok {
				idMap[chId] = playerId
			}
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
