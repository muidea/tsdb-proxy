package core

import (
	"fmt"
	"net/http"
	"time"

	engine "github.com/muidea/magicEngine"
	"supos.ai/data-lake/external/tsdb-proxy/core/database"
	"supos.ai/data-lake/external/tsdb-proxy/core/database/pi"
	"supos.ai/data-lake/external/tsdb-proxy/core/database/std"
	"supos.ai/data-lake/external/tsdb-proxy/model"
)

// New create new batis service
func New(configFile string) (ret *Service) {

	ret = &Service{dbInfoMap: map[string]database.DB{}, configFile: configFile}

	return
}

// Service core service
type Service struct {
	configFile string
	dbInfoMap  map[string]database.DB
}

func (s *Service) loadCfg() (err error) {
	cfgFile := model.ConfigInfo{}
	err = cfgFile.Load(s.configFile)
	if err != nil {
		return
	}

	for idx := range cfgFile.StdService {
		info := cfgFile.StdService[idx]
		_, ok := s.dbInfoMap[info.Name]
		if ok {
			err = fmt.Errorf("duplicate database ,name:%s", info.Name)
			return
		}

		s.dbInfoMap[info.Name] = std.NewStd(info)
	}

	for idx := range cfgFile.PiService {
		info := cfgFile.PiService[idx]
		_, ok := s.dbInfoMap[info.Name]
		if ok {
			err = fmt.Errorf("duplicate database ,name:%s", info.Name)
			return
		}

		s.dbInfoMap[info.Name] = pi.NewPi(info)
	}

	return
}

// Startup startup service
func (s *Service) Startup(router engine.Router) (err error) {
	queryRoute := engine.CreateRoute("/query", "GET", s.queryHandle)
	router.AddRoute(queryRoute)

	notifyRoute := engine.CreateRoute("/notify/:source", "POST", s.notifyHandle)
	router.AddRoute(notifyRoute)

	pingRoute := engine.CreateRoute("/ping", "GET", s.pingHandle)
	router.AddRoute(pingRoute)

	go s.timeCheck()

	return
}

// Teardown teardown service
func (s *Service) Teardown() {
}

func (s *Service) pingHandle(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("X-InfluxDB-Version", "tsdb-proxy")
	res.WriteHeader(http.StatusNoContent)
}

// from client
func (s *Service) queryHandle(res http.ResponseWriter, req *http.Request) {
}

// from database
func (s *Service) notifyHandle(res http.ResponseWriter, req *http.Request) {

}

func (s *Service) timeCheck() {
	timeOutTimer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeOutTimer.C:
			s.checkHealth()
		}
	}
}

func (s *Service) checkHealth() {
	for _, v := range s.dbInfoMap {
		v.CheckHealth()
	}
}
