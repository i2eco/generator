package gen

import (
	"os"
	"strings"
)

// snakeToCamel 下划线转大驼峰
func snakeToCamel(name string) string {
	parts := strings.Split(name, "_")
	newName := ""
	for _, p := range parts {
		if len(p) < 1 {
			continue
		}
		newName = newName + upperFirst(p)
	}
	return newName
}

// upperFirst 首字母大写
func upperFirst(str string) string {
	return strings.Replace(str, string(str[0]), strings.ToUpper(string(str[0])), 1)
}

// lowerFirst 首字母小写
func lowerFirst(str string) string {
	return strings.Replace(str, string(str[0]), strings.ToLower(string(str[0])), 1)
}

// camelToSnake 驼峰转snake
func camelToSnake(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}

// createPath 调用os.MkdirAll递归创建文件夹
func createPath(filePath string) error {
	if !isExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		return err
	}
	return nil
}

// isExist 判断所给路径文件/文件夹是否存在(返回true是存在)
func isExist(path string) bool {
	_, err := os.Stat(path) // os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
