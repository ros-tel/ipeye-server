package main

import (
	"flag"
	"log"

	"ipeye-server/api"
	"ipeye-server/config"
	"ipeye-server/server"

	"github.com/gin-gonic/gin"
)

var (
	cnf = &config.Conf{}

	configFile = flag.String("config", "./config/config.yml", "Usage: -config=<config_file>")
	debug      = flag.Bool("debug", false, "Print debug information on stderr")
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	flag.Parse()

	config.GetConfig(*configFile, cnf)

	if *debug {
		log.Printf("[INFO] Config: %#v", cnf)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	app_api := gin.Default()
	app_api.Use(config.Inject(cnf))

	api.InitRoutes(app_api)

	go app_api.Run(cnf.Listen.Api)

	server.Run(cnf)
}
