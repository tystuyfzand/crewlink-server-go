package main

import (
	"github.com/spf13/viper"
	"log"
	"meow.tf/crewlink-server"
)

func main() {
	viper.SetDefault("address", ":9736")
	viper.SetDefault("name", "CrewLink-Go")

	viper.AutomaticEnv()

	opts := []server.Option{
		server.WithName(viper.GetString("name")),
	}

	s := server.NewServer(opts...)

	s.Setup()

	err := s.Start(viper.GetString("address"))

	if err != nil {
		log.Fatalln("Unable to start server:", err)
	}
}
