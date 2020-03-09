package main

import (
	"flag"
	"log"
	"net"

	engine "github.com/muidea/magicEngine"
	"supos.ai/data-lake/external/tsdb-proxy/core"
)

var bindAddress = "127.0.0.1:8080"
var rtdService = "127.0.0.1:8010"
var configFile = "config.json"

func main() {
	flag.StringVar(&bindAddress, "BindAddress", bindAddress, "bind address.")
	flag.StringVar(&rtdService, "RtdService", rtdService, "rtdService address.")
	flag.StringVar(&configFile, "ConfigFile", configFile, "config file.")
	flag.Parse()

	log.Println("TsDB-Proxy V1.0")

	if bindAddress == "" || rtdService == "" || configFile == "" {
		flag.Usage()
		return
	}

	router := engine.NewRouter()

	_, bindPort, bindErr := net.SplitHostPort(bindAddress)
	if bindErr != nil {
		log.Printf("invalid bindAddress, err:%s", bindErr.Error())
		return
	}

	service := core.New(configFile, bindAddress)
	err := service.Startup(router, rtdService)
	if err == nil {
		svr := engine.NewHTTPServer(bindPort)
		svr.Bind(router)

		svr.Run()

		service.Teardown()
	} else {
		log.Printf("start service faield, err:%s", err.Error())
	}
}
