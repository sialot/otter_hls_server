package hls

import (
	"fmt"
	"strings"
	"strconv"
	"path/filepath"

	config "../config"
	common "../common"
	errors "../errors"
	ts "../ts"
)

// 文件本地路径
var localDir string

// 一级m3u8文件
const M3U8_TYPE_MAIN = 0

// 二级m3u8文件
const M3U8_TYPE_SUB = 1

// 单分片最大时长
const M3U8_TARGET_DURATION = 10

// Init 初始化
func Init(){

	// 提取本地文件路径
	var err error
	localDir, err = config.SysConfig.Get("media.localDir")

	if err != nil {
		fmt.Println("Can't find media.localDir")
	}
}

// GetM3U8 M3U8文件获取
func GetM3U8(m3u8FileUri string) (string, error) {

	// 无后缀的基本文件路径
	fileType, realFileUri, err :=  analysisUri(m3u8FileUri)
	if err != nil {
		return "", err
	}

	// 获取ts二进制索引文件本地路径
	var binaryIndexFilePath = localDir + realFileUri + ".tsidx"

	// 获取ts索引对象
	tsIndex, err := ts.GetTsIndex(binaryIndexFilePath)
	if err != nil {
		return "", err
	}

	fmt.Println(tsIndex)

	switch fileType {
	case M3U8_TYPE_MAIN:
		return createMainM3u8(tsIndex, realFileUri), nil

	case M3U8_TYPE_SUB:
		return createSubM3u8(tsIndex, realFileUri, m3u8FileUri), nil

	default:
		err := errors.NewError(errors.ErrorCodeGetIndexFailed, "M3U8 file get failed!")
		return "", err
	}
}

// analysisUri 分析文件路径
// returns m3u8Type int{一级、二级 m3u8} , baseFileUri string {视频文件基本uri，不带文件后缀}
func analysisUri(m3u8FileUri string) (int, string, error) {

	// 无后缀的基本文件路径
	var baseFileUri =  strings.TrimSuffix(strings.TrimSuffix(m3u8FileUri, ".m3u8"),".M3U8")

	// 计算ts文件路径
	var tsFilePath = localDir + baseFileUri + ".ts"

	// 指定的ts文件存在
	if common.FileExists(tsFilePath) {
		return M3U8_TYPE_MAIN, baseFileUri, nil

	// 有可能是二级索引文件，例如：1_sub_0_0.M3U8
	} else {

		// 从结尾寻找 “_0”，去掉 “_0”
		for {

			// 如果文件不以以 “_0” 结尾，终止循环
			if !strings.HasSuffix(baseFileUri, "_0") {
				break
			} else {
				baseFileUri = strings.TrimSuffix(baseFileUri, "_0")
			}
		}

		// 去掉 _0 后，判断是否符合 _sub 后缀，如果是，有可能是二级m3u8
		if strings.HasSuffix(baseFileUri, "_sub") {
			
			// 获得最终baseFileUri
			baseFileUri = strings.TrimSuffix(baseFileUri, "_sub")

			// 这时候返回的文件可能还是不存在，交由indexer处理
			return M3U8_TYPE_SUB, baseFileUri, nil

		} else {
			err := errors.NewError(errors.ErrorCodeGetIndexFailed, "M3U8 file get failed!")
			return M3U8_TYPE_SUB, "", err
		}
	}
}

// createMainM3u8 创建一级m3u8
//
// #EXTM3U
// #EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=500000
func createMainM3u8(tsIndex *ts.TsIndex, baseFileUri string) string {

	// m3u8 文件内容
	var resultStr = ""

	// #EXTM3U
	resultStr += "#EXTM3U\n"
	
	// #EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=500000
	bindWidth := strconv.FormatUint(uint64(tsIndex.BindWidth), 10)
	resultStr += "#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=" + bindWidth + "\n"

	// video/video_index.M3U8
	// 通过当前文件名后添加 “_sub” 作为二级m3u8文件
	// 循环判断新文件是否被ts文件占用，如果有添加后缀
	var i int = 0
	baseSubM3u8FileUri := baseFileUri + "_sub"
	for {
		if !common.FileExists(localDir + baseSubM3u8FileUri + ".ts") {
			break
		}
		baseSubM3u8FileUri += "_" + strconv.FormatUint(uint64(i), 10)
	}

	_, fileName := filepath.Split(baseSubM3u8FileUri + ".M3U8")

	resultStr += fileName
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
func createSubM3u8(tsIndex *ts.TsIndex, realFileUri string, m3u8Uri string) string {

	fmt.Println(tsIndex)

	// m3u8 文件内容
	var resultStr = ""

	// #EXTM3U TODO
	resultStr += "#EXTM3U\n"
	resultStr += "// TODO"
	return resultStr
}