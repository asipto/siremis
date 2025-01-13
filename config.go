package main

import (
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

type GMConfig struct {
	DefaultViewPath string              `json:"DefaultViewPath"`
	URLDir          string              `json:"URLDir,omitempty"`
	PublicDir       string              `json:"PublicDir,omitempty"`
	SchemaDir       string              `json:"SchemaDir"`
	AuthUsers       []GMConfigAuthUser  `json:"AuthUsers"`
	DBData          GMConfigDB          `json:"DBData"`
	Menu            []GMConfigMenuGroup `json:"Menu"`
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
