package main

type GMSchemaQuery struct {
	IdField   string `json:"IdField"`
	Limit     int    `json:"Limit,omitempty"`
	OrderBy   string `json:"OrderBy,omitempty"`
	OrderType string `json:"OrderType,omitempty"`
}

type GMSchemaFieldEnable struct {
	List   bool `json:"List,omitempty"`
	Insert bool `json:"Insert,omitempty"`
	Edit   bool `json:"Edit,omitempty"`
	Show   bool `json:"Show,omitempty"`
	Filter bool `json:"Filter,omitempty"`
}

type GMSchemaFieldValue struct {
	Func   string   `json:"Func,omitempty"`
	Params []string `json:"Params,omitempty"`
}

type GMSchemaField struct {
	Name        string              `json:"Name"`
	Title       string              `json:"Title"`
	Column      string              `json:"Column"`
	Type        string              `json:"Type"`
	Enable      GMSchemaFieldEnable `json:"Enable,omitempty"`
	ValueInsert GMSchemaFieldValue  `json:"ValueInsert,omitempty"`
	ValueEdit   GMSchemaFieldValue  `json:"ValueEdit,omitempty"`
}

type GMSchema struct {
	Name   string          `json:"Name"`
	Title  string          `json:"Title"`
	Table  string          `json:"Table"`
	Query  GMSchemaQuery   `json:"Query"`
	Fields []GMSchemaField `json:"Fields"`
}
