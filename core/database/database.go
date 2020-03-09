package database

import "net/http"

// DB database interface
type DB interface {
	Initialize(rtdService string) (err error)
	Uninitialize()

	QueryHistory(res http.ResponseWriter, req *http.Request) (err error)
	UpdateValue(res http.ResponseWriter, req *http.Request) (err error)
	CheckHealth()
}
