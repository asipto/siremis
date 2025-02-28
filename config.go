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
	Role     string `json:"Role,omitempty"`
}

type GMConfigAuthUsersFile struct {
	Title       string             `json:"Title,omitempty"`
	Description string             `json:"Description,omitempty"`
	AuthUsers   []GMConfigAuthUser `json:"AuthUsers"`
}

type GMConfigDB struct {
	Database    string `json:"Database"`
	Driver      string `json:"Driver"`
	Host        string `json:"Host"`
	Port        string `json:"Port"`
	Protocol    string `json:"Protocol"`
	Username    string `json:"Username"`
	Password    string `json:"Password"`
	ColumnQuote string `json:"ColumnQuote,omitempty"`
}

type GMConfigMenuItem struct {
	Name        string `json:"Name"`
	Title       string `json:"Title"`
	Inactive    bool   `json:"Inactive,omitempty"`
	MenuTopSkip bool   `json:"MenuTopSkip,omitempty"`
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

type GMConfigJRCommand struct {
	Command string `json:"Command"`
	Title   string `json:"Title"`
}

type GMConfigJRForm struct {
	Type           string              `json:"Type,omitempty"`
	CommandOptions []GMConfigJRCommand `json:"CommandOptions,omitempty"`
}

type GMConfigJR struct {
	Protocol string         `json:"Protocol"`
	LAddress string         `json:"LAddress"`
	RAddress string         `json:"RAddress"`
	ViewForm GMConfigJRForm `json:"ViewForm,omitempty"`
}

type GMConfigChartGroupsFile struct {
	Title       string         `json:"Title,omitempty"`
	Description string         `json:"Description,omitempty"`
	ChartGroups []GMChartGroup `json:"ChartGroups,omitempty"`
}

type GMConfig struct {
	DefaultViewPath     string              `json:"DefaultViewPath"`
	URLDir              string              `json:"URLDir,omitempty"`
	PublicDir           string              `json:"PublicDir,omitempty"`
	PublicDirWebPath    string              `json:"PublicDirWebPath,omitempty"`
	SchemasDir          string              `json:"SchemasDir"`
	TemplatesDir        string              `json:"TemplatesDir"`
	AuthUsersFilePath   string              `json:"AuthUsersFilePath,omitempty"`
	AuthUsers           []GMConfigAuthUser  `json:"AuthUsers,omitempty"`
	DBData              GMConfigDB          `json:"DBData"`
	JSONRPC             GMConfigJR          `json:"JSONRPC,omitempty"`
	ChartGroupsFilePath string              `json:"ChartGroupsFilePath,omitempty"`
	ChartGroups         []GMChartGroup      `json:"ChartGroups,omitempty"`
	MenuFilePath        string              `json:"MenuFilePath,omitempty"`
	Menu                []GMConfigMenuGroup `json:"Menu,omitempty"`
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
	if GMConfigV.DBData.ColumnQuote == "" {
		if GMConfigV.DBData.Driver == "mysql" {
			GMConfigV.DBData.ColumnQuote = "`"
		} else {
			GMConfigV.DBData.ColumnQuote = "\""
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

func GMConfigGetMenu(schemaName string, menuName string) *GMConfigMenuGroup {
	for i, v := range GMConfigV.Menu {
		if schemaName == v.Name || menuName == v.Name {
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

	if len(GMConfigV.TemplatesDir) == 0 {
		GMConfigV.TemplatesDir = "templates"
	}
	if len(GMConfigV.SchemasDir) == 0 {
		GMConfigV.SchemasDir = "schemas"
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

	if len(GMConfigV.AuthUsers) == 0 {
		if len(GMConfigV.AuthUsersFilePath) == 0 {
			log.Printf("no auth users in config file %s\n", GMCLIOptionsV.config)
			os.Exit(1)
		}
		configBytes, err = os.ReadFile(GMConfigV.AuthUsersFilePath)
		if err != nil {
			log.Printf("unavailable auth users file %s\n", GMConfigV.AuthUsersFilePath)
			os.Exit(1)
		}
		var auFile = GMConfigAuthUsersFile{}
		err = json.Unmarshal(configBytes, &auFile)
		if err != nil {
			log.Printf("invalid content in auth users file %s\n", GMConfigV.AuthUsersFilePath)
			os.Exit(1)
		}
		if len(auFile.AuthUsers) == 0 {
			log.Printf("no auth users in file %s\n", GMConfigV.AuthUsersFilePath)
			os.Exit(1)
		}
		GMConfigV.AuthUsers = auFile.AuthUsers
	}

	if len(GMConfigV.ChartGroupsFilePath) > 0 {
		configBytes, err = os.ReadFile(GMConfigV.ChartGroupsFilePath)
		if err != nil {
			log.Printf("unavailable chart groups file %s\n", GMConfigV.ChartGroupsFilePath)
			os.Exit(1)
		}
		var cgFile = GMConfigChartGroupsFile{}
		err = json.Unmarshal(configBytes, &cgFile)
		if err != nil {
			log.Printf("invalid content in chart groups file %s\n", GMConfigV.ChartGroupsFilePath)
			os.Exit(1)
		}
		if len(cgFile.ChartGroups) == 0 {
			log.Printf("no chart groups in file %s\n", GMConfigV.ChartGroupsFilePath)
			os.Exit(1)
		}
		GMConfigV.ChartGroups = cgFile.ChartGroups
	}

	GMConfigEvalVals()
}
