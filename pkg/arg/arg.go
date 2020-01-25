package arg

import (
	"encoding/json"
)

var (
	Database  string
	MySQL     string
	Out       string
	SshTunnel string
	Table     string
	Module    string
	Model     string // 数据库等描述文件
	TmplDir   string
	Debug     string
	TimeType  int // 1为int64, 2为time.Time
)

type CmdDt struct {
	Data json.RawMessage `json:"data"`
}
