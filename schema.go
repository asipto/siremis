package main

type GMSchemaQuery struct {
	IdField   string `json:"IdField"`
	Limit     int    `json:"Limit,omitempty"`
	OrderBy   string `json:"OrderBy,omitempty"`
	OrderType string `json:"OrderType,omitempty"`
}

type GMSchemaInactiveActions struct {
	Insert bool `json:"Insert,omitempty"`
	Edit   bool `json:"Edit,omitempty"`
	Delete bool `json:"Delete,omitempty"`
}

type GMSchemaFieldEnable struct {
	List    bool `json:"List,omitempty"`
	Insert  bool `json:"Insert,omitempty"`
	Edit    bool `json:"Edit,omitempty"`
	Show    bool `json:"Show,omitempty"`
	Search  bool `json:"Search,omitempty"`
	Discard bool `json:"Discard,omitempty"`
}

type GMSchemaFieldOptionValues struct {
	Func   string   `json:"Func,omitempty"`
	Params []string `json:"Params,omitempty"`
}

type GMSchemaInputForm struct {
	Type         string                    `json:"Type,omitempty"`
	OptionValues GMSchemaFieldOptionValues `json:"OptionValues,omitempty"`
}

type GMSchemaFieldValue struct {
	Mode   int      `json:"Mode,omitempty"`
	Func   string   `json:"Func,omitempty"`
	Params []string `json:"Params,omitempty"`
}

type GMSchemaField struct {
	Name        string              `json:"Name"`
	Title       string              `json:"Title"`
	Column      string              `json:"Column"`
	Type        string              `json:"Type"`
	Enable      GMSchemaFieldEnable `json:"Enable,omitempty"`
	InputForm   GMSchemaInputForm   `json:"InputForm,omitempty"`
	ValueInsert GMSchemaFieldValue  `json:"ValueInsert,omitempty"`
	ValueEdit   GMSchemaFieldValue  `json:"ValueEdit,omitempty"`
	ValueShow   GMSchemaFieldValue  `json:"ValueShow,omitempty"`
}

type GMSchema struct {
	Name            string                  `json:"Name"`
	Title           string                  `json:"Title"`
	Table           string                  `json:"Table"`
	MenuGroup       string                  `json:"MenuGroup,omitempty"`
	InactiveActions GMSchemaInactiveActions `json:"InactiveActions,omitempty"`
	Query           GMSchemaQuery           `json:"Query"`
	Fields          []GMSchemaField         `json:"Fields"`
}
