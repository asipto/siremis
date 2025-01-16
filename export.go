package main

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var GMFuncMap = make(map[string]any)

func GMFuncHA1(params []any) string {
	if len(params) != 3 {
		log.Printf("invalid number of parameters\n")
		return ""
	}
	text := params[0].(string) + ":" + params[1].(string) + ":" + params[2].(string)
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func GMFuncHA1B(params []any) string {
	if len(params) != 3 {
		log.Printf("invalid number of parameters\n")
		return ""
	}
	text := params[0].(string) + "@" + params[1].(string) + ":" +
		params[1].(string) + ":" + params[2].(string)
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func GMFuncDateTimeNow() string {
	return time.Now().Format(time.DateTime)
}

func GMFuncDBColumnValues(params []any) []string {
	if len(params) != 2 {
		log.Printf("invalid number of parameters\n")
		return []string{}
	}

	db := dbConn()
	selDB, err := db.Query("SELECT " + params[1].(string) + " FROM " + params[0].(string) +
		" ORDER BY " + params[1].(string) + " ASC")
	if err != nil {
		log.Printf("error [%s]\n", err.Error())
		return []string{}
	}
	defer db.Close()
	dbRes := make([]string, 0)

	for selDB.Next() {
		dbVal := ""
		err := selDB.Scan(&dbVal)
		if err != nil {
			log.Printf("error [%s]\n", err.Error())
			return []string{}
		}
		log.Println("adding option value: " + dbVal)
		dbRes = append(dbRes, dbVal)
	}

	return dbRes
}

func GMFuncParamValues(params []any) []string {
	strRes := make([]string, 0)
	for _, v := range params {
		strRes = append(strRes, v.(string))
	}
	return strRes
}

func GMTemplateFuncRowOn(nitems, idx, crt, cols, mode int) bool {
	if mode == 0 {
		return crt%cols == 0
	}
	if crt%cols == cols-1 {
		return true
	}
	if idx == nitems-1 {
		return true
	}
	return false
}

func GMTemplateFuncAdd(n, v int) int {
	return n + v
}

func GMTemplateFuncSub(n, v int) int {
	return n - v
}

func GMTemplateFuncMod(n, v int) int {
	return n % v
}

func GMTemplateFuncLoop(n int) []int {
	var i int
	var items []int
	for i = 0; i < n; i++ {
		items = append(items, i)
	}
	return items
}

func GMTemplateFuncLastLoop(idx, cols int) []int {
	var n int
	var i int
	var items []int
	if cols == 0 {
		n = 0
	} else if idx%cols == 0 {
		n = cols - 1
	} else {
		n = (idx % cols) - 1
	}
	for i = 0; i < n; i++ {
		items = append(items, i)
	}
	return items
}

func GMTemplateFuncLastIndex(nitems, idx int) bool {
	return idx == (nitems - 1)
}
