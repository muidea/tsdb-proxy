package core

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	engine "github.com/muidea/magicEngine"
	"supos.ai/data-lake/external/tsdb-proxy/core/database"
	"supos.ai/data-lake/external/tsdb-proxy/core/database/pi"
	"supos.ai/data-lake/external/tsdb-proxy/core/database/std"
	"supos.ai/data-lake/external/tsdb-proxy/model"
)

// New create new batis service
func New(configFile, bindAddress string) (ret *Service) {

	ret = &Service{dbInfoMap: map[string]database.DB{}, configFile: configFile, bindAddress: bindAddress}

	return
}

// Service core service
type Service struct {
	configFile  string
	bindAddress string
	dbInfoMap   map[string]database.DB
}

func (s *Service) constructCallBack(dbName string) string {
	return strings.Join([]string{"http:/", s.bindAddress, "notify", dbName}, "/")
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

		s.dbInfoMap[info.Name] = std.NewStd(info, s.constructCallBack(info.Name))
	}

	for idx := range cfgFile.PiService {
		info := cfgFile.PiService[idx]
		_, ok := s.dbInfoMap[info.Name]
		if ok {
			err = fmt.Errorf("duplicate database ,name:%s", info.Name)
			return
		}

		s.dbInfoMap[info.Name] = pi.NewPi(info, s.constructCallBack(info.Name))
	}

	return
}

// Startup startup service
func (s *Service) Startup(router engine.Router, rtdService string) (err error) {
	err = s.loadCfg()
	if err != nil {
		return
	}

	for _, v := range s.dbInfoMap {
		err = v.Initialize(rtdService)
		if err != nil {
			return
		}
	}

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

// from influx client
func (s *Service) queryHandle(res http.ResponseWriter, req *http.Request) {
	dbName := req.URL.Query().Get("db")
	db, ok := s.dbInfoMap[dbName]
	if ok {
		db.QueryHistory(res, req)
	} else {
		log.Printf("invalid database, name:%s", dbName)
	}
}

// from database call back
func (s *Service) notifyHandle(res http.ResponseWriter, req *http.Request) {
	_, dbName := path.Split(req.URL.Path)
	db, ok := s.dbInfoMap[dbName]

	var err error
	if ok {
		err = db.UpdateValue(res, req)
		if err == nil {
			res.WriteHeader(http.StatusOK)
			return
		}
	}

	log.Printf("invalid database, name:%s", dbName)
	res.WriteHeader(http.StatusNotFound)
}

func (s *Service) timeCheck() {
	timeOutTimer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeOutTimer.C:
			s.checkDataBase()
		}
	}
}

func (s *Service) checkDataBase() {
	bindURL := strings.Join([]string{"http://", s.bindAddress, "/ping"}, "")

	httpClient := &http.Client{}
	response, responseErr := httpClient.Get(bindURL)
	if responseErr != nil {
		return
	}
	if response.StatusCode != http.StatusNoContent {
		return
	}

	for _, v := range s.dbInfoMap {
		v.TimerCheck()
	}
}
