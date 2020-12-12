package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"meow.tf/crewlink-server"
	"strings"
)

func main() {
	viper.SetDefault("address", ":9736")
	viper.SetDefault("name", "CrewLink-Go")

	viper.AutomaticEnv()

	viper.SetConfigName("crewlink-server")
	viper.AddConfigPath("/etc/crewlink-server")
	viper.AddConfigPath("$HOME/.crewlink-server")
	viper.AddConfigPath(".")

	// Optionally load config
	viper.ReadInConfig()

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
