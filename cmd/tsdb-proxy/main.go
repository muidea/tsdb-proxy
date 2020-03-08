package main

import (
	"flag"
	"log"

	engine "github.com/muidea/magicEngine"
	"supos.ai/data-lake/external/tsdb-proxy/core"
)

var bindPort = "8080"
var configFile = "config.json"

func main() {
	flag.StringVar(&bindPort, "ListenPort", bindPort, "listen port.")
	flag.StringVar(&configFile, "ConfigFile", configFile, "config file.")
	flag.Parse()

	log.Println("TsDB-Proxy V1.0")

	router := engine.NewRouter()
	service, err := core.New(configFile)

	if err == nil {
		err = service.Startup(router)
		if err == nil {
			svr := engine.NewHTTPServer(bindPort)
			svr.Bind(router)

			svr.Run()

			service.Teardown()
		}
	} else {
		log.Printf("start service faield, err:%s", err.Error())
	}
}
