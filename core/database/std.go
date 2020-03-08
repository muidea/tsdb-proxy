package database

import (
	"supos.ai/data-lake/external/tsdb-proxy/model"
)

type stdImpl struct {
	info *model.DBInfo
}

// NewStd new std DB
func NewStd(info *model.DBInfo) DB {
	return &stdImpl{info: info}
}

func (s *stdImpl) LoadConfig() (err error) {
	return
}

func (s *stdImpl) QueryHistory() (err error) {
	return
}

func (s *stdImpl) UpdateValue() (err error) {
	return
}
