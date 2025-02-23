package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type GMOptionValue struct {
	Title string
	Value string
}

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

func GMFuncDBColumnValues(params []any) []GMOptionValue {
	if len(params) != 2 {
		log.Printf("invalid number of parameters\n")
		return []GMOptionValue{}
	}

	db := dbConn()
	selDB, err := db.Query("SELECT " + params[1].(string) + " FROM " + params[0].(string) +
		" ORDER BY " + params[1].(string) + " ASC")
	if err != nil {
		log.Printf("error [%s]\n", err.Error())
		return []GMOptionValue{}
	}
	defer db.Close()
	dbRes := make([]GMOptionValue, 0)

	for selDB.Next() {
		var oVal = GMOptionValue{}
		err := selDB.Scan(&oVal.Value)
		if err != nil {
			log.Printf("error [%s]\n", err.Error())
			return []GMOptionValue{}
		}
		oVal.Title = oVal.Value
		log.Println("adding option value: " + oVal.Value)
		dbRes = append(dbRes, oVal)
	}

	return dbRes
}

func GMFuncParamValues(params []any) []GMOptionValue {
	lRes := make([]GMOptionValue, 0)
	for _, v := range params {
		var oVal = GMOptionValue{}
		oVal.Value = v.(string)
		oVal.Title = oVal.Value
		lRes = append(lRes, oVal)
	}
	return lRes
}

func GMFuncParamVN(params []any) []GMOptionValue {
	lRes := make([]GMOptionValue, 0)
	for i, v := range params {
		var oVal = GMOptionValue{}
		if i%2 == 1 {
			oVal.Title = v.(string)
			lRes = append(lRes, oVal)
		} else {
			oVal.Value = v.(string)
		}
	}
	return lRes
}

func GMFuncFloat2D(params []any) string {
	return fmt.Sprintf("%.2f", params[0].(float64))
}

func GMFuncTimeStampUTCDate(params []any) string {
	tv := time.Unix(params[0].(int64), 0)
	return tv.Format(time.RFC3339)
}

func GMFuncListBitFlags(params []any) string {
	fv := params[0].(int64)
	if fv == 0 {
		return "0"
	}
	sv := fmt.Sprintf("%d", fv)
	if len(params) == 1 {
		return sv
	}
	sv += " [ "
	for i := 1; i < len(params); i++ {
		sPm := params[i].(string)
		vFN := strings.Split(sPm, ":")
		if len(vFN) == 2 {
			fl, _ := strconv.Atoi(vFN[0])
			if (fv & (1 << fl)) != 0 {
				sv += sPm + " "
			}
		}
	}
	sv += "]"
	return sv
}

func GMTemplateFuncRowOn(nitems, idx, crt, cols, mode int) bool {
	if mode == 0 {
		return (crt-1)%cols == 0
	}
	if (crt-1)%cols == cols-1 {
		return true
	}
	if idx == nitems-1 {
		return true
	}
	return false
}

func GMTemplateFuncRowStart(crt, cols int) bool {
	return crt%cols == 1
}

func GMTemplateFuncRowEnd(crt, cols int) bool {
	return crt%cols == 0
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

func GMTemplateFuncModX(n, v int) bool {
	return (n % v) > 0
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

	if cols-1 == 0 {
		n = 0
	} else if idx%cols == 0 {
		n = 0
	} else {
		n = cols - (idx % cols)
	}
	for i = 0; i < n; i++ {
		items = append(items, i)
	}
	return items
}

func GMTemplateFuncLastIndex(nitems, idx int) bool {
	return idx == (nitems - 1)
}
