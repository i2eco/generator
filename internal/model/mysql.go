package model

import "database/sql"

type TableSchema struct {
	TableName              string
	ColumnName             string
	IsNullable             string
	DataType               string
	CharacterMaximumLength sql.NullInt64
	NumericPrecision       sql.NullInt64
	NumericScale           sql.NullInt64
	ColumnType             string
	ColumnKey              string
	Comment                string
}

type ColumnSchema struct {
	ColumnName             string
	IsNullable             string
	DataType               string
	CharacterMaximumLength sql.NullInt64
	NumericPrecision       sql.NullInt64
	NumericScale           sql.NullInt64
	ColumnType             string
	ColumnKey              string // PRI,UNI,MUL
	Comment                string
}

type Table struct {
	Imports         map[string]struct{} // imports表
	HasPrimaryKey   bool                // 是否是主键
	CamelPrimaryKey string              // 转成驼峰的主键字段名
	PrimaryKey      string              // 主键字段名
	PrimaryKeyType  string              // 主键字段类型
	Columns         []Column            // 所有字段
}

type Column struct {
	Name      string // 字段名
	CamelName string // 驼峰字段名
	Type      string // MySQL中原始数据类型
	ColumnKey string // PRI说明是主键
	Comment   string // Mysql中原始注释

	GoType    string  // Go结构体字段类型
	GoJsonTag string  // GO结构体中json标签
	GoComment Comment // 从注释中json反序列化的Comment
}

type Comment struct {
	Data string `json:"data"` // 注释内容
	Type string `json:"type"` // slice代表数组
}
