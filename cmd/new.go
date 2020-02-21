package cmd

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/i2eco/generator/internal/gen"
	"github.com/i2eco/generator/pkg/arg"
	"github.com/spf13/cobra"
	"log"
)

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new project from mysql database",
	Run:   newProject,
}

func init() {
	newCmd.PersistentFlags().StringVar(&arg.Database, "db", "", `指定数据库名`)
	newCmd.PersistentFlags().StringVar(&arg.MySQL, "mysql",
		"i2eco:i2eco@mysql+tcp(127.0.0.1:3306)/information_schema", `指定存储(MySQL等)地址`)
	newCmd.PersistentFlags().StringVarP(&arg.Out, "out", "o", "./dist", `指定输出目录`)
	newCmd.PersistentFlags().StringVarP(&arg.SshTunnel, "ssh", "s", "", `开启ssh隧道`)
	newCmd.PersistentFlags().StringVarP(&arg.Table, "table", "t", "%%", `指定表名`)
	newCmd.PersistentFlags().StringVarP(&arg.Module, "module", "M", "", `指定项目module`)
	newCmd.PersistentFlags().StringVar(&arg.Model, "model", "", `指定项目model`)
	newCmd.PersistentFlags().StringVar(&arg.Dao, "dao", "", `生成文件地址dao`)
	newCmd.PersistentFlags().StringVar(&arg.OutDao, "outdao", "", `生成文件地址dao`)
	newCmd.PersistentFlags().StringVarP(&arg.TmplDir, "tmpl-dir", "T", "tmpl", `指定渲染模板目录`)
	newCmd.PersistentFlags().StringVar(&arg.Debug, "debug", "false", `调试信息`)
	// newCmd.PersistentFlags().IntVarP(&arg.TimeType, "time-type", "time", 1, `指定时间类型`)
	RootCmd.AddCommand(newCmd)
}

func newProject(cmd *cobra.Command, args []string) {
	fmt.Println(" arg.Database------>", arg.Database)
	// 根据数据库解析mysql的table schema
	tableSchemas, err := gen.GetTableSchemas(arg.MySQL, arg.Database, arg.Table)
	if err != nil {
		log.Panic("[GetTableSchemas] getSchema fail", err.Error())
		return
	}

	// 将解析出来的schema数组转换为map格式
	schemaTpls := gen.GetSchemaTpls(tableSchemas)
	gen.Render(schemaTpls)
}
