package server

// ICEServer configures the TURN/STUN servers for clients
type ICEServer struct {
	URL        string `json:"url" mapstructure:"url"`
	Username   string `json:"username" mapstructure:"username"`
	Credential string `json:"credential" mapstructure:"credential"`
}

// PeerConfig represents the peerConfig event from crewlink
type PeerConfig struct {
	ForceRelayOnly bool        `json:"forceRelayOnly" mapstructure:"forceRelayOnly"`
	StunServers    []ICEServer `json:"stunServers" mapstructure:"stunServers"`
	TurnServers    []ICEServer `json:"turnServers" mapstructure:"turnServers"`
}
