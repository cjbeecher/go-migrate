package gomigrate

import (
	"encoding/json"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"strings"
)

// Parameters stores the necessary parameters for connecting to a DB
type Parameters struct {
	DBType   string `json:"dbType"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Queries  string `json:"queries"`
}

// Migration stores the parameters on the current migrator execution
type Migration struct {
	DBParams map[string]*Parameters `json:"dbParams"`
}

// Init initializes the Migration object
func (m *Migration) Init(fNames ...string) error {
	fName := ""
	user, _ := user.Current()
	homeDir := user.HomeDir
	if len(fNames) != 1 {
		fName = filepath.Join(homeDir, ".go-migrate", "config.json")
	} else {
		fName = fNames[0]
	}
	fName = strings.Replace(fName, "~", homeDir, 0)
	data, err := ioutil.ReadFile(fName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, m)
	return err
}
