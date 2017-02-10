package main

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"reflect"
	"strconv"
)

//ConsoleOut 函数用于向console输出漂亮的config表格
func ConsoleOut(config interface{}) string {
	var buf bytes.Buffer
	tw := tablewriter.NewWriter(&buf)
	tw.SetHeader([]string{"NO", "Config Key", "Config Vaule"})
	value := reflect.ValueOf(config)
	for i := 0; i < value.NumField(); i++ {
		tw.Append([]string{strconv.Itoa(i), value.Type().Field(i).Name, fmt.Sprintf("%# v", value.Field(i))})
	}
	tw.Render()
	return buf.String()
}
