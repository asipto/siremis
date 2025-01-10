package main

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
	SchemaDir       string              `json:"SchemaDir"`
	AuthUsers       []GMConfigAuthUser  `json:"AuthUsers"`
	DBData          GMConfigDB          `json:"DBData"`
	Menu            []GMConfigMenuGroup `json:"Menu"`
}
