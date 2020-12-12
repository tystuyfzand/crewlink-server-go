package server

// StunServer configures the STUN server for clients
type StunServer struct {
	URL string `json:"url" mapstructure:"url"`
}

// TurnServer configures the TURN server for clients
type TurnServer struct {
	URL        string `json:"url" mapstructure:"url"`
	Username   string `json:"username" mapstructure:"username"`
	Credential string `json:"credential" mapstructure:"credential"`
}

// PeerConfig represents the peerConfig event from crewlink
type PeerConfig struct {
	ForceRelayOnly bool         `json:"forceRelayOnly" mapstructure:"forceRelayOnly"`
	StunServers    []StunServer `json:"stunServers" mapstructure:"stunServers"`
	TurnServers    []TurnServer `json:"turnServers" mapstructure:"turnServers"`
}
