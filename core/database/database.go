package database

// DB database interface
type DB interface {
	LoadConfig() (err error)
	QueryHistory() (err error)
	UpdateValue() (err error)
}
