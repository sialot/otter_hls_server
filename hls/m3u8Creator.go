package hls

import (
	"fmt"
	"strings"

	config "../config"
	ts "../ts"
)

// GetM3U8 M3U8文件获取
func GetM3U8(fileUri string) (string, error) {

	// m3u8 文件内容
	var resultStr = ""

	// 提取本地文件路径
	localDir, err := config.SysConfig.Get("media.localDir")

	if err != nil {
		fmt.Println("Can't find media.localDir")
	}

	// 获取ts二进制索引文件本地路径
	var binaryIndexFilePath = strings.TrimSuffix(localDir+fileUri, ".m3u8") + ".tsidx"

	// 获取ts索引对象
	tsIndex, err := ts.GetTsIndex(binaryIndexFilePath)

	if err != nil {
		return "", err
	}

	fmt.Println(tsIndex)

	resultStr += "#EXTM3U\n"

	// TODO
	resultStr += binaryIndexFilePath
	return resultStr, nil
}
