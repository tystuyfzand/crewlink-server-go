package server

import (
	"github.com/pion/turn/v2"
	"github.com/sethvargo/go-password/password"
	"net"
	"sync"
)

type TURNServer struct {
	server        *turn.Server
	userMap       map[string][]byte
	userLock      *sync.RWMutex
	realm         string
	publicAddress string
}

func NewTURNServer() *TURNServer {
	return &TURNServer{}
}

func (t *TURNServer) Start(addr string) error {
	udpListener, err := net.ListenPacket("udp4", addr)

	if err != nil {
		return err
	}

	t.server, err = turn.NewServer(turn.ServerConfig{
		Realm: t.realm,
		// Set AuthHandler callback
		// This is called everytime a user tries to authenticate with the TURN server
		// Return the key for that user, or false when no user is found
		AuthHandler: t.authHandler,
		// PacketConnConfigs is a list of UDP Listeners and the configuration around them
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: net.ParseIP(t.publicAddress), // Claim that we are listening on IP passed by user (This should be your Public IP)
					Address:      "0.0.0.0",                    // But actually be listening on every interface
				},
			},
		},
	})

	return err
}

func (t *TURNServer) AddUser(username string) (string, error) {
	res, err := password.Generate(64, 10, 10, false, false)

	if err != nil {
		return "", err
	}

	t.userLock.Lock()
	defer t.userLock.Unlock()

	t.userMap[username] = turn.GenerateAuthKey(username, t.realm, res)

	return res, nil
}

func (t *TURNServer) RemoveUser(username string) {
	t.userLock.Lock()
	defer t.userLock.Unlock()

	delete(t.userMap, username)
}

func (t *TURNServer) authHandler(username, realm string, srcAddr net.Addr) ([]byte, bool) {
	t.userLock.RLock()
	defer t.userLock.RUnlock()

	if key, ok := t.userMap[username]; ok {
		return key, true
	}

	return nil, false
}
