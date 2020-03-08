package database

import (
	"supos.ai/data-lake/external/tsdb-proxy/collector"
	pb "supos.ai/data-lake/external/tsdb-proxy/common/model"
	"supos.ai/data-lake/external/tsdb-proxy/model"
)

type piImpl struct {
	info      *model.DBInfo
	collector collector.Collector
}

// NewPi new pi DB
func NewPi(info *model.DBInfo) DB {
	return &piImpl{info: info}
}

func (s *piImpl) Initialize(rtdService string) (err error) {
	s.collector = collector.NewCollector(s.info.Name)

	err = s.collector.Start(rtdService, s.OnStatusCallBack)

	return
}

func (s *piImpl) Uninitialize() {
	s.collector.Stop()

}

func (s *piImpl) QueryHistory() (err error) {
	return
}

func (s *piImpl) UpdateValue(values *pb.ValueSequnce) (err error) {
	err = s.collector.UpdateValue(values)
	return
}

func (s *piImpl) OnStatusCallBack(collectName string, status, errorCode int, reason string) {
	if status == collector.LoginStatus {
		if s.collector.IsReady() {
			values := &pb.ValueSequnce{}
			s.collector.UpdateValue(values)
		}
	}
}
