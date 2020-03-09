package std

import (
	"net/http"

	"supos.ai/data-lake/external/tsdb-proxy/collector"
	pb "supos.ai/data-lake/external/tsdb-proxy/common/model"
	"supos.ai/data-lake/external/tsdb-proxy/core/database"
	"supos.ai/data-lake/external/tsdb-proxy/model"
)

type stdImpl struct {
	info              *model.DBInfo
	subscribeCallBack string

	collector collector.Collector
}

// NewStd new std DB
func NewStd(info *model.DBInfo, callBack string) database.DB {
	return &stdImpl{info: info, subscribeCallBack: callBack}
}

func (s *stdImpl) Initialize(rtdService string) (err error) {
	s.collector = collector.NewCollector(s.info.Name)

	err = s.collector.Start(rtdService, s.OnStatusCallBack)

	return
}

func (s *stdImpl) Uninitialize() {
	s.collector.Stop()

}

func (s *stdImpl) QueryHistory(res http.ResponseWriter, req *http.Request) (err error) {
	return
}

func (s *stdImpl) UpdateValue(res http.ResponseWriter, req *http.Request) (err error) {
	values := &pb.ValueSequnce{}
	err = s.collector.UpdateValue(values)
	return
}

func (s *stdImpl) TimerCheck() {

}

func (s *stdImpl) OnStatusCallBack(collectName string, status, errorCode int, reason string) {
	if status == collector.LoginStatus {
		if s.collector.IsReady() {
			values := &pb.ValueSequnce{}
			s.collector.UpdateValue(values)
		}
	}
}
