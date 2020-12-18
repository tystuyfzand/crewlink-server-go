package main

import (
	"github.com/chi-middleware/logrus-logger"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tystuyfzand/proxy"
	"meow.tf/crewlink-server"
	"strings"
)

// setupFlags sets up base pflag values
func setupFlags() {
	pflag.String("address", "", "Server bind address")
	pflag.String("name", "", "Server name")
	pflag.String("trustedProxies", "", "Trusted proxy addresses, comma separated")
	pflag.Bool("logRequests", false, "Flag to enable request logging in the router")
	pflag.String("certificatePath", "", "SSL Certificate Path")
	pflag.String("dataPath", "", "Web Data Path")
	pflag.String("peerConfig", "", "Peer config file path")

	pflag.Parse()
}

// setupConfiguration binds all the viper default values, sets up pflag integration
// and parses the config file, if possible.
func setupConfiguration() {
	viper.SetDefault("address", ":9736")
	viper.SetDefault("name", "CrewLink-Go")
	viper.SetDefault("trustedProxies", "127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16")
	viper.SetDefault("logRequests", false)
	viper.SetDefault("certificatePath", "")
	viper.SetDefault("dataPath", "")

	viper.AutomaticEnv()

	viper.SetConfigName("crewlink-server")
	viper.AddConfigPath("/etc/crewlink-server")
	viper.AddConfigPath("$HOME/.crewlink-server")
	viper.AddConfigPath(".")

	viper.BindPFlags(pflag.CommandLine)

	// Optionally load config
	viper.ReadInConfig()
}

func main() {
	setupFlags()
	setupConfiguration()

	if logLevelStr := viper.GetString("logLevel"); logLevelStr != "" {
		level, err := log.ParseLevel(logLevelStr)

		if err != nil {
			log.WithError(err).Fatalln("Unable to configure log level")
		}

		log.SetLevel(level)
	}

	opts := []server.Option{
		server.WithName(viper.GetString("name")),
	}

	if dataPath := viper.GetString("dataPath"); dataPath != "" {
		opts = append(opts, server.WithDataPath(dataPath))
	}

	if versions := viper.GetString("versions"); versions != "" {
		opts = append(opts, server.WithVersions(strings.Split(versions, ",")))
	}

	if peerConfigFile := viper.GetString("peerConfig"); peerConfigFile != "" {
		peerConfig, err := loadPeerConfig(peerConfigFile)

		if err != nil {
			log.WithError(err).Fatalln("Unable to load peer config")
		}

		opts = append(opts, server.WithPeerConfig(peerConfig))
	}

	if trustedProxies := viper.GetString("trustedProxies"); trustedProxies != "" {
		forwardedOpts := proxy.NewForwardedHeadersOptions()

		networks := strings.Split(trustedProxies, ",")

		for _, network := range networks {
			if strings.Contains(network, "/") {
				forwardedOpts.AddTrustedNetwork(network)
			} else {
				forwardedOpts.AddTrustedProxy(network)
			}
		}

		opts = append(opts, server.WithMiddleware(proxy.ForwardedHeaders(forwardedOpts)))
	}

	if viper.GetBool("logRequests") {
		opts = append(opts, server.WithMiddleware(logger.Logger("router", log.StandardLogger())))
	}

	if certificatePath := viper.GetString("certificatePath"); certificatePath != "" {
		opts = append(opts, server.WithCertificates(certificatePath))
	}

	s := server.NewServer(opts...)

	s.Setup()

	err := s.Start(viper.GetString("address"))

	if err != nil {
		log.WithError(err).Fatalln("Unable to start server")
	}
}

// loadPeerConfig will attempt to load a peer config from a file.
func loadPeerConfig(file string) (*server.PeerConfig, error) {
	v := viper.New()

	v.SetConfigFile(file)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var config server.PeerConfig

	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
