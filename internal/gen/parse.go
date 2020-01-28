// parse.go 主要负责解析table
package gen

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/goecology/generator/internal/model"
)

func GetSchemaTpls(tableSchemas []model.TableSchema) (resp map[string]model.Table) {
	resp = make(map[string]model.Table)
	for _, value := range tableSchemas {
		if _, ok := resp[value.TableName]; !ok {
			resp[value.TableName] = model.Table{
				Columns:         make([]model.Column, 0),
				Imports:         make(map[string]struct{}),
				HasPrimaryKey:   false,
				CamelPrimaryKey: "",
			}
		}
		schema := model.ColumnSchema{
			ColumnName:             snakeToCamel(value.ColumnName),
			IsNullable:             value.IsNullable,
			DataType:               value.DataType,
			CharacterMaximumLength: value.CharacterMaximumLength,
			NumericPrecision:       value.NumericPrecision,
			NumericScale:           value.NumericScale,
			ColumnType:             value.ColumnType,
			ColumnKey:              value.ColumnKey,
			Comment:                value.Comment,
		}

		schemaTpl := resp[value.TableName]
		if schema.ColumnKey == "PRI" {
			schemaTpl.HasPrimaryKey = true
			schemaTpl.CamelPrimaryKey = snakeToCamel(schema.ColumnName)
			schemaTpl.PrimaryKey = value.ColumnName
			gotype, _, _ := getGoType(schema, value.TableName)
			schemaTpl.PrimaryKeyType = gotype
		}
		if schema.ColumnKey != "" {
			//log.Printf("[GetSchemaTpls] schema %s:%s\n", schema.ColumnKey, schema.ColumnName)
		}

		gotype, impt, err := getGoType(schema, value.TableName)
		if err != nil {
			log.Println("[GetSchemaTpls] getSchemaMap fail", err.Error())
			return
		}
		if impt != "" {
			resp[value.TableName].Imports[impt] = struct{}{}
		}
		if schema.DataType == "json" {
			resp[value.TableName].Imports["database/sql/driver"] = struct{}{}
			resp[value.TableName].Imports["encoding/json"] = struct{}{}
		}

		tplInfo := model.Column{
			CamelName: snakeToCamel(value.ColumnName),
			Name:      value.ColumnName,
			Type:      schema.DataType,
			GoType:    gotype,
			ColumnKey: value.ColumnKey,
			GoJsonTag: value.ColumnName, // lowerFirst(snakeToCamel(value.CamelName)),
			Comment:   value.Comment,
		}
		cmt := model.Comment{}
		if err = json.Unmarshal([]byte(value.Comment), &cmt); err == nil {
			tplInfo.GoComment = cmt
		}
		schemaTpl.Columns = append(schemaTpl.Columns, tplInfo)
		resp[value.TableName] = schemaTpl
	}
	return
}

func getGoType(col model.ColumnSchema, tableName string) (string, string, error) {
	requiredImport := ""
	var gt = ""
	switch col.DataType {
	case "char", "varchar", "enum", "set", "text", "longtext", "mediumtext", "tinytext":
		if col.IsNullable == "YES" {
			gt = "sql.NullString"
		} else {
			gt = "string"
		}
	case "blob", "mediumblob", "longblob", "varbinary", "binary":
		gt = "[]byte"
	case "date", "time", "datetime", "timestamp":
		gt, requiredImport = "time.Time", "time"
		if strings.Contains(col.ColumnName, "DeletedAt") {
			gt = "*time.Time"
		}
	case "tinyint", "smallint", "int", "mediumint":
		if col.IsNullable == "YES" {
			gt = "sql.NullInt64"
		} else {
			if strings.Contains(col.ColumnName, "Time") ||
				strings.Contains(col.ColumnName, "Date") ||
				strings.Contains(col.ColumnName, "UpdatedAt") ||
				strings.Contains(col.ColumnName, "CreatedAt") {
				gt = "int64"
			} else {
				gt = "int"
			}
		}
	case "bit", "bigint":
		if col.IsNullable == "YES" {
			gt = "sql.NullInt64"
		} else {
			gt = "int64"
		}
	case "float", "decimal", "double":
		if col.IsNullable == "YES" {
			gt = "sql.NullFloat64"
		} else {
			gt = "float64"
		}
	case "json":
		// 如果是json类型，那么创建一个新的类型，类型名为转成驼峰后的列名
		gt = strings.Title(snakeToCamel(tableName + col.ColumnName + "Json"))
	}
	if gt == "" {
		return "", "", errors.New("No compatible datatype (" + col.DataType + ", CamelName: " + col.ColumnName + ")  found")
	}
	return gt, requiredImport, nil
}
