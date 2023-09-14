package userconfmanager

import (
	utils "eboox/Utils"
	"encoding/json"
	"fmt"
	"os"
)

const (
	CONF_PATH         string = "/eboox/conf.json"
	DEFAULT_BOOKS_DIR string = "/Documents/Library/"
)

type UserConfManager struct {
	LibraryDir      string
	UserActivityDir string

	Theme string `default:"light"`
}

func Default() UserConfManager {
	homeDir, _ := os.UserHomeDir()
	return UserConfManager{
		LibraryDir: homeDir + DEFAULT_BOOKS_DIR,
		// Implement handler
		UserActivityDir: homeDir + DEFAULT_BOOKS_DIR + "Activity/",
	}
}

func UserConfManagerInit() *UserConfManager {
	if path, exists := utils.ConfExists(CONF_PATH); exists {
		//load
		cfgData, err := os.ReadFile(path + "/conf.json")
		utils.Assert(err == nil, "Assersion failure : Cannot load data from confFile.")
		userConf := UserConfManager{}
		json.Unmarshal(cfgData, &userConf)
		println("loaded config successfully from [%s]\n", path)
		return &userConf
	} else {
		// make and save
		utils.MakeConfig(path)
		defaultConf := Default()
		cfgData, _ := json.MarshalIndent(defaultConf, "", "  ")
		err := os.WriteFile(path+"/conf.json", cfgData, 0644)
		utils.Assert(err == nil, "Assersion failure : Cannot write conf to Conffile.")
		fmt.Printf("Initialized Config at [%s]\n", path)
		return &defaultConf
	}
}
