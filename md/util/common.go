// 通用工具方法
package util

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Pre-compiled regex for whitespace removal
var blankRegex = regexp.MustCompile(`\s+`)

// 去除字符串空格或制表符
func RemoveBlank(str string) string {
	if str == "" {
		return ""
	}
	return blankRegex.ReplaceAllString(str, "")
}

// 字符串长度
func StringLength(str string) int {
	return utf8.RuneCountInString(str)
}

// 补全路径后的/
func PathCompletion(path string) string {
	if path == "" {
		path = "./"
	} else if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

// 判断文件是否存在，返回error则不确定
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// 获取文件后缀
func FileExt(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		ext = strings.ToLower(ext)
	}
	return ext
}

// UTF8转GBK
func UTF82GBK(src string) ([]byte, error) {
	GB18030 := simplifiedchinese.All[0]
	return io.ReadAll(transform.NewReader(bytes.NewReader([]byte(src)), GB18030.NewEncoder()))
}

// 字符串排序（GBK）
func StringSort(str1, str2 string) bool {
	a, _ := UTF82GBK(strings.ToLower(str1))
	b, _ := UTF82GBK(strings.ToLower(str2))
	bLen := len(b)
	for idx, chr := range a {
		if idx > bLen-1 {
			return false
		}
		if chr != b[idx] {
			return chr < b[idx]
		}
	}
	return true
}
