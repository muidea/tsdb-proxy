package database

import "net/http"

const (
	// Init initialize
	Init = iota
	// LoginOK login ok
	LoginOK
	//EnumTags enum tags
	EnumTags
	// Subscribed subscribe
	Subscribed
)

// DB database interface
type DB interface {
	Initialize(rtdService string) (err error)
	Uninitialize()

	QueryHistory(res http.ResponseWriter, req *http.Request) (err error)
	UpdateValue(res http.ResponseWriter, req *http.Request) (err error)
	TimerCheck()
}
