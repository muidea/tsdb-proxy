package model

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// DBInfo database info
type DBInfo struct {
	Name     string `json:"name"`
	Desc     string `json:"description"`
	Address  string `json:"address"`
	Account  string `json:"account"`
	Password string `json:"password"`
}

// ConfigInfo config info
type ConfigInfo struct {
	StdService []*DBInfo `json:"stdService"`
	PiService  []*DBInfo `json:"piService"`
}

// Load load config
func (s *ConfigInfo) Load(cfgFile string) (err error) {
	fileObj, fileErr := os.Open(cfgFile)
	if fileErr != nil {
		err = fileErr
		return
	}

	defer fileObj.Close()

	content, contentErr := ioutil.ReadAll(fileObj)
	if contentErr != nil {
		err = contentErr
		return
	}

	err = json.Unmarshal(content, s)

	return
}
