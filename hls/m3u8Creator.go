package hls

import (
	"fmt"
	"strconv"
	"strings"

	config "../config"
	logger "../log"
	ts "../ts"
	"github.com/sialot/ezlog"
)

// Log 系统日志
var Log *ezlog.Log

// MediaRootPath 文件本地路径
var MediaRootPath string

// 索引文件 本地路径
var IndexRootPath string

// 服务器域名前缀
var serverDomainName string

// m3u8单片最大时长
var targetDuration int

// Init 初始化
func Init() {

	// 提取本地文件路径
	var err error
	MediaRootPath, err = config.SysConfig.Get("media.mediaRootPath")

	if err != nil {
		panic(err.Error())
	}

	IndexRootPath, err = config.SysConfig.Get("media.indexRootPath")

	if err != nil {
		panic(err.Error())
	}

	serverDomainName, err = config.SysConfig.Get("server.domainName")

	if err != nil {
		panic(err.Error())
	}

	var targetDurationStr string
	targetDurationStr, err = config.SysConfig.Get("m3u8.target_duration")

	if err != nil {
		panic(err.Error())
	}

	targetDuration, err = strconv.Atoi(targetDurationStr)
	if err != nil {
		panic(err.Error())
	}

	Log = logger.Log
}

// GetM3U8 M3U8文件获取
func GetM3U8(m3u8FileURI string, mainFlag bool) (string, error) {

	// 无后缀的基本文件路径
	var baseFileURINoSuffix = strings.TrimSuffix(strings.TrimSuffix(m3u8FileURI, ".m3u8"), ".M3U8")

	// 获取ts二进制索引文件本地路径
	var indexFileURI = baseFileURINoSuffix + ".tsidx"

	// 获取ts索引对象
	mediaFileIndex, err := ts.GetMediaFileIndex(indexFileURI)

	if err != nil {
		Log.Error(err.Error())
		return "", err
	}
	if !mainFlag {
		return createSubM3u8(mediaFileIndex, baseFileURINoSuffix), nil
	}
	return createMainM3u8(mediaFileIndex, baseFileURINoSuffix), nil
}

// createMainM3u8 创建一级m3u8
//
// #EXTM3U
// #EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=500000
func createMainM3u8(mediaFileIndex *ts.MediaFileIndex, baseFileURINoSuffix string) string {

	Log.Debug(">>> GetMainM3u8 Start: " + baseFileURINoSuffix + ".m3u8")

	// m3u8 文件内容
	var resultStr = ""

	// #EXTM3U
	resultStr += "#EXTM3U\n"

	// #EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=500000
	bindWidth := strconv.FormatUint(uint64(mediaFileIndex.BindWidth), 10)
	resultStr += "#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=" + bindWidth + "\n"

	// ./hls_sub/video_index.M3U8
	// 作为二级m3u8文件"
	resultStr += serverDomainName + "hls_sub/" + baseFileURINoSuffix + ".m3u8"

	Log.Debug("<<<GetMainM3u8 End, Result: \n" + resultStr)
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
func createSubM3u8(mediaFileIndex *ts.MediaFileIndex, baseFileURINoSuffix string) string {

	Log.Debug(">>> GetSubnM3u8 Start: " + baseFileURINoSuffix + ".m3u8")

	// m3u8 文件内容
	var resultStr = ""

	// #EXTM3U
	resultStr += "#EXTM3U\n"

	// #EXT-X-VERSION:4
	resultStr += "#EXT-X-VERSION:4 \n"

	// #EXT-X-TARGETDURATION:{M3U8_TARGET_DURATION}
	targetDurationStr := strconv.FormatUint(uint64(targetDuration), 10)
	resultStr += "#EXT-X-TARGETDURATION:" + targetDurationStr + "\n"

	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXT-X-PLAYLIST-TYPE:VOD
	resultStr += "#EXT-X-MEDIA-SEQUENCE:0\n"
	resultStr += "#EXT-X-PLAYLIST-TYPE:VOD\n"

	// 获取文件列表
	videoList := GetVideoList(mediaFileIndex, float64(targetDuration))

	var i int
	for i = 0; i < len(videoList); i++ {

		// #EXTINF:6.006,
		resultStr += "#EXTINF:" + fmt.Sprintf("%.2f", videoList[i].Duration) + "\n"

		// ./video/video_index.M3U8
		// 作为二级m3u8文件"
		sequenceStr := strconv.FormatUint(uint64(videoList[i].Sequence), 10)
		resultStr += serverDomainName + "video/" + baseFileURINoSuffix + "_" + sequenceStr + ".ts\n"
	}

	// #EXT-X-ENDLIST
	resultStr += "#EXT-X-ENDLIST"

	Log.Debug("<<< GetSubM3u8 End")
	return resultStr
}
