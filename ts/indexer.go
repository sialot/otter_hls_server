package ts

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	common "../common"
	errors "../errors"
)

// 索引速度优化、区分文件的索引锁
// 索引计算更新时间，对比ts文件防止索引失效

// Indexer TS文件索引创建器
type Indexer struct {
	indexFilePath string  // 索引文件路径
	frameArray    []Frame // 帧时间片集合列表
	minTime       int     // 最小显示时间戳
	maxTime       int     // 最大显示时间戳
}

// Frame 以秒为单位的时间片
type Frame struct {
	Time        float32 // 最小时间
	StartOffset uint64  // 开始偏移量
}

// MediaFileIndex ts文件索引
type MediaFileIndex struct {
	BindWidth  uint32      // 带宽(比特率)
	Duration   uint32      // 总时长
	TimesArray []TimeSlice // 时间片集合列表
}

// TimeSlice 以秒为单位的时间片
type TimeSlice struct {
	MinTime     float32 // 最小时间
	MaxTime     float32 // 最大时间
	StartOffset uint64  // 开始偏移量
}

// VERSION 索引版本号
const VERSION uint8 = 1

// GetMediaFileIndex 获取ts文件索引
func GetMediaFileIndex(indexFilePath string) (*MediaFileIndex, error) {

	var MediaFileIndex *MediaFileIndex
	var err error

	// 判断是否已存在二进制索引
	if !common.FileExists(indexFilePath) {

		fmt.Println("tsidx file not exist,now try to build one.")

		// 索引器 TODO 需要加锁
		var indexer Indexer
		err := indexer.createIndex(indexFilePath)

		// 创建索引失败
		if err != nil {
			err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file get failed!")
			return nil, err
		}
	}

	MediaFileIndex, err = readIndexFile(indexFilePath)
	if err != nil {
		return nil, err
	}

	return MediaFileIndex, nil
}

// feedFrame 输入帧数据
func (indexer *Indexer) feedFrame(pts int64, offset uint64) {

	if indexer.minTime < 0 {
		indexer.minTime = int(pts / 90)
	} else if indexer.minTime > int(pts/90) {
		indexer.minTime = int(pts / 90)
	}

	if indexer.maxTime < 0 {
		indexer.maxTime = int(pts / 90)
	} else if indexer.maxTime < int(pts/90) {
		indexer.maxTime = int(pts / 90)
	}

	var f Frame
	f.Time = float32(pts / 90)
	f.StartOffset = offset

	indexer.frameArray = append(indexer.frameArray, f)
}

// writeIndexFile 将索引文件写入硬盘
//
// 索引文件构成
// 每个包 144bit
// HEADER[0xf(4bit),type(4bit)],PAYLOAD(128bit),ENDFLAG[0xff(8bit)]
//
// type = 0 时表示基本索引信息
// PAYLOAD[version(8bit), bindWidth(32bit),duration(32bit),reserve(56bit)]
// type = 1 时表示帧数据
// PAYLOAD[mintime(32bit),maxtime(32bit),startOffset(64bit)]]
//
// version：索引版本
// bindWidth: 媒体码率
// duration: 总时长
// reserve: 保留位，默认0
// mintime		|最小帧时间（单位秒）(32bit)
// maxtime		|最大帧时间（单位秒）(32bit)
// startOffset	|分片偏移量(64bit)
func writeIndexFile(pMediaFileIndex *MediaFileIndex, idxFilePath string) error {

	var file *os.File
	var err error

	// 已存在索引清空文件
	if common.FileExists(idxFilePath) {
		file, err = os.Open(idxFilePath)
		if err != nil {
			return err
		}

		// 清空文件
		file.Truncate(0)

		defer file.Close()

		// 文件不存在创建文件
	} else {
		file, err = os.Create(idxFilePath)
		defer file.Close()
	}

	var binBuf bytes.Buffer

	// ========= 写入基础文件信息 START=========
	// 头信息 HEADER[0xf(4bit),type=0(4bit)]
	binary.Write(&binBuf, binary.BigEndian, uint8(0xF0))

	// 载荷 PAYLOAD[version=1(8bit), bindWidth(32bit),duration(32bit),reserve(56bit)]
	binary.Write(&binBuf, binary.BigEndian, uint8(0x1))
	binary.Write(&binBuf, binary.BigEndian, pMediaFileIndex.BindWidth)
	binary.Write(&binBuf, binary.BigEndian, pMediaFileIndex.Duration)

	// 保留位
	binary.Write(&binBuf, binary.BigEndian, uint32(0))
	binary.Write(&binBuf, binary.BigEndian, uint16(0))
	binary.Write(&binBuf, binary.BigEndian, uint8(0))

	// ENDFLAG
	binary.Write(&binBuf, binary.BigEndian, uint8(0xFF))

	// ========= 写入基础文件信息 END=========

	// ========= 写入帧数据信息 START=========
	var i int
	for i = 0; i < len(pMediaFileIndex.TimesArray); i++ {

		// 获得时间片
		slice := pMediaFileIndex.TimesArray[i]

		// 头信息 HEADER[0xf(4bit),type=1(4bit)]
		binary.Write(&binBuf, binary.BigEndian, uint8(0xF1))

		// 载荷 PAYLOAD[mintime(32bit),maxtime(32bit),startOffset(64bit)]]
		minTimeBits := math.Float32bits(slice.MinTime)
		binary.Write(&binBuf, binary.BigEndian, minTimeBits)
		maxTimeBits := math.Float32bits(slice.MaxTime)
		binary.Write(&binBuf, binary.BigEndian, maxTimeBits)
		binary.Write(&binBuf, binary.BigEndian, slice.StartOffset)

		// ENDFLAG
		binary.Write(&binBuf, binary.BigEndian, uint8(0xFF))
	}

	// ========= 写入帧数据信息 END =========
	_, err = file.Write(binBuf.Bytes())
	if err != nil {
		fmt.Println("write index file failed, ", err.Error())
		return err
	}

	return nil
}

// readIndexFile 从磁盘读取索引文件
func readIndexFile(idxFilePath string) (*MediaFileIndex, error) {

	var file *os.File
	var err error
	var MediaFileIndex MediaFileIndex
	MediaFileIndex.TimesArray = make([]TimeSlice, 0)

	// 打开文件
	file, err = os.Open(idxFilePath)
	if err != nil {
		err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file read failed!")
		return nil, err
	}
	defer file.Close()

	// 预加载包字节
	data := make([]byte, 18)

	// 取文件
	for {
		_, err := file.Read(data)

		// 读取文件失败
		if err != nil {
			if err != io.EOF {
				err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file read failed!")
				return nil, err
			}
			break
		}

		// 校验同步位
		var syncData uint8 = data[0] >> 4
		if syncData != 0x0f {
			err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file read failed!")
			return nil, err
		}

		// 检验结束位
		var endSyncData uint16 = uint16(data[17])
		if endSyncData != 0xFF {
			err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file read failed!")
			return nil, err
		}

		var dataType uint8 = data[0] & 0x0F
		switch dataType {
		case 0:

			version := data[1]
			if version != VERSION {
				err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file read failed!")
				return nil, err
			}

			MediaFileIndex.BindWidth = uint32(data[2])<<24 | uint32(data[3])<<16 | uint32(data[4])<<8 | uint32(data[5])
			MediaFileIndex.Duration = uint32(data[6])<<24 | uint32(data[7])<<16 | uint32(data[8])<<8 | uint32(data[9])
		case 1:

			var slice TimeSlice
			slice.MinTime = math.Float32frombits(uint32(data[1])<<24 | uint32(data[2])<<16 | uint32(data[3])<<8 | uint32(data[4]))
			slice.MaxTime = math.Float32frombits(uint32(data[5])<<24 | uint32(data[6])<<16 | uint32(data[7])<<8 | uint32(data[8]))
			slice.StartOffset = uint64(data[9])<<56 | uint64(data[10])<<48 | uint64(data[11])<<40 | uint64(data[12])<<32 |
				uint64(data[13])<<24 | uint64(data[14])<<16 | uint64(data[15])<<8 | uint64(data[16])

			MediaFileIndex.TimesArray = append(MediaFileIndex.TimesArray, slice)
		}
	}

	return &MediaFileIndex, nil
}

// createIndex Info
func (indexer *Indexer) createIndex(idxFilePath string) error {

	// 初始化成员变量
	indexer.minTime = -1
	indexer.maxTime = -1
	indexer.frameArray = make([]Frame, 0)

	// 打开ts文件
	var tsFilePath = strings.TrimSuffix(idxFilePath, ".tsidx") + ".ts"
	file, err := os.Open(tsFilePath)
	if err != nil {
		err := errors.NewError(errors.ErrorCodeDemuxFailed, "TsFile not exist!")
		return err
	}

	defer file.Close()

	// 预加载ts包字节 切片
	preLoadData := make([]byte, TsPkgSize*TsReloadNum)

	// 创建解封装器
	var d Demuxer

	// 初始化解封装器
	d.Init()

	fmt.Printf("Demuxing\n")

	// 取ts文件
	for {
		_, err := file.Read(preLoadData)

		// 读取文件失败
		if err != nil {
			if err != io.EOF {
				err := errors.NewError(errors.ErrorCodeDemuxFailed, "TsFile read failed!")
				return err
			}
			break
		}

		// 解封装
		var i int
		for i = 0; i < TsReloadNum; i++ {
			var pKgBuf []byte = preLoadData[i*188 : (i*188 + 188)]
			pes, err := d.DemuxPkg(pKgBuf)

			// 解封装失败 TODO
			if err != nil {
				return err
			}
			if pes != nil {
				indexer.feedFrame(pes.PTS, pes.PkgOffset)
			}

		}
	}
	fmt.Printf("Demux finish\n")

	fmt.Printf("Indexing\n")

	// 索引对象
	var MediaFileIndex MediaFileIndex
	MediaFileIndex.Duration = uint32(indexer.maxTime-indexer.minTime) / 1000
	MediaFileIndex.BindWidth = uint32(common.GetFileSize(tsFilePath) / uint64(MediaFileIndex.Duration))
	MediaFileIndex.TimesArray = make([]TimeSlice, 0)

	// 整理切片时间,time单位为秒，改为每秒一个切片
	var i int

	// 帧真实时长
	var newSlice bool = true
	var lastSliceMaxTime float32 = 0
	var sliceOffset uint64
	var sliceMaxTime float32 = 0
	var slice TimeSlice
	for i = 0; i < len(indexer.frameArray); i++ {

		// offset
		if newSlice {
			sliceOffset = indexer.frameArray[i].StartOffset
			slice.MinTime = lastSliceMaxTime
			slice.StartOffset = sliceOffset
			newSlice = false
		}

		slice.MaxTime = sliceMaxTime

		// 当前帧的真实时间
		curFrameTime := (indexer.frameArray[i].Time - float32(indexer.minTime)) / 1000
		if sliceMaxTime < curFrameTime {
			sliceMaxTime = curFrameTime
		}

		// 计算预计时长
		nextSliceDuration := sliceMaxTime - lastSliceMaxTime

		// 分片时长超过一秒了
		if nextSliceDuration > 1 {

			// 插入分片
			MediaFileIndex.TimesArray = append(MediaFileIndex.TimesArray, slice)

			// 重置分片信息
			lastSliceMaxTime = slice.MaxTime
			sliceMaxTime = -1
			newSlice = true
		}
	}

	// 最后一个分片
	MediaFileIndex.TimesArray = append(MediaFileIndex.TimesArray, slice)

	// 写索引文件
	fileWriteErr := writeIndexFile(&MediaFileIndex, idxFilePath)
	if fileWriteErr != nil {
		return fileWriteErr
	}

	fmt.Printf("Indexing finish\n")

	return nil
}
