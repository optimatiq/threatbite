package main

import (
	"flag"

	"github.com/labstack/gommon/log"
	"github.com/optimatiq/threatbite/api/transport"
	"github.com/optimatiq/threatbite/config"
)

var (
	date   = "unknown"
	tag    = "dev"
	commit = "unknown"
)

func main() {
	configFile := flag.String("config", "", "a path to the configuration file")
	flag.Parse()

	conf, err := config.NewConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	log.SetPrefix("threatbite")
	if conf.Debug {
		log.SetLevel(log.DEBUG)
	}

	log.Debugf("Starting the app, build date: %s, git tag: %s, git commit: %s", date, tag, commit)

	server, err := transport.NewAPI(conf)
	if err != nil {
		log.Fatal(err)
	}

	server.Run()
}
