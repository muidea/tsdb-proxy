package core

import (
	"fmt"
	"net/http"

	engine "github.com/muidea/magicEngine"
	"supos.ai/data-lake/external/tsdb-proxy/core/database"
	"supos.ai/data-lake/external/tsdb-proxy/model"
)

// New create new batis service
func New(configFile string) (ret *Service, err error) {
	cfgFile := model.ConfigInfo{}
	err = cfgFile.Load(configFile)
	if err != nil {
		return
	}

	svr := &Service{dbInfoMap: map[string]database.DB{}}

	for idx := range cfgFile.StdService {
		info := cfgFile.StdService[idx]
		_, ok := svr.dbInfoMap[info.Name]
		if ok {
			err = fmt.Errorf("duplicate database ,name:%s", info.Name)
			return
		}

		svr.dbInfoMap[info.Name] = database.NewStd(info)
	}

	for idx := range cfgFile.StdService {
		info := cfgFile.StdService[idx]
		_, ok := svr.dbInfoMap[info.Name]
		if ok {
			err = fmt.Errorf("duplicate database ,name:%s", info.Name)
			return
		}

		svr.dbInfoMap[info.Name] = database.NewStd(info)
	}

	ret = svr

	return
}

// Service core service
type Service struct {
	dbInfoMap map[string]database.DB
}

// Startup startup service
func (s *Service) Startup(router engine.Router) (err error) {
	pingRoute := engine.CreateRoute("/ping", "GET", s.pingHandle)
	router.AddRoute(pingRoute)

	queryRoute := engine.CreateRoute("/query", "GET", s.queryHandle)
	router.AddRoute(queryRoute)

	notifyRoute := engine.CreateRoute("/notify/:source", "POST", s.notifyHandle)
	router.AddRoute(notifyRoute)

	return
}

// Teardown teardown service
func (s *Service) Teardown() {
}

func (s *Service) pingHandle(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("X-InfluxDB-Version", "tsdb-proxy")
	res.WriteHeader(http.StatusNoContent)
}

func (s *Service) queryHandle(res http.ResponseWriter, req *http.Request) {
}

func (s *Service) notifyHandle(res http.ResponseWriter, req *http.Request) {

}
