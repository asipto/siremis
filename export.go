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

func GMParamString(params []any, idx int) (string, bool) {
	if idx < 0 || idx >= len(params) {
		log.Printf("invalid parameter index %d\n", idx)
		return "", false
	}
	val, ok := params[idx].(string)
	if !ok {
		log.Printf("invalid parameter type at index %d\n", idx)
		return "", false
	}
	return val, true
}

func GMParamFloat64(params []any, idx int) (float64, bool) {
	if idx < 0 || idx >= len(params) {
		log.Printf("invalid parameter index %d\n", idx)
		return 0, false
	}
	val, ok := params[idx].(float64)
	if !ok {
		log.Printf("invalid parameter type at index %d\n", idx)
		return 0, false
	}
	return val, true
}

func GMParamInt64(params []any, idx int) (int64, bool) {
	if idx < 0 || idx >= len(params) {
		log.Printf("invalid parameter index %d\n", idx)
		return 0, false
	}
	val, ok := params[idx].(int64)
	if !ok {
		log.Printf("invalid parameter type at index %d\n", idx)
		return 0, false
	}
	return val, true
}

func GMFuncHA1(params []any) string {
	if len(params) != 3 {
		log.Printf("invalid number of parameters\n")
		return ""
	}
	p0, ok := GMParamString(params, 0)
	if !ok {
		return ""
	}
	p1, ok := GMParamString(params, 1)
	if !ok {
		return ""
	}
	p2, ok := GMParamString(params, 2)
	if !ok {
		return ""
	}
	text := p0 + ":" + p1 + ":" + p2
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func GMFuncHA1B(params []any) string {
	if len(params) != 3 {
		log.Printf("invalid number of parameters\n")
		return ""
	}
	p0, ok := GMParamString(params, 0)
	if !ok {
		return ""
	}
	p1, ok := GMParamString(params, 1)
	if !ok {
		return ""
	}
	p2, ok := GMParamString(params, 2)
	if !ok {
		return ""
	}
	text := p0 + "@" + p1 + ":" + p1 + ":" + p2
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
	defer db.Close()
	tableName, ok := GMParamString(params, 0)
	if !ok {
		return []GMOptionValue{}
	}
	columnName, ok := GMParamString(params, 1)
	if !ok {
		return []GMOptionValue{}
	}
	selDB, err := db.Query("SELECT " + columnName + " FROM " + tableName +
		" ORDER BY " + columnName + " ASC")
	if err != nil {
		log.Printf("error [%s]\n", err.Error())
		return []GMOptionValue{}
	}
	defer selDB.Close()
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
	if err := selDB.Err(); err != nil {
		log.Printf("error [%s]\n", err.Error())
		return []GMOptionValue{}
	}

	return dbRes
}

func GMFuncParamValues(params []any) []GMOptionValue {
	lRes := make([]GMOptionValue, 0)
	for i := range params {
		value, ok := GMParamString(params, i)
		if !ok {
			return []GMOptionValue{}
		}
		var oVal = GMOptionValue{}
		oVal.Value = value
		oVal.Title = oVal.Value
		lRes = append(lRes, oVal)
	}
	return lRes
}

func GMFuncParamVN(params []any) []GMOptionValue {
	lRes := make([]GMOptionValue, 0)
	for i := 0; i+1 < len(params); i += 2 {
		value, ok := GMParamString(params, i)
		if !ok {
			return []GMOptionValue{}
		}
		title, ok := GMParamString(params, i+1)
		if !ok {
			return []GMOptionValue{}
		}
		oVal := GMOptionValue{
			Value: value,
			Title: title,
		}
		lRes = append(lRes, oVal)
	}
	return lRes
}

func GMFuncFloat2D(params []any) string {
	val, ok := GMParamFloat64(params, 0)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%.2f", val)
}

func GMFuncTimeStampUTCDate(params []any) string {
	val, ok := GMParamInt64(params, 0)
	if !ok {
		return ""
	}
	tv := time.Unix(val, 0)
	return tv.Format(time.RFC3339)
}

func GMFuncListBitFlags(params []any) string {
	fv, ok := GMParamInt64(params, 0)
	if !ok {
		return ""
	}
	if fv == 0 {
		return "0"
	}
	sv := fmt.Sprintf("%d", fv)
	if len(params) == 1 {
		return sv
	}
	sv += " [ "
	for i := 1; i < len(params); i++ {
		sPm, ok := GMParamString(params, i)
		if !ok {
			return ""
		}
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
