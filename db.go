package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func dbConn() (db *sql.DB) {
	log.Println("Database: " + GMConfigV.DBData.Database + " (" + GMConfigV.DBData.Host +
		":" + GMConfigV.DBData.Port + ")")
	db, err := sql.Open(GMConfigV.DBData.Driver, GMConfigV.DBData.Username+":"+
		GMConfigV.DBData.Password+"@"+GMConfigV.DBData.Protocol+
		"("+GMConfigV.DBData.Host+":"+GMConfigV.DBData.Port+")/"+
		GMConfigV.DBData.Database)
	if err != nil {
		panic(err.Error())
	}
	return db
}

func dbColumnQuoted(cName string) string {
	return GMConfigV.DBData.ColumnQuote + cName + GMConfigV.DBData.ColumnQuote
}
