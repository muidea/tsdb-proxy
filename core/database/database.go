package database

import "supos.ai/data-lake/external/tsdb-proxy/common/model"

// DB database interface
type DB interface {
	Initialize(rtdService string) (err error)
	Uninitialize()
	QueryHistory() (err error)
	UpdateValue(values *model.ValueSequnce) (err error)
}
