package util

import (
	"os"
	"path/filepath"
)

// FileExists 判断所给路径文件/文件夹是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// GetFileSize 获取文件大小
func GetFileSize(filename string) uint64 {
	var result uint64
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		result = uint64(f.Size())
		return nil
	})
	return result
}
