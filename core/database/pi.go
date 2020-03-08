package database

import (
	"supos.ai/data-lake/external/tsdb-proxy/model"
)

type piImpl struct {
	info *model.DBInfo
}

// NewPi new pi DB
func NewPi(info *model.DBInfo) DB {
	return &piImpl{info: info}
}

func (s *piImpl) LoadConfig() (err error) {
	return
}

func (s *piImpl) QueryHistory() (err error) {
	return
}

func (s *piImpl) UpdateValue() (err error) {
	return
}
