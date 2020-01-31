// emit.go 主要负责写文件
package gen

import (
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/goecology/generator/internal/model"
	"github.com/goecology/generator/pkg/arg"

	"github.com/flosch/pongo2"
	"github.com/pkg/errors"
	"github.com/smartwalle/pongo2render"
)

var (
	// tpls 存放tmpl文件，key是相对路径
	tpls map[string]string
	// tplDirs 存放tmpl文件夹，key是文件夹相对路径
	tplDirs map[string]bool
)

func init() {
	pongo2.RegisterFilter("lowerfirst", lwfirst)
	pongo2.RegisterFilter("upperfirst", upperfirst)
	tpls = make(map[string]string)
	tplDirs = make(map[string]bool)
}

// lwfirst 首字母小写，注意不要和go关键字冲突
func lwfirst(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.Len() <= 0 {
		return pongo2.AsValue(""), nil
	}
	t := in.String()
	r, size := utf8.DecodeRuneInString(t)
	return pongo2.AsValue(strings.ToLower(string(r)) + t[size:]), nil
}

func upperfirst(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.Len() <= 0 {
		return pongo2.AsValue(""), nil
	}
	t := in.String()
	return pongo2.AsValue(strings.Replace(t, string(t[0]), strings.ToUpper(string(t[0])), 1)), nil
}

// loadTmpl 递归地将tmpl文件加载到内存中
func loadTmpl() {
	tmplRepoDir := arg.TmplDir
	err := filepath.Walk(tmplRepoDir,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}
			if info.IsDir() && path != tmplRepoDir {
				relPath, _ := filepath.Rel(tmplRepoDir, path)
				tplDirs[relPath] = true
			}
			if err != nil {
				return err
			}
			b, e := ioutil.ReadFile(path)
			if e != nil {
				return nil
			}
			relPath, e := filepath.Rel(tmplRepoDir, path)
			if e != nil {
				return nil
			}

			tpls[relPath] = string(b)
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

// render 渲染模板
func Render(schemaTpls map[string]model.Table) {
	loadTmpl()
	camelTableNames := make(map[string]struct{})
	for tableName, _ := range schemaTpls {
		camelTableNames[snakeToCamel(tableName)] = struct{}{}
	}
	ctx := pongo2.Context{
		"camelTableNames": camelTableNames,
	}
	render(ctx, schemaTpls)
}

// getPath 替换path中的特殊变量，返回最终生成的go文件路径
func getPath(path string, tableName string) string {
	path = strings.ReplaceAll(path, "TABLE_NAME", tableName)
	path = strings.ReplaceAll(path, ".go.tmpl", ".gen.go")
	return path
}

// loadImports 载入默认imports表
func loadImports() map[string]struct{} {
	imports := make(map[string]struct{})
	for relPath := range tplDirs {
		// 过滤掉dao
		if relPath == "dao" {
			continue
		}
		imports[arg.Module+"/"+relPath] = struct{}{}
	}
	imports[arg.Module+"/pkg/mus"] = struct{}{}
	imports[arg.Model+"/mysql"] = struct{}{}
	imports[arg.Dao] = struct{}{}
	imports[arg.Model+"/trans"] = struct{}{}
	imports["go.uber.org/zap"] = struct{}{}
	imports["github.com/jinzhu/gorm"] = struct{}{}
	imports["strings"] = struct{}{}
	imports["time"] = struct{}{}
	if arg.Debug == "true" {
		fmt.Println("imports module: ", imports)
	}
	return imports
}

// render 渲染dist文件
func render(ctx pongo2.Context, schemas map[string]model.Table) {
	var render = pongo2render.NewRender(arg.TmplDir)
	for path, content := range tpls {
		var globalImports = loadImports()
		// 剔除自己所在的包，防止循环引用
		delete(globalImports, arg.Module+"/"+filepath.Dir(path))
		for tableName, schema := range schemas {
			schema.Imports = globalImports
			var hasOpenId, hasDeleteTime bool
			for _, value := range schema.Columns {
				if value.CamelName == "DeleteTime" {
					schema.Imports["time"] = struct{}{}
					hasDeleteTime = true
				}
			}
			// todo 莫名其毛有时候有time，有时候没有，估计数据串了
			schema.Imports["time"] = struct{}{}

			for _, value := range schema.Columns {
				if value.CamelName == "OpenId" && value.ColumnKey != "PRI" {
					hasOpenId = true
				}
			}

			ctx["tableName"] = tableName
			ctx["camelTableName"] = snakeToCamel(tableName)
			ctx["lcamelTableName"] = lowerFirst(snakeToCamel(tableName))
			ctx["hasOpenId"] = hasOpenId
			ctx["hasDeleteTime"] = hasDeleteTime
			ctx["imports"] = schema.Imports
			ctx["columns"] = schema.Columns
			ctx["hasPrimaryKey"] = schema.HasPrimaryKey
			ctx["camelPrimaryKey"] = schema.CamelPrimaryKey
			ctx["primaryKey"] = schema.PrimaryKey
			ctx["primaryKeyType"] = schema.PrimaryKeyType

			// 如果包含到dao，就换个目录输出
			if strings.Contains(path, "dao/") {
				buf, err := render.TemplateFromString(content).Execute(ctx)
				if err = write(filepath.Join(arg.OutDao, getPath(strings.Replace(path, "dao/", "", 1), tableName)), buf); err != nil {
					log.Panicln("[render] write err: ", err.Error(), path, tableName, buf)
					return
				}
			} else {
				buf, err := render.TemplateFromString(content).Execute(ctx)
				if err = write(filepath.Join(arg.Out, getPath(path, tableName)), buf); err != nil {
					log.Panicln("[render] write err: ", err.Error(), path, tableName, buf)
					return
				}
			}

		}
	}
}

// write 写bytes到文件
func write(filename string, buf string) (err error) {
	filePath := path.Dir(filename)
	err = createPath(filePath)
	if err != nil {
		err = errors.New("write create path " + err.Error())
		return
	}

	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		err = errors.New("write create file " + err.Error())
		return
	}

	// 格式化代码
	bts, err := format.Source([]byte(buf))
	if err != nil {
		err = errors.New("format buf error " + err.Error())
		return
	}

	err = ioutil.WriteFile(filename, bts, 0644)
	if err != nil {
		err = errors.New("write write file " + err.Error())
		return
	}

	if arg.Debug == "true" {
		fmt.Println("write file success, file name: ", filename)
	}

	return
}
