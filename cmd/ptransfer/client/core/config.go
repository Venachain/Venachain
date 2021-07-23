package core

import (
	"encoding/json"
	"os"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/ptransfer/client/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
)

// Config store the values from config.json file
type Config struct {
	From     string `json:"from"`     // the address used to send the transaction
	Contract string `json:"contract"` // the address of the privacy token
	Url      string `json:"url"`      // the ip address of the remote node
	Verbosity      int `json:"verbosity"`      // the log level
}

var config = &Config{}

const (
	defaultConfigFilePath = "./config.json"
)

// configInit read values from config file
func configInit() {
	runPath := utl.GetRunningTimePath()
	configFile := runPath + defaultConfigFilePath

	_, err := os.Stat(configFile)
	if !os.IsNotExist(err) {
		config = ParseConfigJson(configFile)
	}
}

// WriteConfigFile writes data into config.json
func WriteConfig(filePath string, config *Config) {
	// Open or create file
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		utils.Fatalf(utl.ErrOpenFileFormat, "config", err.Error())
	}
	defer file.Close()

	fileBytes, _ := json.Marshal(config)

	// write file
	_ = file.Truncate(0)
	_, err = file.Write(fileBytes)
	if err != nil {
		utils.Fatalf(utl.ErrWriteFileFormat, err.Error())
	}
}

// ParseConfigJson parses the data in config.json to Config object
func ParseConfigJson(configPath string) *Config {

	var config = &Config{}

	configBytes, err := utl.ParseFileToBytes(configPath)
	if err != nil {
		utils.Fatalf(utl.ErrParseFileFormat, configPath, err.Error())
	}

	if len(configBytes) == 0 {
		return &Config{}
	}

	err = json.Unmarshal(configBytes, config)
	if err != nil {
		// utils.Fatalf(utl.ErrUnmarshalBytesFormat, configPath, err.Error())
		return &Config{}
	}

	return config
}
