package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

type GMConfigAuthUser struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
	Role     string `json:"Role"`
}

type GMConfigDB struct {
	Database string `json:"Database"`
	Driver   string `json:"Driver"`
	Host     string `json:"Host"`
	Port     string `json:"Port"`
	Protocol string `json:"Protocol"`
	Username string `json:"Username"`
	Password string `json:"Password"`
}

type GMConfigMenuItem struct {
	Name        string `json:"Name"`
	Title       string `json:"Title"`
	Inactive    bool   `json:"Inactive,omitempty"`
	URLPath     string `json:"URLPath"`
	Description string `json:"Description,omitempty"`
}

type GMConfigMenuGroup struct {
	Name     string             `json:"Name"`
	Title    string             `json:"Title"`
	Inactive bool               `json:"Inactive,omitempty"`
	Items    []GMConfigMenuItem `json:"Items"`
}

type GMConfigMenuFile struct {
	Title       string              `json:"Title,omitempty"`
	Description string              `json:"Description,omitempty"`
	Menu        []GMConfigMenuGroup `json:"Menu"`
}

type GMConfig struct {
	DefaultViewPath string              `json:"DefaultViewPath"`
	URLDir          string              `json:"URLDir,omitempty"`
	PublicDir       string              `json:"PublicDir,omitempty"`
	SchemaDir       string              `json:"SchemaDir"`
	AuthUsers       []GMConfigAuthUser  `json:"AuthUsers"`
	DBData          GMConfigDB          `json:"DBData"`
	MenuFilePath    string              `json:"MenuFilePath,omitempty"`
	Menu            []GMConfigMenuGroup `json:"Menu,omitempty"`
}

var GMConfigV = GMConfig{}

func GMConfigEvalVals() {
	if strings.HasPrefix(GMConfigV.DBData.Database, "@env:") {
		eVal, ok := os.LookupEnv(GMConfigV.DBData.Database[5:])
		if ok {
			GMConfigV.DBData.Database = eVal
		}
	}
	if strings.HasPrefix(GMConfigV.DBData.Driver, "@env:") {
		eVal, ok := os.LookupEnv(GMConfigV.DBData.Driver[5:])
		if ok {
			GMConfigV.DBData.Driver = eVal
		}
	}
	if strings.HasPrefix(GMConfigV.DBData.Host, "@env:") {
		eVal, ok := os.LookupEnv(GMConfigV.DBData.Host[5:])
		if ok {
			GMConfigV.DBData.Host = eVal
		}
	}
	if strings.HasPrefix(GMConfigV.DBData.Port, "@env:") {
		eVal, ok := os.LookupEnv(GMConfigV.DBData.Port[5:])
		if ok {
			GMConfigV.DBData.Port = eVal
		}
	}
	if strings.HasPrefix(GMConfigV.DBData.Protocol, "@env:") {
		eVal, ok := os.LookupEnv(GMConfigV.DBData.Protocol[5:])
		if ok {
			GMConfigV.DBData.Protocol = eVal
		}
	}
	if strings.HasPrefix(GMConfigV.DBData.Username, "@env:") {
		eVal, ok := os.LookupEnv(GMConfigV.DBData.Username[5:])
		if ok {
			GMConfigV.DBData.Username = eVal
		}
	}
	if strings.HasPrefix(GMConfigV.DBData.Password, "@env:") {
		eVal, ok := os.LookupEnv(GMConfigV.DBData.Password[5:])
		if ok {
			GMConfigV.DBData.Password = eVal
		}
	}
	for i, v := range GMConfigV.AuthUsers {
		if strings.HasPrefix(v.Username, "@env:") {
			eVal, ok := os.LookupEnv(v.Username[5:])
			if ok {
				GMConfigV.AuthUsers[i].Username = eVal
			}
		}
		if strings.HasPrefix(v.Password, "@env:") {
			eVal, ok := os.LookupEnv(v.Password[5:])
			if ok {
				GMConfigV.AuthUsers[i].Password = eVal
			}
		}
	}
}

func GMConfigGetSchemaMenu(schemaName string) *GMConfigMenuGroup {
	for i, v := range GMConfigV.Menu {
		if schemaName == v.Name {
			return &GMConfigV.Menu[i]
		}
	}
	return nil
}

func GMConfigLoad() {
	configBytes, err := os.ReadFile(GMCLIOptionsV.config)
	if err != nil {
		log.Printf("unavailable config file %s\n", GMCLIOptionsV.config)
		os.Exit(1)
	}
	err = json.Unmarshal(configBytes, &GMConfigV)
	if err != nil {
		log.Printf("invalid content in config file %s\n", GMCLIOptionsV.config)
		os.Exit(1)
	}

	if len(GMConfigV.Menu) == 0 {
		if len(GMConfigV.MenuFilePath) == 0 {
			log.Printf("no menu in config file %s\n", GMCLIOptionsV.config)
			os.Exit(1)
		}
		configBytes, err = os.ReadFile(GMConfigV.MenuFilePath)
		if err != nil {
			log.Printf("unavailable menu file %s\n", GMConfigV.MenuFilePath)
			os.Exit(1)
		}
		var menuFile = GMConfigMenuFile{}
		err = json.Unmarshal(configBytes, &menuFile)
		if err != nil {
			log.Printf("invalid content in menu file %s\n", GMConfigV.MenuFilePath)
			os.Exit(1)
		}
		if len(menuFile.Menu) == 0 {
			log.Printf("no menu in file %s\n", GMConfigV.MenuFilePath)
			os.Exit(1)
		}
		GMConfigV.Menu = menuFile.Menu
	}
	GMConfigEvalVals()
}
