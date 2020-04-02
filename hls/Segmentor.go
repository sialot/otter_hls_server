package hls

import (
	"fmt"
	"strconv"
	"strings"

	errors "../errors"
	path "../path"
	ts "../ts"
)

// VideoInfo 视频文件信息
type VideoInfo struct {
	Sequence    int     // 序号
	StartOffset uint64  // 开始偏移量（字节）
	Size        uint64  // 大小（字节）
	Duration    float64 // 时长
}

// GetVideoList 计算视频列表
func GetVideoList(mediaFileIndex *ts.MediaFileIndex, targetDuration float64) []VideoInfo {
	Log.Debug("TargetDuration: " + fmt.Sprint(targetDuration))

	var videoList []VideoInfo = make([]VideoInfo, 0)
	var curSeq int = 0

	var file VideoInfo
	file.Sequence = curSeq
	file.StartOffset = 0
	file.Size = 0
	file.Duration = 0

	// 开始分片ts文件路径处理
	var i int
	for i = 0; i < len(mediaFileIndex.TimesArray); i++ {

		// 预计添加了这个时间片后的时长
		sliceDuration := mediaFileIndex.TimesArray[i].MaxTime - mediaFileIndex.TimesArray[i].MinTime
		nextDuration := file.Duration + float64(sliceDuration)

		// 累加大小
		file.Size = mediaFileIndex.TimesArray[i].StartOffset - file.StartOffset

		// 添加后超过最大限制
		if nextDuration > targetDuration {

			// 插入旧文件
			videoList = append(videoList, file)

			// 文件数增加
			curSeq++

			// 当前片为新一个文件的开始
			file.Sequence = curSeq
			file.Size = 0
			file.StartOffset = mediaFileIndex.TimesArray[i].StartOffset
			file.Duration = float64(sliceDuration)

		} else {

			// 累加时长
			file.Duration += float64(sliceDuration)
		}
	}

	// 插入最后的一片
	// 格式化时长
	file.Size = mediaFileIndex.VideoSize - file.StartOffset
	videoList = append(videoList, file)

	Log.Debug("GetVideoList, list size: " + fmt.Sprint(len(videoList)))
	return videoList
}

// GetVideoStream 获取视频流
func GetVideoStream(videoFileURI string) (*VideoInfo, string, error) {

	Log.Debug("GetVideoStream, videoFileURI:" + videoFileURI)

	if strings.Index(videoFileURI, "/") < 0 {
		Log.Error("Can't get group_name from url!")
		err := errors.NewError(errors.ErrorCodeGetIndexFailed, "GetVideoStream failed, can't get group_name from url, read index file failed")
		return nil, "", err
	}

	// 真实媒体路径
	var realMediaLocalPath string

	// 无后缀的基本文件路径
	videoFileURINoSuffix := strings.TrimSuffix(strings.TrimSuffix(videoFileURI, ".ts"), ".TS")

	// 获取视频序号
	sequenceStr := videoFileURINoSuffix[strings.LastIndex(videoFileURINoSuffix, "_")+1 : len(videoFileURINoSuffix)]

	// 获取视频分片序号
	sequence, err := strconv.Atoi(sequenceStr)
	if err != nil {
		err := errors.NewError(errors.ErrorCodeGetStreamFailed, "GetVideoStream failed, can't get fileNumber!")
		return nil, "", err
	}

	Log.Debug("GetVideoStream, sequenceStr:" + sequenceStr)

	// 组名
	groupName := videoFileURINoSuffix[0: strings.Index(videoFileURINoSuffix, "/")]

	// 视频相对路径
	mediaFileURI := videoFileURINoSuffix[strings.Index(videoFileURINoSuffix, "/") + 1: len(videoFileURINoSuffix)]

	// 真实媒体文件路径
	realMediaLocalPath = path.MediaFileFolders[groupName].LocalPath + mediaFileURI[0:strings.LastIndex(mediaFileURI, "_")] + ".ts"
	Log.Debug("GetVideoStream, realMediaLocalPath:" + realMediaLocalPath)

	// 获取ts索引对象
	baseFileURINoSuffix := videoFileURINoSuffix[0:strings.LastIndex(videoFileURINoSuffix, "_")]
	mediaFileIndex, err := ts.GetMediaFileIndex(baseFileURINoSuffix)
	if err != nil {
		return nil, "", err
	}

	// 获取文件列表
	videoList := GetVideoList(mediaFileIndex, float64(TargetDuration))

	// 找文件
	var left int = 0
	var right int = len(videoList) - 1

	for {
		if left > right {
			break
		}

		mid := (right + left) / 2
		if videoList[mid].Sequence == sequence {
			Log.Debug("Seek video info:" + fmt.Sprint(videoList[mid]))
			return &videoList[mid], realMediaLocalPath, nil
		} else if videoList[mid].Sequence < sequence {
			left = mid + 1
		} else if videoList[mid].Sequence > sequence {
			right = mid - 1
		}
	}

	err = errors.NewError(errors.ErrorCodeGetStreamFailed, "GetVideoStream failed, can't get videoFile!")
	return nil, "", err
}
