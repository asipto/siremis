package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type GMAlert struct {
	Active bool
	Type   string
	Text   string
}

type GMResultAttrs struct {
	NrRows      int
	NrGroup     int
	NrGroupPrev int
	NrGroupNext int
}

type GMJRPCData struct {
	Command  string
	Result   string
	ViewForm GMConfigJRForm
}

type GMViewContext struct {
	Action       string
	AuthOK       bool
	SchemaName   string
	SchemaTitle  string
	SchemaMenu   *GMConfigMenuGroup
	IdField      GMSchemaField
	IdFieldValue any
	ResultAttrs  GMResultAttrs
	Alert        GMAlert
}

type GMViewList struct {
	Fields []GMSchemaField
	Values []any
}

type GMViewFormField struct {
	Field        GMSchemaField
	Value        any
	SValue       string
	OptionValues []GMOptionValue
}

type GMChartsData struct {
	Name  string
	Title string
	Data  string
}

type GMChartsInfo struct {
	Name       string
	Title      string
	ChartsData []GMChartsData
}
type GMViewData struct {
	Config      GMConfig
	Schema      GMSchema
	Context     GMViewContext
	ViewList    GMViewList
	FormFieldId GMViewFormField
	FormFields  []GMViewFormField
	JRPCData    *GMJRPCData
	ChartsInfo  *GMChartsInfo
}

type GMDBField struct {
	Name    string
	Column  string
	Op      string
	Value   any
	Discard bool
}

func GMSchemaExists(schemaName string) bool {
	if _, err := os.Stat(GMConfigV.SchemasDir + "/" + schemaName + ".json"); err == nil {
		return true
	}
	return false
}

func GMGetSchema(w http.ResponseWriter, r *http.Request, schemaName string) (*GMSchema, bool) {
	schemaFile := GMConfigV.SchemasDir + "/" + schemaName + ".json"
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

func GMLoadSchema(schemaName string) (*GMSchema, bool) {
	schemaFile := GMConfigV.SchemasDir + "/" + schemaName + ".json"
	schemaBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Printf("unavailable schema file %s\n", schemaFile)
		return nil, false
	}
	var schemaV = GMSchema{}
	err = json.Unmarshal(schemaBytes, &schemaV)
	if err != nil {
		log.Printf("invalid content in schema file %s\n", schemaFile)
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
	viewData.Context.SchemaMenu = GMConfigGetMenu(schemaName, "")

	viewData.Context.Alert.Active = true
	viewData.Context.Alert.Type = "alert"
	viewData.Context.Alert.Text = alertText

	GMTemplatesV.ExecuteTemplate(w, "alert", viewData)
}

func GMList(w http.ResponseWriter, r *http.Request, schemaName string,
	listParams []string) {

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
			strQuery += dbColumnQuoted(v.Column)
		} else {
			strQuery += ", " + dbColumnQuoted(v.Column)
		}
	}
	strQuery += " FROM " + schemaV.Table

	if len(schemaV.Query.OrderBy) > 0 {
		strQuery += " ORDER BY " + dbColumnQuoted(schemaV.Query.OrderBy)
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
		log.Printf("query failed [[%s]]\n", strQuery)
		panic(err.Error())
	}
	defer db.Close()
	dbRes := make([]any, 0)

	for selDB.Next() {
		dbRow := make([]any, len(selFields))
		for i, v := range selFields {
			if v.Type == "int" {
				dbRow[i] = new(sql.NullInt64)
			} else if v.Type == "float" {
				dbRow[i] = new(sql.NullFloat64)
			} else if v.Type == "str" || v.Type == "string" {
				dbRow[i] = new(sql.NullString)
			} else {
				dbRow[i] = new(sql.NullString)
			}
		}
		err := selDB.Scan(dbRow...)
		if err != nil {
			panic(err.Error())
		}
		dbVals := make([]any, len(selFields))
		for i := range dbRow {
			switch fv := dbRow[i].(type) {
			case *sql.NullString:
				if fv.Valid {
					dbVals[i] = fv.String
				} else {
					dbVals[i] = ""
				}
			case *sql.NullInt64:
				if fv.Valid {
					dbVals[i] = fv.Int64
				} else {
					dbVals[i] = 0
				}
			case *sql.NullFloat64:
				if fv.Valid {
					dbVals[i] = fv.Float64
				} else {
					dbVals[i] = 0.0
				}
			default:
				dbVals[i] = dbRow[i]
			}
		}
		//log.Println("listing row: id: " + strconv.Itoa(*dbRow[0].(*int)))
		dbRes = append(dbRes, dbVals)
	}

	GMAuthRefresh(w, r)

	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.Action = "list"
	viewData.Context.AuthOK = true
	viewData.Context.SchemaName = schemaV.Name
	viewData.Context.SchemaTitle = schemaV.Title
	viewData.Context.SchemaMenu = GMConfigGetMenu(schemaName, schemaV.MenuGroup)
	viewData.Context.ResultAttrs.NrRows = len(dbRes)
	viewData.Context.ResultAttrs.NrGroup = groupV
	if schemaV.Query.Limit > 0 {
		viewData.Context.ResultAttrs.NrGroupPrev = groupV - 1
		if schemaV.Query.Limit == viewData.Context.ResultAttrs.NrRows {
			viewData.Context.ResultAttrs.NrGroupNext = groupV + 1
		}
	}
	viewData.Schema = *schemaV
	viewData.ViewList.Fields = selFields
	viewData.ViewList.Values = dbRes

	GMTemplatesV.ExecuteTemplate(w, "list", viewData)
}

func GMSchemaFieldEnableBoolVal(v *GMSchemaFieldEnable, field string) bool {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	return bool(f.Bool())
}

func GMFormView(w http.ResponseWriter, r *http.Request, schemaName string, sId string,
	sEnableField string, sTemplate string, sAction string) {
	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}
	if sAction == "edit" {
		if schemaV.InactiveActions.Edit {
			GMAlertView(w, r, schemaV.Name, schemaV.Title,
				"Edit operation not permitted for: "+schemaV.Name)
			return
		}
	} else if sAction == "delete" {
		if schemaV.InactiveActions.Delete {
			GMAlertView(w, r, schemaV.Name, schemaV.Title,
				"Delete operation not permitted for: "+schemaV.Name)
			return
		}
	}
	var formFields = []GMViewFormField{}
	for _, v := range schemaV.Fields {
		if v.Name == schemaV.Query.IdField {
			log.Printf("using field %s (%s)\n", v.Name, sEnableField)
			var fField = GMViewFormField{}
			fField.Field = v
			formFields = append(formFields, fField)
			break
		}
	}
	for _, v := range schemaV.Fields {
		if v.Name != schemaV.Query.IdField {
			if GMSchemaFieldEnableBoolVal(&v.Enable, sEnableField) {
				log.Printf("using field %s (%s)\n", v.Name, sEnableField)
				var fField = GMViewFormField{}
				fField.Field = v
				if sAction == "edit" {
					fField.Field = v
					if len(v.InputForm.OptionValues.Func) > 0 {
						if len(v.InputForm.OptionValues.Params) > 0 {
							var vParams = make([]any, 0)
							for _, p := range v.InputForm.OptionValues.Params {
								vParams = append(vParams, p)
							}
							fField.OptionValues = GMFuncMap[v.InputForm.OptionValues.Func].(func([]any) []GMOptionValue)(vParams)
						} else {
							fField.OptionValues = GMFuncMap[v.InputForm.OptionValues.Func].(func() []GMOptionValue)()
						}
					}
				}
				formFields = append(formFields, fField)
			} else {
				log.Printf("skipping field %s (%s)\n", v.Name, sEnableField)
			}
		}
	}
	strQuery := "SELECT "
	for i, v := range formFields {
		if i == 0 {
			strQuery += dbColumnQuoted(v.Field.Column)
		} else {
			strQuery += ", " + dbColumnQuoted(v.Field.Column)
		}
	}
	strQuery += " FROM " + schemaV.Table + " WHERE " + formFields[0].Field.Column + " = ?"
	db := dbConn()
	defer db.Close()
	var vId any
	if formFields[0].Field.Type == "int" {
		nId := 0
		nId, _ = strconv.Atoi(sId)
		vId = nId
	} else {
		vId = sId
	}
	selDB, err := db.Query(strQuery, vId)
	if err != nil {
		log.Printf("failed query [[%s]] (%v)\n", strQuery, vId)
		panic(err.Error())
	}

	for selDB.Next() {
		dbRow := make([]any, len(formFields))
		for i, v := range formFields {
			if v.Field.Type == "int" {
				dbRow[i] = new(sql.NullInt64)
			} else if v.Field.Type == "float" {
				dbRow[i] = new(sql.NullFloat64)
			} else if v.Field.Type == "str" || v.Field.Type == "string" {
				dbRow[i] = new(sql.NullString)
			} else {
				dbRow[i] = new(sql.NullString)
			}
		}
		err := selDB.Scan(dbRow...)
		if err != nil {
			panic(err.Error())
		}
		log.Println("listing row: id: " + strconv.FormatInt((*(dbRow[0].(*sql.NullInt64))).Int64, 10))

		for i, v := range dbRow {
			var rv any
			if formFields[i].Field.Type == "int" {
				fv := *(v.(*sql.NullInt64))
				if fv.Valid {
					rv = fv.Int64
					formFields[i].SValue = strconv.FormatInt(fv.Int64, 10)
				} else {
					rv = int64(0)
					formFields[i].SValue = "0"
				}
			} else if formFields[i].Field.Type == "float" {
				fv := *(v.(*sql.NullFloat64))
				if fv.Valid {
					rv = fv.Float64
					formFields[i].SValue = fmt.Sprintf("%.2f", fv.Float64)
				} else {
					formFields[i].Value = float64(0.00)
					formFields[i].SValue = "0.00"
				}
			} else {
				fv := *(v.(*sql.NullString))
				if fv.Valid {
					rv = fv.String
					formFields[i].SValue = fv.String
				} else {
					rv = ""
					formFields[i].SValue = ""
				}
			}

			if sAction == "show" && len(formFields[i].Field.ValueShow.Func) > 0 {
				var vParams = make([]any, 0)
				if len(formFields[i].Field.ValueShow.Params) > 0 {
					for _, p := range formFields[i].Field.ValueShow.Params {
						if strings.HasPrefix(p, "@fld:") {
							fldName := strings.TrimPrefix(p, "@fld:")
							if fldName == formFields[i].Field.Name {
								vParams = append(vParams, rv)
							} else {
								for j := 0; j < i; j++ {
									if fldName == formFields[j].Field.Name {
										vParams = append(vParams, formFields[j].Value)
										break
									}
								}
							}
						} else {
							vParams = append(vParams, p)
						}
					}
				} else {
					vParams = append(vParams, rv)
				}
				formFields[i].Value = GMFuncMap[formFields[i].Field.ValueShow.Func].(func([]any) string)(vParams)
			} else {
				formFields[i].Value = rv
			}
		}
	}

	GMAuthRefresh(w, r)
	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.AuthOK = true
	viewData.Context.Action = sTemplate
	viewData.Context.SchemaName = schemaV.Name
	viewData.Context.SchemaTitle = schemaV.Title
	viewData.Context.SchemaMenu = GMConfigGetMenu(schemaName, schemaV.MenuGroup)
	viewData.Context.IdField = formFields[0].Field
	viewData.Context.IdFieldValue = formFields[0].Value
	viewData.FormFields = formFields
	viewData.Schema = *schemaV
	GMTemplatesV.ExecuteTemplate(w, sTemplate, viewData)
}

func GMShow(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	GMFormView(w, r, schemaName, sId, "Show", "show", "show")
}

func GMNew(w http.ResponseWriter, r *http.Request, schemaName string) {
	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}
	if schemaV.InactiveActions.Insert {
		GMAlertView(w, r, schemaV.Name, schemaV.Title,
			"Insert operation not permitted for: "+schemaV.Name)
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
						fField.OptionValues = GMFuncMap[v.InputForm.OptionValues.Func].(func([]any) []GMOptionValue)(vParams)
					} else {
						fField.OptionValues = GMFuncMap[v.InputForm.OptionValues.Func].(func() []GMOptionValue)()
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
	viewData.Context.SchemaMenu = GMConfigGetMenu(schemaName, schemaV.MenuGroup)
	viewData.Schema = *schemaV
	viewData.FormFields = formFields
	GMTemplatesV.ExecuteTemplate(w, "new", viewData)
}

func GMInsert(w http.ResponseWriter, r *http.Request, schemaName string) {
	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}
	if schemaV.InactiveActions.Insert {
		GMAlertView(w, r, schemaV.Name, schemaV.Title,
			"Insert operation not permitted for : "+schemaV.Name)
		return
	}
	var valFields = []GMDBField{}
	for _, v := range schemaV.Fields {
		if v.Name != schemaV.Query.IdField {
			var vField = GMDBField{}
			vField.Name = v.Name
			vField.Column = v.Column
			vField.Discard = v.Enable.Discard
			sVal := ""
			if v.Enable.Insert {
				sVal = r.FormValue(v.Name)
				log.Printf("insert form field: %s / value: '%v'\n", v.Name, sVal)
			}
			if len(v.ValueInsert.Func) > 0 {
				if len(v.ValueInsert.Params) > 0 {
					var vParams = make([]any, 0)
					for _, p := range v.ValueInsert.Params {
						if strings.HasPrefix(p, "@fld:") {
							fldName := strings.TrimPrefix(p, "@fld:")
							if fldName == vField.Name {
								vParams = append(vParams, sVal)
							} else {
								for _, f := range valFields {
									if fldName == f.Name {
										vParams = append(vParams, f.Value)
										break
									}
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
				log.Printf("insert func field: %s / value: '%v'/ discard: %v\n",
					v.Name, sVal, vField.Discard)
			}
			if !v.Enable.Insert && len(v.ValueInsert.Func) == 0 {
				log.Printf("skipping insert field %s\n", v.Name)
			} else {
				if v.Type == "int" {
					vField.Value, _ = strconv.Atoi(sVal)
				} else if v.Type == "float" {
					vField.Value, _ = strconv.ParseFloat(sVal, 32)
				} else {
					vField.Value = sVal
				}
				valFields = append(valFields, vField)
			}
		}
	}

	db := dbConn()
	defer db.Close()
	strQCols := ""
	strQValQ := ""
	var dbVals = make([]any, 0)
	comma := false
	for _, v := range valFields {
		if !v.Discard {
			if comma {
				strQCols += ", "
				strQValQ += ", "
			} else {
				comma = true
			}
			strQCols += dbColumnQuoted(v.Column)
			strQValQ += "?"
			dbVals = append(dbVals, v.Value)
		}
	}
	insForm, err := db.Prepare("INSERT INTO " + schemaV.Table + " (" + strQCols + ") VALUES (" + strQValQ + ")")
	if err != nil {
		panic(err.Error())
	}
	insForm.Exec(dbVals...)
	GMAuthRefresh(w, r)
	if len(schemaV.MenuGroup) > 0 {
		GMSMenuPage(w, r, schemaV.MenuGroup)
	} else {
		GMSMenuPage(w, r, schemaName)
	}
}

func GMEdit(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	GMFormView(w, r, schemaName, sId, "Edit", "edit", "edit")
}

func GMUpdate(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}
	if schemaV.InactiveActions.Edit {
		GMAlertView(w, r, schemaV.Name, schemaV.Title,
			"Edit operation not permitted for: "+schemaV.Name)
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
			vField.Discard = v.Enable.Discard
			sVal := ""
			if v.Enable.Edit {
				sVal = r.FormValue(v.Name)
				log.Printf("edit form field: %s / value: '%v'\n", v.Name, sVal)
			}
			if len(v.ValueEdit.Func) > 0 {
				if len(v.ValueEdit.Params) > 0 {
					var vParams = make([]any, 0)
					for _, p := range v.ValueEdit.Params {
						if strings.HasPrefix(p, "@fld:") {
							fldName := strings.TrimPrefix(p, "@fld:")
							if fldName == vField.Name {
								vParams = append(vParams, sVal)
							} else {
								for _, f := range valFields {
									if fldName == f.Name {
										vParams = append(vParams, f.Value)
										break
									}
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
				log.Printf("edit func field: %s / value: '%v'/ discard: %v\n",
					v.Name, sVal, vField.Discard)
			}
			if !v.Enable.Edit && len(v.ValueEdit.Func) == 0 {
				log.Printf("skipping edit field %s\n", v.Name)
			} else {
				if v.Type == "int" {
					vField.Value, _ = strconv.Atoi(sVal)
				} else if v.Type == "float" {
					vField.Value, _ = strconv.ParseFloat(sVal, 32)
				} else {
					vField.Value = sVal
				}
				valFields = append(valFields, vField)
			}
		}
	}

	db := dbConn()
	defer db.Close()

	strQuery := "UPDATE " + schemaV.Table + " SET "
	var dbVals = make([]any, 0)
	comma := false
	for _, v := range valFields {
		if !v.Discard {
			if comma {
				strQuery += ", "
			} else {
				comma = true
			}
			strQuery += dbColumnQuoted(v.Column) + "=?"
			dbVals = append(dbVals, v.Value)
		}
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
	GMFormView(w, r, schemaName, sId, "Show", "show", "show")
}

func GMDelete(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	GMFormView(w, r, schemaName, sId, "Show", "delete", "delete")
}

func GMRemove(w http.ResponseWriter, r *http.Request, schemaName string, sId string) {
	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}
	if schemaV.InactiveActions.Delete {
		GMAlertView(w, r, schemaV.Name, schemaV.Title,
			"Delete operation not permitted for: "+schemaV.Name)
		return
	}
	var selFields = []GMSchemaField{}
	for _, v := range schemaV.Fields {
		if v.Name == schemaV.Query.IdField {
			selFields = append(selFields, v)
			break
		}
	}
	strQuery := " DELETE FROM " + schemaV.Table + " WHERE " +
		dbColumnQuoted(selFields[0].Column) + " = ?"
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
	if len(schemaV.MenuGroup) > 0 {
		GMSMenuPage(w, r, schemaV.MenuGroup)
	} else {
		GMSMenuPage(w, r, schemaName)
	}
}

func GMSearch(w http.ResponseWriter, r *http.Request, schemaName string) {
	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}
	var formFields = []GMViewFormField{}
	for _, v := range schemaV.Fields {
		if v.Name != schemaV.Query.IdField {
			if v.Enable.Search {
				log.Printf("using field %s (insert)\n", v.Name)
				var fField = GMViewFormField{}
				fField.Field = v
				formFields = append(formFields, fField)
			} else {
				log.Printf("skipping field %s (search)\n", v.Name)
			}
		}
	}

	GMAuthRefresh(w, r)
	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.Action = "search"
	viewData.Context.AuthOK = true
	viewData.Context.SchemaName = schemaV.Name
	viewData.Context.SchemaTitle = schemaV.Title
	viewData.Context.SchemaMenu = GMConfigGetMenu(schemaName, schemaV.MenuGroup)
	viewData.Schema = *schemaV
	viewData.FormFields = formFields
	GMTemplatesV.ExecuteTemplate(w, "search", viewData)
}

func GMFind(w http.ResponseWriter, r *http.Request, schemaName string) {
	schemaV, okey := GMGetSchema(w, r, schemaName)
	if !okey {
		return
	}
	var valFields = []GMDBField{}
	sMode := r.FormValue("search_mode")
	if sMode == "any" {
		sMode = " OR "
	} else {
		sMode = " AND "
	}
	for _, v := range schemaV.Fields {
		if v.Enable.Search {
			oVal := r.FormValue(v.Name + "_op")
			if len(oVal) > 0 && oVal != "skip" {
				var vField = GMDBField{}
				vField.Name = v.Name
				vField.Column = v.Column
				switch oVal {
				case "ne":
					vField.Op = "!="
				case "gt":
					vField.Op = ">"
				case "lt":
					vField.Op = "<"
				case "ge":
					vField.Op = ">="
				case "le":
					vField.Op = "<="
				case "like":
					vField.Op = "LIKE"
				case "regexp":
					vField.Op = "REGEXP"
				default:
					vField.Op = "="
				}
				sVal := r.FormValue(v.Name)
				log.Printf("search form field: %s / value: '%v'\n", v.Name, sVal)
				if v.Type == "int" {
					vField.Value, _ = strconv.Atoi(sVal)
				} else if v.Type == "float" {
					vField.Value, _ = strconv.Atoi(sVal)
					strconv.ParseFloat(sVal, 32)
				} else {
					vField.Value = sVal
				}
				valFields = append(valFields, vField)
			} else {
				log.Printf("skipping form search field %s\n", v.Name)
			}
		} else {
			log.Printf("skipping search field %s\n", v.Name)
		}
	}

	if len(valFields) == 0 {
		GMList(w, r, schemaName, []string{})
		return
	}

	var selFields = []GMSchemaField{}
	for _, v := range schemaV.Fields {
		if v.Name == schemaV.Query.IdField {
			log.Printf("using list field %s (list)\n", v.Name)
			selFields = append(selFields, v)
			break
		}
	}
	for _, v := range schemaV.Fields {
		if v.Name != schemaV.Query.IdField {
			if v.Enable.List {
				log.Printf("using list field %s (list)\n", v.Name)
				selFields = append(selFields, v)
			} else {
				log.Printf("skipping list field %s (list)\n", v.Name)
			}
		}
	}
	strQuery := "SELECT "
	for i, v := range selFields {
		if i == 0 {
			strQuery += dbColumnQuoted(v.Column)
		} else {
			strQuery += ", " + dbColumnQuoted(v.Column)
		}
	}
	strQuery += " FROM " + schemaV.Table + " WHERE "

	var selVals = make([]any, 0)
	sNext := false
	for _, v := range valFields {
		if sNext {
			strQuery += sMode
		} else {
			sNext = true
		}
		strQuery += v.Column + " " + v.Op + " ?"
		selVals = append(selVals, v.Value)
	}

	if len(schemaV.Query.OrderBy) > 0 {
		strQuery += " ORDER BY " + dbColumnQuoted(schemaV.Query.OrderBy)
		if len(schemaV.Query.OrderType) > 0 {
			strQuery += " " + schemaV.Query.OrderType
		}
	}

	groupV := 0
	if schemaV.Query.Limit > 0 {
		offsetV := groupV * schemaV.Query.Limit
		strQuery += " LIMIT " + strconv.Itoa(offsetV) + ", " +
			strconv.Itoa(schemaV.Query.Limit)
	}

	db := dbConn()
	selDB, err := db.Query(strQuery, selVals...)
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
			} else if v.Type == "float" {
				dbRow[i] = new(float32)
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
	viewData.Context.SchemaMenu = GMConfigGetMenu(schemaName, schemaV.MenuGroup)
	viewData.Context.ResultAttrs.NrRows = len(dbRes)
	viewData.Context.ResultAttrs.NrGroup = groupV
	if schemaV.Query.Limit > 0 {
		viewData.Context.ResultAttrs.NrGroupPrev = groupV - 1
		if schemaV.Query.Limit == viewData.Context.ResultAttrs.NrRows {
			viewData.Context.ResultAttrs.NrGroupNext = groupV + 1
		}
	}
	viewData.Schema = *schemaV
	viewData.ViewList.Fields = selFields
	viewData.ViewList.Values = dbRes

	GMTemplatesV.ExecuteTemplate(w, "list", viewData)
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

	GMViewGuestPage(w, r, "login")
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

func GMSMenuPage(w http.ResponseWriter, r *http.Request, menuName string) {
	var viewData = GMViewData{}
	viewData.Config = GMConfigV

	viewData.Context.AuthOK = GMSessionAuthActive(w, r)
	if !viewData.Context.AuthOK {
		GMViewGuestPage(w, r, "login")
		return
	}
	if GMSchemaExists(menuName) {
		schemaV, okey := GMGetSchema(w, r, menuName)
		if !okey {
			return
		}
		viewData.Schema = *schemaV
	}

	viewData.Context.SchemaMenu = GMConfigGetMenu(menuName, viewData.Schema.MenuGroup)
	viewData.Context.SchemaName = viewData.Context.SchemaMenu.Name
	viewData.Context.SchemaTitle = viewData.Context.SchemaMenu.Title

	GMTemplatesV.ExecuteTemplate(w, "smenu", viewData)
}

func GMDoJSONRPC(w http.ResponseWriter, r *http.Request) {
	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.AuthOK = GMSessionAuthActive(w, r)
	if !viewData.Context.AuthOK {
		GMViewGuestPage(w, r, "login")
		return
	}

	var jData = GMJRPCData{}

	jData.Command = r.FormValue("command")
	if len(jData.Command) > 0 {
		log.Printf("jsonrpc command { %s }\n", jData.Command)

		jData.Result, _ = GMJSONRPCExec(jData.Command)
	}
	viewData.Context.SchemaName = "jsonrpc"
	viewData.Context.SchemaTitle = "JSONRPC"
	viewData.JRPCData = &jData
	viewData.JRPCData.ViewForm = GMConfigV.JSONRPC.ViewForm

	GMTemplatesV.ExecuteTemplate(w, "jsonrpc", viewData)
}

func GMDoAction(w http.ResponseWriter, r *http.Request, sAction string) {
	if sAction == "login" {
		GMLogin(w, r)
	} else if sAction == "logout" {
		GMLogout(w, r)
	} else if sAction == "jsonrpc" {
		GMDoJSONRPC(w, r)
	} else {
		GMViewGuestPage(w, r, "login")
		return
	}
}

func GMViewCharts(w http.ResponseWriter, r *http.Request, sChartGroup string) {
	var viewData = GMViewData{}
	viewData.Config = GMConfigV
	viewData.Context.AuthOK = GMSessionAuthActive(w, r)
	if !viewData.Context.AuthOK {
		GMViewGuestPage(w, r, "login")
		return
	}
	g, _ := GMChartGroupGet(sChartGroup)
	if g == nil {
		GMAlertView(w, r, "charts", "Charts", "Charts group "+sChartGroup+" not found")
		return
	}
	var cInfo = GMChartsInfo{}
	cInfo.Name = g.Name
	cInfo.Title = g.Title
	viewData.ChartsInfo = &cInfo
	viewData.ChartsInfo.ChartsData = make([]GMChartsData, 0)
	for i, c := range g.Charts {
		var cData = GMChartsData{}
		cData.Name = c.Name
		cData.Title = c.Title
		cData.Data, _ = GMChartGetData(g, &g.Charts[i])
		log.Printf("incoming chart value: || %s ||\n", cData.Data)
		viewData.ChartsInfo.ChartsData = append(viewData.ChartsInfo.ChartsData, cData)
	}

	viewData.Context.SchemaName = "charts"
	viewData.Context.SchemaTitle = "Charts"

	GMTemplatesV.ExecuteTemplate(w, "chart", viewData)
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
	if tURL[2] != "do" || tURL[3] != "login" {
		if GMSessionAuthCheck(w, r) < 0 {
			GMViewGuestPage(w, r, "login")
			return
		}
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
	} else if tURL[2] == "search" {
		GMSearch(w, r, tURL[3])
	} else if tURL[2] == "find" {
		GMFind(w, r, tURL[3])
	} else if tURL[2] == "charts" {
		GMViewCharts(w, r, tURL[3])
	} else {
		log.Printf("invalid URL value: %s\n", sURL)
		http.Error(w, "Invalid URL value", http.StatusBadRequest)
		return
	}
}
