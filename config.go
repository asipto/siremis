package main

type GMConfigAccess struct {
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

type GMConfig struct {
	DefaultSchema string           `json:"DefaultSchema"`
	DefaultAction string           `json:"DefaultAction"`
	URLDir        string           `json:"URLDir,omitempty"`
	SchemaDir     string           `json:"SchemaDir"`
	Access        []GMConfigAccess `json:"Access"`
	DBData        GMConfigDB       `json:"DBData"`
}
