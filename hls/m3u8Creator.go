package hls

import (
	"fmt"
	"strconv"
	"strings"

	config "../config"
	ts "../ts"
)

// 文件本地路径
var localDir string

// 服务器域名前缀
var serverDomainName string

// Init 初始化
func Init() {

	// 提取本地文件路径
	var err error
	localDir, err = config.SysConfig.Get("media.localDir")

	if err != nil {
		fmt.Println("Can't find media.localDir")
	}

	serverDomainName, err = config.SysConfig.Get("server.domain")

	if err != nil {
		fmt.Println("Can't find media.localDir")
	}
}

// GetM3U8 M3U8文件获取
func GetM3U8(m3u8FileURI string, mainFlag bool) (string, error) {

	// 无后缀的基本文件路径
	var baseFileURI = strings.TrimSuffix(strings.TrimSuffix(m3u8FileURI, ".m3u8"), ".M3U8")

	// 获取ts二进制索引文件本地路径
	var binaryIndexFilePath = localDir + baseFileURI + ".tsidx"

	// 获取ts索引对象
	MediaFileIndex, err := ts.GetMediaFileIndex(binaryIndexFilePath)
	if err != nil {
		return "", err
	}

	fmt.Println(MediaFileIndex)

	if !mainFlag {
		return createSubM3u8(MediaFileIndex, m3u8FileURI), nil
	}
	return createMainM3u8(MediaFileIndex, m3u8FileURI), nil
}

// createMainM3u8 创建一级m3u8
//
// #EXTM3U
// #EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=500000
func createMainM3u8(MediaFileIndex *ts.MediaFileIndex, baseFileURI string) string {

	// m3u8 文件内容
	var resultStr = ""

	// #EXTM3U
	resultStr += "#EXTM3U\n"

	// #EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=500000
	bindWidth := strconv.FormatUint(uint64(MediaFileIndex.BindWidth), 10)
	resultStr += "#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=" + bindWidth + "\n"

	// ./video/video_index.M3U8
	// 作为二级m3u8文件"
	resultStr += serverDomainName + "hls_sub/" + baseFileURI
	return resultStr
}

// createMainM3u8 创建二级m3u8
// #EXTM3U
// #EXT-X-VERSION:4
// #EXT-X-TARGETDURATION:{M3U8_TARGET_DURATION}
// #EXT-X-MEDIA-SEQUENCE:0
// #EXT-X-PLAYLIST-TYPE:VOD
// #EXTINF:6.006,
// 2000_vod_00001.ts
// #EXT-X-ENDLIST
func createSubM3u8(MediaFileIndex *ts.MediaFileIndex, baseFileURI string) string {

	// m3u8 文件内容
	var resultStr = ""

	// #EXTM3U TODO
	resultStr += "#EXTM3U\n"
	resultStr += "// TODO"
	return resultStr
}
