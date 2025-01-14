package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func GMGetSchema(w http.ResponseWriter, r *http.Request, schemaName string) (*GMSchema, bool) {
	schemaFile := GMConfigV.SchemaDir + "/" + schemaName + ".json"
	schemaBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Printf("unavailable schema file %s\n", schemaFile)
		GMAlertView(w, r, "", "",
			"Unavailable schema file for: "+schemaName)
		return nil, false
	}
	var schemaV = GMSchema{}
	err = json.Unmarshal(schemaBytes, &schemaV)
	if err != nil {
		log.Printf("invalid content in schema file %s\n", schemaFile)
		GMAlertView(w, r,
			schemaName, strings.ToUpper(schemaName[:1])+strings.ToLower(schemaName[1:]),
			"Invalid content in schema file for: "+schemaName)
		return nil, false
	}

	return &schemaV, true
}

func GMAlertView(w http.ResponseWriter, r *http.Request, schemaName string,
	schemaTitle string, alertText string) {

	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.AuthOK = GMSessionAuthActive(w, r)

	viewData.Context.SchemaName = schemaName
	viewData.Context.SchemaTitle = schemaTitle

	viewData.Context.Alert.Active = true
	viewData.Context.Alert.Type = "alert"
	viewData.Context.Alert.Text = alertText

	GMTemplatesV.ExecuteTemplate(w, "alert", viewData)
}

func GMList(w http.ResponseWriter, r *http.Request, schemaName string,
	listParams []string) {

	if GMSessionAuthCheck(w, r) < 0 {
		GMViewGuestPage(w, r, "login")
		return
	}
	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}
	var selFields = []GMSchemaField{}
	for _, v := range schemaV.Fields {
		if v.Name == schemaV.Query.IdField {
			log.Printf("using field %s (list)\n", v.Name)
			selFields = append(selFields, v)
			break
		}
	}
	for _, v := range schemaV.Fields {
		if v.Name != schemaV.Query.IdField {
			if v.Enable.List {
				log.Printf("using field %s (list)\n", v.Name)
				selFields = append(selFields, v)
			} else {
				log.Printf("skipping field %s (list)\n", v.Name)
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
	strQuery += " FROM " + schemaV.Table

	if len(schemaV.Query.OrderBy) > 0 {
		strQuery += " ORDER BY " + schemaV.Query.OrderBy
		if len(schemaV.Query.OrderType) > 0 {
			strQuery += " " + schemaV.Query.OrderType
		}
	}

	groupV := 0
	if schemaV.Query.Limit > 0 {
		if len(listParams) > 1 {
			if listParams[0] == "group" {
				groupV, _ = strconv.Atoi(listParams[1])
				if groupV < 0 {
					groupV = 0
				}
			}
		}
		offsetV := groupV * schemaV.Query.Limit
		strQuery += " LIMIT " + strconv.Itoa(offsetV) + ", " +
			strconv.Itoa(schemaV.Query.Limit)
	}

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
		//log.Println("listing row: id: " + strconv.Itoa(*dbRow[0].(*int)))
		dbRes = append(dbRes, dbRow)
	}

	GMAuthRefresh(w, r)

	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.Action = "list"
	viewData.Context.AuthOK = true
	viewData.Context.SchemaName = schemaV.Name
	viewData.Context.SchemaTitle = schemaV.Title
	viewData.Context.ResultAttrs.NrRows = len(dbRes)
	viewData.Context.ResultAttrs.NrGroup = groupV
	if schemaV.Query.Limit > 0 {
		viewData.Context.ResultAttrs.NrGroupPrev = groupV - 1
		if schemaV.Query.Limit == viewData.Context.ResultAttrs.NrRows {
			viewData.Context.ResultAttrs.NrGroupNext = groupV + 1
		}
	}
	viewData.Schema = *schemaV
	viewData.Fields = selFields
	viewData.Values = dbRes

	GMTemplatesV.ExecuteTemplate(w, "list", viewData)
}

func GMSchemaFieldEnableBoolVal(v *GMSchemaFieldEnable, field string) bool {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	return bool(f.Bool())
}

func GMFormView(w http.ResponseWriter, r *http.Request, schemaName string, sId string,
	sEnableField string, sTemplate string) {

	if GMSessionAuthCheck(w, r) < 0 {
		GMViewGuestPage(w, r, "login")
		return
	}

	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}
	var selFields = []GMSchemaField{}
	for _, v := range schemaV.Fields {
		if v.Name == schemaV.Query.IdField {
			log.Printf("using field %s (%s)\n", v.Name, sEnableField)
			selFields = append(selFields, v)
			break
		}
	}
	for _, v := range schemaV.Fields {
		if v.Name != schemaV.Query.IdField {
			if GMSchemaFieldEnableBoolVal(&v.Enable, sEnableField) {
				log.Printf("using field %s (%s)\n", v.Name, sEnableField)
				selFields = append(selFields, v)
			} else {
				log.Printf("skipping field %s (%s)\n", v.Name, sEnableField)
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

	GMAuthRefresh(w, r)
	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.AuthOK = true
	viewData.Context.Action = sTemplate
	viewData.Context.SchemaName = schemaV.Name
	viewData.Context.SchemaTitle = schemaV.Title
	viewData.Schema = *schemaV
	viewData.Fields = selFields
	viewData.Values = dbRes
	GMTemplatesV.ExecuteTemplate(w, sTemplate, viewData)
}

func GMShow(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	GMFormView(w, r, schemaName, sId, "Show", "show")
}

func GMNew(w http.ResponseWriter, r *http.Request, schemaName string) {
	if GMSessionAuthCheck(w, r) < 0 {
		GMViewGuestPage(w, r, "login")
		return
	}
	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}
	var formFields = []GMViewFormField{}
	for _, v := range schemaV.Fields {
		if v.Name != schemaV.Query.IdField {
			if v.Enable.Insert {
				log.Printf("using field %s (insert)\n", v.Name)
				var fField = GMViewFormField{}
				fField.Field = v
				if len(v.InputForm.OptionValues.Func) > 0 {
					if len(v.InputForm.OptionValues.Params) > 0 {
						var vParams = make([]any, 0)
						for _, p := range v.InputForm.OptionValues.Params {
							vParams = append(vParams, p)
						}
						fField.OptionValues = GMFuncMap[v.InputForm.OptionValues.Func].(func([]any) []string)(vParams)
					} else {
						fField.OptionValues = GMFuncMap[v.InputForm.OptionValues.Func].(func() []string)()
					}
				}
				formFields = append(formFields, fField)
			} else {
				log.Printf("skipping field %s (insert)\n", v.Name)
			}
		}
	}

	GMAuthRefresh(w, r)
	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.Action = "new"
	viewData.Context.AuthOK = true
	viewData.Context.SchemaName = schemaV.Name
	viewData.Context.SchemaTitle = schemaV.Title
	viewData.Schema = *schemaV
	viewData.FormFields = formFields
	GMTemplatesV.ExecuteTemplate(w, "new", viewData)
}

func GMInsert(w http.ResponseWriter, r *http.Request, schemaName string) {
	if GMSessionAuthCheck(w, r) < 0 {
		GMViewGuestPage(w, r, "login")
		return
	}
	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}
	var valFields = []GMDBField{}
	for _, v := range schemaV.Fields {
		if v.Name != schemaV.Query.IdField {
			var vField = GMDBField{}
			vField.Name = v.Name
			vField.Column = v.Column
			if v.Enable.Insert {
				if v.Type == "int" {
					vField.Value, _ = strconv.Atoi(r.FormValue(v.Name))
				} else {
					vField.Value = r.FormValue(v.Name)
				}
				log.Printf("insert form field: %s / value: '%v'\n", v.Name, vField.Value)
				valFields = append(valFields, vField)
			} else if len(v.ValueInsert.Func) > 0 {
				sVal := ""
				if len(v.ValueInsert.Params) > 0 {
					var vParams = make([]any, 0)
					for _, p := range v.ValueInsert.Params {
						if strings.HasPrefix(p, "@fld:") {
							fldName := strings.TrimPrefix(p, "@fld:")
							for _, f := range valFields {
								if fldName == f.Name {
									vParams = append(vParams, f.Value)
									break
								}
							}
						} else {
							vParams = append(vParams, p)
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
				log.Printf("insert func field: %s / value: '%v'\n", v.Name, vField.Value)
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
	GMAuthRefresh(w, r)
	GMSMenuPage(w, r, schemaName)
}

func GMEdit(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	GMFormView(w, r, schemaName, sId, "Edit", "edit")
}

func GMUpdate(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	if GMSessionAuthCheck(w, r) < 0 {
		GMViewGuestPage(w, r, "login")
		return
	}
	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
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
			if v.Enable.Edit {
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
						if strings.HasPrefix(p, "@fld:") {
							fldName := strings.TrimPrefix(p, "@fld:")
							for _, f := range valFields {
								if fldName == f.Name {
									vParams = append(vParams, f.Value)
									break
								}
							}
						} else {
							vParams = append(vParams, p)
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

	GMAuthRefresh(w, r)
	GMFormView(w, r, schemaName, sId, "Show", "show")
}

func GMDelete(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	GMFormView(w, r, schemaName, sId, "Show", "delete")
}

func GMRemove(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	if GMSessionAuthCheck(w, r) < 0 {
		GMViewGuestPage(w, r, "login")
		return
	}
	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}
	var selFields = []GMSchemaField{}
	for _, v := range schemaV.Fields {
		if v.Name == schemaV.Query.IdField {
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
	GMAuthRefresh(w, r)
	GMSMenuPage(w, r, schemaName)
}

func GMLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		GMViewGuestPage(w, r, "login")
		return
	}
	res := GMLoginCheck(w, r)

	if res < 0 {
		// w.WriteHeader(http.StatusUnauthorized)
		GMViewGuestPage(w, r, "login")
		return
	}

	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.AuthOK = true
	GMTemplatesV.ExecuteTemplate(w, "main", viewData)
}

func GMLogout(w http.ResponseWriter, r *http.Request) int {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// w.WriteHeader(http.StatusUnauthorized)
			return -1
		}
		// w.WriteHeader(http.StatusBadRequest)
		return -2
	}
	sessionToken := c.Value

	delete(GMSessions, sessionToken)

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now(),
	})

	return 0
}

func GMViewPage(w http.ResponseWriter, r *http.Request, sTemplate string) {
	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.AuthOK = GMSessionAuthActive(w, r)

	GMTemplatesV.ExecuteTemplate(w, sTemplate, viewData)
}

func GMViewAuthPage(w http.ResponseWriter, r *http.Request, sTemplate string) {
	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.AuthOK = GMSessionAuthActive(w, r)
	if !viewData.Context.AuthOK {
		GMViewGuestPage(w, r, "login")
		return
	}

	GMTemplatesV.ExecuteTemplate(w, sTemplate, viewData)
}

func GMViewGuestPage(w http.ResponseWriter, r *http.Request, sTemplate string) {
	var viewData = GMViewData{}
	viewData.Config = GMConfigV

	GMTemplatesV.ExecuteTemplate(w, sTemplate, viewData)
}

func GMSMenuPage(w http.ResponseWriter, r *http.Request, schemaName string) {
	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.AuthOK = GMSessionAuthActive(w, r)
	if !viewData.Context.AuthOK {
		GMViewGuestPage(w, r, "login")
		return
	}

	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}

	viewData.Context.SchemaName = schemaV.Name
	viewData.Context.SchemaTitle = schemaV.Title

	GMTemplatesV.ExecuteTemplate(w, "smenu", viewData)
}

func GMDoAction(w http.ResponseWriter, r *http.Request, sAction string) {
	if sAction == "login" {
		GMLogin(w, r)
	} else if sAction == "logout" {
		GMLogout(w, r)
	} else {
		GMViewGuestPage(w, r, "login")
		return
	}
}

func GMRequestHandler(w http.ResponseWriter, r *http.Request) {
	sURL := strings.TrimSpace(r.URL.Path)
	if sURL == "/favicon.ico" {
		return
	}
	log.Printf("incoming URL value: %s\n", sURL)
	if sURL == "/" {
		sURL = GMConfigV.URLDir + "/" + GMConfigV.DefaultViewPath
	}
	if !strings.HasPrefix(sURL, "/") {
		sURL = "/" + sURL
	}
	if (sURL == GMConfigV.URLDir) || (sURL == "/"+GMConfigV.URLDir) ||
		(sURL == "/"+GMConfigV.URLDir+"/") || (sURL == GMConfigV.URLDir+"/") {
		sURL = GMConfigV.URLDir + "/" + GMConfigV.DefaultViewPath
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
		if len(tURL) > 4 {
			GMList(w, r, tURL[3], tURL[4:])
		} else {
			GMList(w, r, tURL[3], []string{})
		}
	} else if tURL[2] == "view" {
		GMViewPage(w, r, tURL[3])
	} else if tURL[2] == "do" {
		GMDoAction(w, r, tURL[3])
	} else if tURL[2] == "new" {
		GMNew(w, r, tURL[3])
	} else if tURL[2] == "insert" {
		GMInsert(w, r, tURL[3])
	} else if tURL[2] == "menu" {
		GMSMenuPage(w, r, tURL[3])
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
