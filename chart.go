package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

type GMChartAxis struct {
	Name   string `json:"Name"`
	Column string `json:"Column"`
	Title  string `json:"Title"`
	Type   string `json:"Type,omitempty"`
	Color  string `json:"Color,omitempty"`
}

type GMChart struct {
	Name      string        `json:"Name"`
	Table     string        `json:"Table"`
	Title     string        `json:"Title"`
	Type      string        `json:"Type"`
	BGColor   string        `json:"BGColor,omitempty"`
	Order     string        `json:"Order,omitempty"`
	OrderBy   string        `json:"OrderBy"`
	OrderMode string        `json:"OrderMode"`
	Limit     int           `json:"Limit"`
	XAxis     GMChartAxis   `json:"XAxis"`
	YAxis     []GMChartAxis `json:"YAxis"`
}

type GMChartGroup struct {
	Name   string    `json:"Name"`
	Title  string    `json:"Title"`
	Charts []GMChart `json:"Charts"`
}

type GMEChartXAxis struct {
	Type string   `json:"type,omitempty"`
	Data []string `json:"data"`
}

type GMEChartYAxis struct {
	Type string `json:"type,omitempty"`
}

type GMEChartTitle struct {
	Text      string `json:"text,omitempty"`
	TextStyle struct {
		FontSize int `json:"fontSize,omitempty"`
	} `json:"textStyle,omitempty"`
	Top  int `json:"top,omitempty"`
	Left int `json:"left,omitempty"`
}

type GMEChartTooltip struct {
	Trigger string `json:"trigger,omitempty"`
}

type GMEChartLegend struct {
	Data  []string `json:"data,omitempty"`
	Top   int      `json:"top,omitempty"`
	Right int      `json:"right,omitempty"`
}

// series":[{"name":"Used Size","type":"line","smooth":0,
//    "itemStyle":{"normal":{"areaStyle":{"type":"default"}}}}]

type GMEChartSeries struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Smooth    int    `json:"smooth,omitempty"`
	ItemStyle struct {
		Normal struct {
			AreaStyle struct {
				Type string `json:"type,omitempty"`
			} `json:"areaStyle,omitempty"`
		} `json:"normal,omitempty"`
	} `json:"itemStyle,omitempty"`
	Data []int `json:"data"`
}

type GMEChart struct {
	Title   GMEChartTitle    `json:"title"`
	Tooltip GMEChartTooltip  `json:"tooltip"`
	Legend  GMEChartLegend   `json:"legend,omitempty"`
	Color   []string         `json:"color,omitempty"`
	XAxis   GMEChartXAxis    `json:"xAxis"`
	YAxis   GMEChartYAxis    `json:"yAxis"`
	Series  []GMEChartSeries `json:"series"`
}

func GMChartGroupGet(vChartGroup string) (*GMChartGroup, bool) {
	if len(GMConfigV.ChartGroups) == 0 {
		log.Printf("no chart groups\n")
		return nil, false
	}
	var g GMChartGroup
	var i int
	for i, g = range GMConfigV.ChartGroups {
		if g.Name == vChartGroup {
			break
		}
	}
	if i == len(GMConfigV.ChartGroups) {
		log.Printf("chart group %s not found\n", vChartGroup)
		return nil, false
	}
	return &GMConfigV.ChartGroups[i], true
}

func GMChartGroupGetChart(g *GMChartGroup, vChart string) (*GMChart, bool) {
	var c GMChart
	var i int
	for i, c = range g.Charts {
		if c.Name == vChart {
			break
		}
	}
	if i == len(g.Charts) {
		log.Printf("chart %s not found\n", vChart)
		return nil, false
	}
	return &g.Charts[i], true
}

func GMChartGetData(g *GMChartGroup, c *GMChart) (string, bool) {
	db := dbConn()
	defer db.Close()

	sqlQuery := "SELECT " + c.XAxis.Column
	for _, y := range c.YAxis {
		sqlQuery += ", " + y.Column
	}
	sqlQuery += " FROM " + c.Table + " ORDER BY " + c.OrderBy + " " + c.OrderMode
	sqlQuery += " LIMIT " + strconv.Itoa(c.Limit)
	selDB, err := db.Query(sqlQuery)
	if err != nil {
		log.Printf("error [%s]\n", err.Error())
		return "", false
	}

	dbRes := make([][]int, c.Limit)
	i := 0
	j := 0
	for i = 0; i < c.Limit; i++ {
		dbRes[i] = make([]int, 1+len(c.YAxis))
	}
	nrRows := 0
	for selDB.Next() {
		dbRow := make([]any, 1+len(c.YAxis))
		for j = 0; j < 1+len(c.YAxis); j++ {
			dbRow[j] = new(int)
		}
		err := selDB.Scan(dbRow...)
		if err != nil {
			log.Printf("error [%s]\n", err.Error())
			return "", false
		}
		for j = 0; j < 1+len(c.YAxis); j++ {
			dbRes[nrRows][j] = *(dbRow[j].(*int))
		}
		nrRows++
	}
	if nrRows == 0 {
		log.Printf("no data in db query result\n")
		return "", false
	}

	if c.Order == "reverse" {
		for i, j = 0, nrRows-1; i < j; i, j = i+1, j-1 {
			for k := 0; k < 1+len(c.YAxis); k++ {
				dbRes[i][k], dbRes[j][k] = dbRes[j][k], dbRes[i][k]
			}
		}
	}

	var vEChart = GMEChart{}
	vEChart.Title.Text = c.Title
	vEChart.Title.TextStyle.FontSize = 12
	vEChart.Title.Left = 20
	vEChart.Title.Top = 5
	vEChart.Color = make([]string, len(c.YAxis))
	var vColor = []string{"#FF7588", "#40C7CA", "#FFA87D"}
	for j := 0; j < len(c.YAxis); j++ {
		vEChart.Color[j] = vColor[j%len(vColor)]
	}
	vEChart.Tooltip.Trigger = "axis"
	vEChart.XAxis.Data = make([]string, nrRows)
	vEChart.Legend.Data = make([]string, len(c.YAxis))
	vEChart.Legend.Top = 28
	vEChart.Legend.Right = 20
	vEChart.Series = make([]GMEChartSeries, len(c.YAxis))
	for j := 0; j < len(c.YAxis); j++ {
		vEChart.Legend.Data[j] = c.YAxis[j].Title
		if len(c.YAxis[j].Color) > 0 {
			vEChart.Color[j] = c.YAxis[j].Color
		}
		if c.Type == "area" {
			vEChart.Series[j].Type = "line"
		} else {
			vEChart.Series[j].Type = c.Type
		}
		vEChart.Series[j].Smooth = 0
		vEChart.Series[j].ItemStyle.Normal.AreaStyle.Type = "default"
		vEChart.Series[j].Name = c.YAxis[j].Title
		vEChart.Series[j].Data = make([]int, nrRows)
	}

	for i = 0; i < nrRows; i++ {
		if c.XAxis.Type == "timestamp" {
			tm := time.Unix(int64(dbRes[i][0]), 0)
			hh, mm, ss := tm.Clock()
			vEChart.XAxis.Data[i] = fmt.Sprintf("%02d:%02d:%02d", hh, mm, ss)
		} else {
			vEChart.XAxis.Data[i] = strconv.Itoa(dbRes[i][0])
		}
		for j := 0; j < len(c.YAxis); j++ {
			vEChart.Series[j].Data[i] = dbRes[i][j+1]
		}
	}

	if c.XAxis.Type == "timestamp" {
		tmFirst := time.Unix(int64(dbRes[0][0]), 0)
		tmLast := time.Unix(int64(dbRes[nrRows-1][0]), 0)
		vEChart.Title.Text += " (" + tmFirst.Format(time.RFC3339) + " - " +
			tmLast.Format(time.RFC3339) + ")"
	}

	bJson, err := json.Marshal(&vEChart)
	if err != nil {
		log.Printf("failed to generate json for [%s/%s]\n", g.Title, c.Title)
		return "", false
	}
	return string(bJson), true
}
