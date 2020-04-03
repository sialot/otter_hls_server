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

// TargetDuration m3u8单片最大时长
var TargetDuration int

// Init 初始化
func Init() {

	// 提取本地文件路径
	var err error

	var targetDurationStr string
	targetDurationStr, err = config.SysConfig.Get("m3u8.target_duration")

	if err != nil {
		panic(err.Error())
	}

	TargetDuration, err = strconv.Atoi(targetDurationStr)
	if err != nil {
		panic(err.Error())
	}

	Log = logger.Log
}

// GetM3U8 M3U8文件获取
func GetM3U8(m3u8FileURI string, host string, mainFlag bool) (string, error) {

	// 无后缀的基本文件路径
	var baseFileURINoSuffix = strings.TrimSuffix(strings.TrimSuffix(m3u8FileURI, ".m3u8"), ".M3U8")

	// 获取ts索引对象
	mediaFileIndex, err := ts.GetMediaFileIndex(baseFileURINoSuffix)

	if err != nil {
		Log.Error(err.Error())
		return "", err
	}
	if !mainFlag {
		return createSubM3u8(mediaFileIndex, baseFileURINoSuffix, host), nil
	}
	return createMainM3u8(mediaFileIndex, baseFileURINoSuffix, host), nil
}

// createMainM3u8 创建一级m3u8
//
// #EXTM3U
// #EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=500000
func createMainM3u8(mediaFileIndex *ts.MediaFileIndex, baseFileURINoSuffix string, host string) string {

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
	resultStr += "http://" + host + "/hls_sub/" + baseFileURINoSuffix + ".m3u8"

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
func createSubM3u8(mediaFileIndex *ts.MediaFileIndex, baseFileURINoSuffix string, host string) string {

	Log.Debug(">>> GetSubnM3u8 Start: " + baseFileURINoSuffix + ".m3u8")

	// m3u8 文件内容
	var resultStr = ""

	// #EXTM3U
	resultStr += "#EXTM3U\n"

	// #EXT-X-VERSION:4
	resultStr += "#EXT-X-VERSION:4 \n"

	// #EXT-X-TARGETDURATION:{M3U8_TARGET_DURATION}
	targetDurationStr := strconv.FormatUint(uint64(TargetDuration), 10)
	resultStr += "#EXT-X-TARGETDURATION:" + targetDurationStr + "\n"

	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXT-X-PLAYLIST-TYPE:VOD
	resultStr += "#EXT-X-MEDIA-SEQUENCE:0\n"
	resultStr += "#EXT-X-PLAYLIST-TYPE:VOD\n"

	// 获取文件列表
	videoList := GetVideoList(mediaFileIndex, float64(TargetDuration))

	var i int
	for i = 0; i < len(videoList); i++ {

		// #EXTINF:6.006,
		resultStr += "#EXTINF:" + fmt.Sprintf("%.2f", videoList[i].Duration) + "\n"

		// ./video/video_index.M3U8
		// 作为二级m3u8文件"
		sequenceStr := strconv.FormatUint(uint64(videoList[i].Sequence), 10)
		resultStr += "http://" + host + "/video/" + baseFileURINoSuffix + "_" + sequenceStr + ".ts\n"
	}

	// #EXT-X-ENDLIST
	resultStr += "#EXT-X-ENDLIST"

	Log.Debug("<<< GetSubM3u8 End")
	return resultStr
}
