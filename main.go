package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

const siregisVersion = "1.00"

// CLIOptions - structure for command line options
type GMCLIOptions struct {
	config      string
	domain      string
	httpsrv     string
	httpssrv    string
	httpsusele  bool
	httpspubkey string
	httpsprvkey string
	version     bool
}

var GMCLIOptionsV = GMCLIOptions{
	config:      "etc/config.json",
	domain:      "",
	httpsrv:     "127.0.0.1:8040",
	httpssrv:    "",
	httpsusele:  false,
	httpspubkey: "",
	httpsprvkey: "",
	version:     false,
}

// SIPUser struct
type SIPUser struct {
	Id       int
	Username string
	Domain   string
	Password string
	Ha1      string
	Ha1b     string
}

type GMViewData struct {
	Config GMConfig
	Schema GMSchema
	Fields []GMSchemaField
	Values []any
}

type GMDBField struct {
	Name   string
	Column string
	Value  any
}

var GMConfigV = GMConfig{}

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

func dbConn() (db *sql.DB) {
	log.Println("Database: " + GMConfigV.DBData.Host + " (" + GMConfigV.DBData.Host +
		":" + GMConfigV.DBData.Port + ")")
	log.Println("Database host: " + GMConfigV.DBData.Host)
	log.Println("Database port: " + GMConfigV.DBData.Port)
	db, err := sql.Open(GMConfigV.DBData.Driver, GMConfigV.DBData.Username+":"+
		GMConfigV.DBData.Password+"@"+GMConfigV.DBData.Protocol+
		"("+GMConfigV.DBData.Host+":"+GMConfigV.DBData.Port+")/"+
		GMConfigV.DBData.Database)
	if err != nil {
		panic(err.Error())
	}
	return db
}

var tmpl = template.Must(template.ParseGlob("templates/*"))

func GMList(w http.ResponseWriter, r *http.Request, schemaName string) {
	schemaFile := GMConfigV.SchemaDir + "/" + schemaName + ".json"
	schemaBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Printf("unavailable schema file %s\n", schemaFile)
		os.Exit(1)
	}
	var schemaV = GMSchema{}
	err = json.Unmarshal(schemaBytes, &schemaV)
	if err != nil {
		log.Printf("invalid content in schema file %s\n", schemaFile)
		os.Exit(1)
	}
	var selFields = []GMSchemaField{}
	for _, v := range schemaV.Fields {
		if v.Name == schemaV.Query.IdField {
			log.Printf("using field %s\n", v.Name)
			selFields = append(selFields, v)
			break
		}
	}
	for _, v := range schemaV.Fields {
		if v.Name != schemaV.Query.IdField {
			if v.Display.List {
				log.Printf("using field %s\n", v.Name)
				selFields = append(selFields, v)
			} else {
				log.Printf("skipping field %s (%v)\n", v.Name, v.Display.List)
			}
		}
	}
	strQuery := "SELECT "
	for i, v := range selFields {
		if i == 0 {
			strQuery += v.Column
		} else {
			strQuery += ", " + v.Column
		}
	}
	strQuery += " FROM " + schemaV.Table + " ORDER BY " + selFields[0].Column + " DESC"
	db := dbConn()
	selDB, err := db.Query(strQuery)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	dbRes := make([]any, 0)

	for selDB.Next() {
		dbRow := make([]any, len(selFields))
		for i, v := range selFields {
			if v.Type == "int" {
				dbRow[i] = new(int)
			} else if v.Type == "str" || v.Type == "string" {
				dbRow[i] = new(string)
			} else {
				dbRow[i] = new(string)
			}
		}
		err := selDB.Scan(dbRow...)
		if err != nil {
			panic(err.Error())
		}
		log.Println("listing row: id: " + strconv.Itoa(*dbRow[0].(*int)))

		dbRes = append(dbRes, dbRow)
	}

	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Schema = schemaV
	viewData.Fields = selFields
	viewData.Values = dbRes
	tmpl.ExecuteTemplate(w, "list", viewData)
}

func GMSchemaFieldDisplayBoolVal(v *GMSchemaFieldDisplay, field string) bool {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	return bool(f.Bool())
}

func GMView(w http.ResponseWriter, r *http.Request, schemaName string, sId string,
	sDisplayField string, sTemplate string) {
	schemaFile := GMConfigV.SchemaDir + "/" + schemaName + ".json"
	schemaBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Printf("unavailable schema file %s\n", schemaFile)
		os.Exit(1)
	}
	var schemaV = GMSchema{}
	err = json.Unmarshal(schemaBytes, &schemaV)
	if err != nil {
		log.Printf("invalid content in schema file %s\n", schemaFile)
		os.Exit(1)
	}
	var selFields = []GMSchemaField{}
	for _, v := range schemaV.Fields {
		if v.Name == schemaV.Query.IdField {
			log.Printf("using field %s\n", v.Name)
			selFields = append(selFields, v)
			break
		}
	}
	for _, v := range schemaV.Fields {
		if v.Name != schemaV.Query.IdField {
			if GMSchemaFieldDisplayBoolVal(&v.Display, sDisplayField) {
				log.Printf("using field %s\n", v.Name)
				selFields = append(selFields, v)
			} else {
				log.Printf("skipping field %s (%v)\n", v.Name, v.Display.List)
			}
		}
	}
	strQuery := "SELECT "
	for i, v := range selFields {
		if i == 0 {
			strQuery += v.Column
		} else {
			strQuery += ", " + v.Column
		}
	}
	strQuery += " FROM " + schemaV.Table + " WHERE " + selFields[0].Column + " = ?"
	db := dbConn()
	defer db.Close()
	var vId any
	if selFields[0].Type == "int" {
		nId := 0
		nId, _ = strconv.Atoi(sId)
		vId = nId
	} else {
		vId = sId
	}
	selDB, err := db.Query(strQuery, vId)
	if err != nil {
		panic(err.Error())
	}
	dbRes := make([]any, 0)

	for selDB.Next() {
		dbRow := make([]any, len(selFields))
		for i, v := range selFields {
			if v.Type == "int" {
				dbRow[i] = new(int)
			} else if v.Type == "str" || v.Type == "string" {
				dbRow[i] = new(string)
			} else {
				dbRow[i] = new(string)
			}
		}
		err := selDB.Scan(dbRow...)
		if err != nil {
			panic(err.Error())
		}
		log.Println("listing row: id: " + strconv.Itoa(*dbRow[0].(*int)))

		dbRes = append(dbRes, dbRow)
	}

	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Schema = schemaV
	viewData.Fields = selFields
	viewData.Values = dbRes
	tmpl.ExecuteTemplate(w, sTemplate, viewData)
}

func GMShow(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	GMView(w, r, schemaName, sId, "Show", "show")
}

func GMNew(w http.ResponseWriter, r *http.Request, schemaName string) {
	schemaFile := GMConfigV.SchemaDir + "/" + schemaName + ".json"
	schemaBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Printf("unavailable schema file %s\n", schemaFile)
		os.Exit(1)
	}
	var schemaV = GMSchema{}
	err = json.Unmarshal(schemaBytes, &schemaV)
	if err != nil {
		log.Printf("invalid content in schema file %s\n", schemaFile)
		os.Exit(1)
	}
	var selFields = []GMSchemaField{}
	for _, v := range schemaV.Fields {
		if v.Name != schemaV.Query.IdField {
			if v.Display.Insert {
				log.Printf("using field %s\n", v.Name)
				selFields = append(selFields, v)
			} else {
				log.Printf("skipping field %s (%v)\n", v.Name, v.Display.List)
			}
		}
	}

	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Schema = schemaV
	viewData.Fields = selFields
	tmpl.ExecuteTemplate(w, "new", viewData)
}

func GMInsert(w http.ResponseWriter, r *http.Request, schemaName string) {
	schemaFile := GMConfigV.SchemaDir + "/" + schemaName + ".json"
	schemaBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Printf("unavailable schema file %s\n", schemaFile)
		os.Exit(1)
	}
	var schemaV = GMSchema{}
	err = json.Unmarshal(schemaBytes, &schemaV)
	if err != nil {
		log.Printf("invalid content in schema file %s\n", schemaFile)
		os.Exit(1)
	}
	var valFields = []GMDBField{}
	for _, v := range schemaV.Fields {
		if v.Name != schemaV.Query.IdField {
			var vField = GMDBField{}
			vField.Name = v.Name
			vField.Column = v.Column
			if v.Display.Insert {
				log.Printf("insert form field %s\n", v.Name)
				if v.Type == "int" {
					vField.Value, _ = strconv.Atoi(r.FormValue(v.Name))
				} else {
					vField.Value = r.FormValue(v.Name)
				}
				valFields = append(valFields, vField)
			} else if len(v.ValueInsert.Func) > 0 {
				log.Printf("insert value field %s\n", v.Name)
				sVal := ""
				if len(v.ValueInsert.Params) > 0 {
					var vParams = make([]any, 0)
					for _, p := range v.ValueInsert.Params {
						for _, f := range valFields {
							if p == f.Name {
								vParams = append(vParams, f.Value)
								break
							}
						}
					}
					sVal = GMFuncMap[v.ValueInsert.Func].(func([]any) string)(vParams)
				} else {
					sVal = GMFuncMap[v.ValueInsert.Func].(func() string)()
				}
				if v.Type == "int" {
					vField.Value, _ = strconv.Atoi(sVal)
				} else {
					vField.Value = sVal
				}
				valFields = append(valFields, vField)
			} else {
				log.Printf("skipping field %s\n", v.Name)
			}
		}
	}

	db := dbConn()
	defer db.Close()
	strQCols := ""
	strQValQ := ""
	var dbVals = make([]any, 0)
	for i, v := range valFields {
		if i != 0 {
			strQCols += ", "
			strQValQ += ", "
		}
		strQCols += v.Column
		strQValQ += "?"
		dbVals = append(dbVals, v.Value)
	}
	insForm, err := db.Prepare("INSERT INTO " + schemaV.Table + " (" + strQCols + ") VALUES (" + strQValQ + ")")
	if err != nil {
		panic(err.Error())
	}
	insForm.Exec(dbVals...)
	http.Redirect(w, r, GMConfigV.URLDir+"/"+GMConfigV.DefaultAction+"/"+schemaName, http.StatusMovedPermanently)
}

func GMEdit(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	GMView(w, r, schemaName, sId, "Edit", "edit")
}

func GMUpdate(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	schemaFile := GMConfigV.SchemaDir + "/" + schemaName + ".json"
	schemaBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Printf("unavailable schema file %s\n", schemaFile)
		os.Exit(1)
	}
	var schemaV = GMSchema{}
	err = json.Unmarshal(schemaBytes, &schemaV)
	if err != nil {
		log.Printf("invalid content in schema file %s\n", schemaFile)
		os.Exit(1)
	}
	var valFields = []GMDBField{}
	var idField = GMSchemaField{}
	for _, v := range schemaV.Fields {
		if v.Name == schemaV.Query.IdField {
			idField = v
		} else {
			var vField = GMDBField{}
			vField.Name = v.Name
			vField.Column = v.Column
			if v.Display.Edit {
				log.Printf("update form field %s\n", v.Name)
				if v.Type == "int" {
					vField.Value, _ = strconv.Atoi(r.FormValue(v.Name))
				} else {
					vField.Value = r.FormValue(v.Name)
				}
				valFields = append(valFields, vField)
			} else if len(v.ValueEdit.Func) > 0 {
				log.Printf("update value field %s\n", v.Name)
				sVal := ""
				if len(v.ValueEdit.Params) > 0 {
					var vParams = make([]any, 0)
					for _, p := range v.ValueEdit.Params {
						for _, f := range valFields {
							if p == f.Name {
								vParams = append(vParams, f.Value)
								break
							}
						}
					}
					sVal = GMFuncMap[v.ValueEdit.Func].(func([]any) string)(vParams)
				} else {
					sVal = GMFuncMap[v.ValueEdit.Func].(func() string)()
				}
				if v.Type == "int" {
					vField.Value, _ = strconv.Atoi(sVal)
				} else {
					vField.Value = sVal
				}
				valFields = append(valFields, vField)
			} else {
				log.Printf("skipping field %s\n", v.Name)
			}
		}
	}

	db := dbConn()
	defer db.Close()

	strQuery := "UPDATE " + schemaV.Table + " SET "
	var dbVals = make([]any, 0)
	for i, v := range valFields {
		if i != 0 {
			strQuery += ", "
		}
		strQuery += v.Column + "=?"
		dbVals = append(dbVals, v.Value)
	}
	var vId any
	if idField.Type == "int" {
		nId := 0
		nId, _ = strconv.Atoi(sId)
		vId = nId
	} else {
		vId = sId
	}
	dbVals = append(dbVals, vId)

	strQuery += " WHERE id=?"
	log.Printf("prepare query [%s]\n", strQuery)
	insForm, err := db.Prepare(strQuery)
	if err != nil {
		panic(err.Error())
	}
	insForm.Exec(dbVals...)

	http.Redirect(w, r, GMConfigV.URLDir+"/"+GMConfigV.DefaultAction+"/"+schemaName, http.StatusMovedPermanently)
}

func GMDelete(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	GMView(w, r, schemaName, sId, "Show", "delete")
}

func GMRemove(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	schemaFile := GMConfigV.SchemaDir + "/" + schemaName + ".json"
	schemaBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Printf("unavailable schema file %s\n", schemaFile)
		os.Exit(1)
	}
	var schemaV = GMSchema{}
	err = json.Unmarshal(schemaBytes, &schemaV)
	if err != nil {
		log.Printf("invalid content in schema file %s\n", schemaFile)
		os.Exit(1)
	}
	var selFields = []GMSchemaField{}
	for _, v := range schemaV.Fields {
		if v.Name == schemaV.Query.IdField {
			log.Printf("using field %s\n", v.Name)
			selFields = append(selFields, v)
			break
		}
	}
	strQuery := " DELETE FROM " + schemaV.Table + " WHERE " + selFields[0].Column + " = ?"
	var vId any
	if selFields[0].Type == "int" {
		nId := 0
		nId, _ = strconv.Atoi(sId)
		vId = nId
	} else {
		vId = sId
	}
	db := dbConn()
	defer db.Close()
	delForm, err := db.Prepare(strQuery)
	if err != nil {
		panic(err.Error())
	}
	delForm.Exec(vId)
	http.Redirect(w, r, GMConfigV.URLDir+"/"+GMConfigV.DefaultAction+"/"+schemaName, http.StatusMovedPermanently)
}

func GMRequestHandler(w http.ResponseWriter, r *http.Request) {
	sURL := strings.TrimSpace(r.URL.Path)
	if sURL == "/favicon.ico" {
		return
	}
	log.Printf("incoming URL value: %s\n", sURL)
	if sURL == "/" {
		sURL = GMConfigV.URLDir + "/" + GMConfigV.DefaultAction + "/" + GMConfigV.DefaultSchema
	}
	if !strings.HasPrefix(sURL, "/") {
		sURL = "/" + sURL
	}
	if (sURL == GMConfigV.URLDir) || (sURL == "/"+GMConfigV.URLDir) ||
		(sURL == "/"+GMConfigV.URLDir+"/") || (sURL == GMConfigV.URLDir+"/") {
		sURL = GMConfigV.URLDir + "/" + GMConfigV.DefaultAction + "/" + GMConfigV.DefaultSchema
	}
	log.Printf("updated URL value: %s\n", sURL)
	tURL := strings.Split(sURL, "/")
	if len(tURL) < 4 {
		log.Printf("too few tokens in URL: %s\n", sURL)
		http.Error(w, "Too few tokens", http.StatusBadRequest)
		return
	}
	if strings.HasPrefix(GMConfigV.URLDir, "/") {
		tURL[1] = "/" + tURL[1]
	}
	if GMConfigV.URLDir != tURL[1] {
		log.Printf("invalid URL prefix: %s (%s | %s)\n", sURL, GMConfigV.URLDir, tURL[1])
		http.Error(w, "Invalid URL prefix", http.StatusBadRequest)
		return
	}
	if tURL[2] == "list" {
		GMList(w, r, tURL[3])
	} else if tURL[2] == "new" {
		GMNew(w, r, tURL[3])
	} else if tURL[2] == "insert" {
		GMInsert(w, r, tURL[3])
	} else if tURL[2] == "show" {
		if len(tURL) < 5 {
			log.Printf("too few tokens in URL: %s\n", sURL)
			http.Error(w, "Too few tokens", http.StatusBadRequest)
			return
		}
		GMShow(w, r, tURL[3], tURL[4])
	} else if tURL[2] == "edit" {
		if len(tURL) < 5 {
			log.Printf("too few tokens in URL: %s\n", sURL)
			http.Error(w, "Too few tokens", http.StatusBadRequest)
			return
		}
		GMEdit(w, r, tURL[3], tURL[4])
	} else if tURL[2] == "update" {
		if len(tURL) < 5 {
			log.Printf("too few tokens in URL: %s\n", sURL)
			http.Error(w, "Too few tokens", http.StatusBadRequest)
			return
		}
		GMUpdate(w, r, tURL[3], tURL[4])
	} else if tURL[2] == "delete" {
		if len(tURL) < 5 {
			log.Printf("too few tokens in URL: %s\n", sURL)
			http.Error(w, "Too few tokens", http.StatusBadRequest)
			return
		}
		GMDelete(w, r, tURL[3], tURL[4])
	} else if tURL[2] == "remove" {
		if len(tURL) < 5 {
			log.Printf("too few tokens in URL: %s\n", sURL)
			http.Error(w, "Too few tokens", http.StatusBadRequest)
			return
		}
		GMRemove(w, r, tURL[3], tURL[4])
	} else {
		log.Printf("invalid URL value: %s\n", sURL)
		http.Error(w, "Invalid URL value", http.StatusBadRequest)
		return
	}
}

func printCLIOptions() {
	type CLIOptionDef struct {
		Ops      []string
		Usage    string
		DefValue string
		VType    string
	}
	var items []CLIOptionDef
	flag.VisitAll(func(f *flag.Flag) {
		var found bool = false
		for idx, it := range items {
			if it.Usage == f.Usage {
				found = true
				it.Ops = append(it.Ops, f.Name)
				items[idx] = it
			}
		}
		if !found {
			items = append(items, CLIOptionDef{
				Ops:      []string{f.Name},
				Usage:    f.Usage,
				DefValue: f.DefValue,
				VType:    fmt.Sprintf("%T", f.Value),
			})
		}
	})
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Ops[0]) <
			strings.ToLower(items[j].Ops[0])
	})
	for _, val := range items {
		vtype := val.VType[6 : len(val.VType)-5]
		if vtype[len(vtype)-2:] == "64" {
			vtype = vtype[:len(vtype)-2]
		}
		for _, opt := range val.Ops {
			if vtype == "bool" {
				fmt.Printf("  -%s\n", opt)
			} else {
				fmt.Printf("  -%s %s\n", opt, vtype)
			}
		}
		if vtype != "bool" && len(val.DefValue) > 0 {
			fmt.Printf("      %s [default: %s]\n", val.Usage, val.DefValue)
		} else {
			fmt.Printf("      %s\n", val.Usage)
		}
	}
}

// initialize application components
func init() {
	// command line arguments
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s (v%s):\n", filepath.Base(os.Args[0]), siregisVersion)
		printCLIOptions()
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}

	flag.StringVar(&GMCLIOptionsV.config, "config", GMCLIOptionsV.config, "path to json config file")
	flag.StringVar(&GMCLIOptionsV.domain, "domain", GMCLIOptionsV.domain, "http service domain")
	flag.StringVar(&GMCLIOptionsV.httpsrv, "http-srv", GMCLIOptionsV.httpsrv, "http server bind address")
	flag.StringVar(&GMCLIOptionsV.httpssrv, "https-srv", GMCLIOptionsV.httpssrv, "https server bind address")
	flag.StringVar(&GMCLIOptionsV.httpspubkey, "https-pubkey", GMCLIOptionsV.httpspubkey, "https server public key")
	flag.StringVar(&GMCLIOptionsV.httpsprvkey, "https-prvkey", GMCLIOptionsV.httpsprvkey, "https server private key")
	flag.BoolVar(&GMCLIOptionsV.httpsusele, "use-letsencrypt", GMCLIOptionsV.httpsusele,
		"use local letsencrypt certificates (requires domain)")
	flag.BoolVar(&GMCLIOptionsV.version, "version", GMCLIOptionsV.version, "print version")
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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

	GMFuncMap["ha1"] = GMFuncHA1
	GMFuncMap["ha1b"] = GMFuncHA1B

	log.Println("Starting server on: http://localhost:8284")
	http.HandleFunc("/", GMRequestHandler)
	// http.HandleFunc("/show", Show)
	http.ListenAndServe(":8284", nil)
}
