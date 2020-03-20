package hls

import (
	"fmt"
	"strings"
	"strconv"

	ts "../ts"
	errors "../errors"
)

// 视频文件信息
type VideoFile struct {
	Sequence    int     // 序号
	StartOffset uint64  // 开始偏移量（字节）
	Size        uint64  // 大小（字节）
	Duration    float64 // 时长
}

// GetVideoList 计算视频列表
func GetVideoList(mediaFileIndex *ts.MediaFileIndex, targetDuration float64) []VideoFile {
	
	var videoList []VideoFile = make([]VideoFile, 0)
	var curSeq int = 0

	var file VideoFile
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

	return videoList
}

// 获取视频流
func GetVideoStream(videoFileURI string) (*VideoFile, error){

	// 无后缀的基本文件路径
	baseVideoFileURI := strings.TrimSuffix(strings.TrimSuffix(videoFileURI, ".ts"), ".TS")
	sequenceStr := baseVideoFileURI[strings.LastIndex(baseVideoFileURI, "_") + 1 : len(baseVideoFileURI)]
	
	// 获取视频分片序号
	sequence, err := strconv.Atoi(sequenceStr)
	if err!= nil {
		err := errors.NewError(errors.ErrorCodeGetStreamFailed, "GetVideoStream failed, can't get fileNumber!")
		return nil, err
	}
	fmt.Println(sequence)

	// 无后缀的基本文件路径
	baseFileURI := baseVideoFileURI[0 : strings.LastIndex(baseVideoFileURI, "_")]

	// 获取ts二进制索引文件本地路径
	var binaryIndexFilePath = localDir + baseFileURI + ".tsidx"
	fmt.Println(binaryIndexFilePath)

	// 获取ts索引对象
	mediaFileIndex, err := ts.GetMediaFileIndex(binaryIndexFilePath)
	if err != nil {
		return nil, err
	}
	
	// 获取文件列表
	videoList := GetVideoList(mediaFileIndex, float64(targetDuration))

	// 找文件
	var i int
	for i=0; i<len(videoList); i++ {
		if videoList[i].Sequence == sequence {
			return &videoList[i], nil
		}
	}

	err = errors.NewError(errors.ErrorCodeGetStreamFailed, "GetVideoStream failed, can't get videoFile!")
	return nil, err
}