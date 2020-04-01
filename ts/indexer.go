package ts

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	common "../common"
	config "../config"
	errors "../errors"
	logger "../log"
	"github.com/sialot/ezlog"
)

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
	VideoSize  uint64      // 视频文件大小
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

// MediaRootPath 文件本地路径
var MediaRootPath string

// IndexRootPath 索引文件 本地路径
var IndexRootPath string

// Log 系统日志
var Log *ezlog.Log

// VERSION 索引版本号
const VERSION uint8 = 0

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

	Log = logger.Log
}

// GetMediaFileIndex 获取ts文件索引
func GetMediaFileIndex(indexFileURI string) (*MediaFileIndex, error) {

	var mediaFileIndex *MediaFileIndex
	var err error

	// 尝试读取索引文件
	mediaFileIndex, err = readIndexFile(indexFileURI)

	// 读取索引文件失败，重新创建索引
	if err != nil {
		Log.Debug("readIndexFile file failed: " + err.Error())
		Log.Debug("now try to build new one.")

		// 索引器 TODO 需要加锁
		var indexer Indexer
		mediaFileIndex, err = indexer.createIndex(indexFileURI)

		// 创建索引失败
		if err != nil {
			Log.Error("CreateIndex file failed: " + err.Error())
			return nil, err
		}

		return mediaFileIndex, nil
	}

	Log.Debug("Read tsidx success!")

	return mediaFileIndex, nil
}

// CreateMediaFileIndex 获取ts文件索引
func CreateMediaFileIndex(indexFileURI string) error {

	Log.Debug("CreateMediaFileIndex indexFileURI:" + indexFileURI)

	var err error

	// 尝试读取索引文件
	_, err = readIndexFile(indexFileURI)

	// 读取索引文件失败，重新创建索引
	if err != nil {
		Log.Debug("readIndexFile file failed: " + err.Error())
		Log.Debug("now try to build new one.")

		var indexer Indexer
		_, err = indexer.createIndex(indexFileURI)

		// 创建索引失败
		if err != nil {
			Log.Error("CreateIndex file failed: " + err.Error())
			return err
		}
	}

	Log.Debug("CreateMediaFileIndex success!")
	return nil
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
// type = 0 时表示索引基本信息
// PAYLOAD[version(8bit), bindWidth(32bit),duration(32bit),reserve(56bit)]
// type = 1 时表示视频文件基本信息
// PAYLOAD[video_size(64bit), reserve(64bit)]
// type = 2 时表示帧数据
// PAYLOAD[mintime(32bit),maxtime(32bit),startOffset(64bit)]]
//
// version：索引版本
// bindWidth: 媒体码率
// duration: 总时长
// reserve: 保留位，默认0
// mintime		|最小帧时间（单位秒）(32bit)
// maxtime		|最大帧时间（单位秒）(32bit)
// startOffset	|分片偏移量(64bit)
func writeIndexFile(pMediaFileIndex *MediaFileIndex, indexFileURI string) error {

	var err error

	// 索引文件路径
	var indexFilePath = IndexRootPath + indexFileURI

	// 父目录
	indexFileDirPath, _ := filepath.Split(indexFilePath)

	// 父文件夹不存在，创建文件夹
	if !common.FileExists(indexFileDirPath) {

		Log.Debug("index dir not exist, try to create one")

		err = os.MkdirAll(indexFileDirPath, os.ModePerm)
		if err != nil {
			Log.Error("create index dir failed:" + err.Error())
			return err
		}
	}

	// 索引文件
	var file *os.File
	file, err = os.OpenFile(indexFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		Log.Error("openfile failed:" + err.Error())
		return err
	}

	// 清空文件
	file.Truncate(0)
	defer file.Close()

	var binBuf bytes.Buffer

	// ========= 写入索引文件基本信息 START=========
	// 头信息 HEADER[0xf(4bit),type=0(4bit)]
	binary.Write(&binBuf, binary.BigEndian, uint8(0xF0))

	// 载荷 PAYLOAD[version=1(8bit), bindWidth(32bit),duration(32bit),reserve(56bit)]
	binary.Write(&binBuf, binary.BigEndian, uint8(VERSION))
	binary.Write(&binBuf, binary.BigEndian, pMediaFileIndex.BindWidth)
	binary.Write(&binBuf, binary.BigEndian, pMediaFileIndex.Duration)

	// 保留位
	binary.Write(&binBuf, binary.BigEndian, uint32(0))
	binary.Write(&binBuf, binary.BigEndian, uint16(0))
	binary.Write(&binBuf, binary.BigEndian, uint8(0))

	// ENDFLAG
	binary.Write(&binBuf, binary.BigEndian, uint8(0xFF))

	// ========= 写入索引文件基本信息 END=========

	// ========= 写入视频文件信息 START=========
	// 头信息 HEADER[0xf(4bit),type=0(4bit)]
	binary.Write(&binBuf, binary.BigEndian, uint8(0xF1))

	// 载荷 PAYLOAD[video_size(64bit), reserve(64bit)]
	binary.Write(&binBuf, binary.BigEndian, pMediaFileIndex.VideoSize)

	// 保留位
	binary.Write(&binBuf, binary.BigEndian, uint64(0))

	// ENDFLAG
	binary.Write(&binBuf, binary.BigEndian, uint8(0xFF))

	// ========= 写入帧数据信息 END =========

	// ========= 写入帧数据信息 START=========
	var i int
	for i = 0; i < len(pMediaFileIndex.TimesArray); i++ {

		// 获得时间片
		slice := pMediaFileIndex.TimesArray[i]

		// 头信息 HEADER[0xf(4bit),type=1(4bit)]
		binary.Write(&binBuf, binary.BigEndian, uint8(0xF2))

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
		Log.Error("write index file failed" + err.Error())
		return err
	}

	return nil
}

// readIndexFile 从磁盘读取索引文件
func readIndexFile(indexFileURI string) (*MediaFileIndex, error) {

	var file *os.File
	var err error
	var MediaFileIndex MediaFileIndex
	MediaFileIndex.TimesArray = make([]TimeSlice, 0)

	// 打开文件
	file, err = os.Open(IndexRootPath + indexFileURI)
	if err != nil {
		err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file read failed, can't open file!")
		return nil, err
	}
	defer file.Close()

	// 获取索引文件大小和修改时间
	fi, err := file.Stat()
	if err != nil {
		err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file getSize failed!")
		return nil, err
	}

	// 大小为零认为是错误
	if fi.Size() == 0 {
		err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file read failed, empty file!")
		return nil, err
	}

	// 获取ts文件修改时间
	var tsFilePath = MediaRootPath + strings.TrimSuffix(indexFileURI, ".tsidx") + ".ts"
	tsFile, err := os.Open(tsFilePath)
	if err != nil {
		err := errors.NewError(errors.ErrorCodeGetIndexFailed, "TsFile not exist!")
		return nil, err
	}

	tsfi, err := tsFile.Stat()
	if err != nil {
		err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file getSize failed!")
		return nil, err
	}

	// 媒体文件被修改
	if tsfi.ModTime().Unix() > fi.ModTime().Unix() {
		err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file is out of data!")
		return nil, err
	}

	defer tsFile.Close()

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
			err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file syncData error!")
			return nil, err
		}

		// 检验结束位
		var endSyncData uint16 = uint16(data[17])
		if endSyncData != 0xFF {
			err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file endSyncData error!")
			return nil, err
		}

		var dataType uint8 = data[0] & 0x0F
		switch dataType {
		case 0:

			version := data[1]
			if version != VERSION {
				err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file version error!")
				return nil, err
			}

			MediaFileIndex.BindWidth = uint32(data[2])<<24 | uint32(data[3])<<16 | uint32(data[4])<<8 | uint32(data[5])
			MediaFileIndex.Duration = uint32(data[6])<<24 | uint32(data[7])<<16 | uint32(data[8])<<8 | uint32(data[9])
		case 1:

			MediaFileIndex.VideoSize = uint64(data[1])<<56 | uint64(data[2])<<48 | uint64(data[3])<<40 | uint64(data[4])<<32 |
				uint64(data[5])<<24 | uint64(data[6])<<16 | uint64(data[7])<<8 | uint64(data[8])

		case 2:

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
func (indexer *Indexer) createIndex(indexFileURI string) (*MediaFileIndex, error) {

	// 标记当前处理正在被别人抢占
	var waitProcess bool = false

	// 自旋锁，同文件处理需要等待
	for {

		// 注册开始处理文件
		needWait := startProcess(indexFileURI)

		// 当前处理被抢占
		if !waitProcess && needWait {
			Log.Debug("Someone is creating index file, need wait. indexFileURI:" + indexFileURI)
			waitProcess = true
		}

		// 结束等待
		if !needWait {
			break
		}
	}

	// 结束处理
	defer finishProcess(indexFileURI)

	// 轮到我处理，再次尝试读索引
	if waitProcess {
		Log.Debug("Wait stop! Retry to read tsidx:" + indexFileURI)

		// 再次尝试读取索引文件
		pMediaFileIndex, err := readIndexFile(IndexRootPath + indexFileURI)

		// 获取成功
		if err == nil {
			Log.Debug("Read tsidx success")
			return pMediaFileIndex, nil
		}

		Log.Debug("Retry failed :" + err.Error())
	}

	Log.Debug("Start create indexfile")

	// 初始化成员变量
	indexer.minTime = -1
	indexer.maxTime = -1
	indexer.frameArray = make([]Frame, 0)

	// 打开ts文件
	var tsFilePath = MediaRootPath + strings.TrimSuffix(indexFileURI, ".tsidx") + ".ts"

	Log.Debug("open ts file:" + tsFilePath)

	file, err := os.Open(tsFilePath)
	if err != nil {
		err := errors.NewError(errors.ErrorCodeDemuxFailed, "TsFile not exist!")
		return nil, err
	}

	defer file.Close()

	// 获取文件状态
	fileStat, err := file.Stat()
	if err != nil {
		err := errors.NewError(errors.ErrorCodeDemuxFailed, "TsFile not exist!")
		return nil, err
	}

	// 预加载ts包字节 切片
	preLoadData := make([]byte, min(int64(TsPkgSize*TsReloadNum), fileStat.Size()))

	// 创建解封装器
	var d Demuxer

	// 初始化解封装器
	d.Init()

	// 取ts文件
	for {
		_, err := file.Read(preLoadData)

		// 读取文件失败
		if err != nil {
			if err != io.EOF {
				err := errors.NewError(errors.ErrorCodeDemuxFailed, "TsFile read failed!")
				return nil, err
			}
			break
		}

		// 解封装
		var i int
		for i = 0; i < TsReloadNum; i++ {

			if (len(preLoadData) - i*188) < 188 {
				Log.Debug("Wrong ts file length!")
				break
			}

			var pKgBuf []byte = preLoadData[i*188 : (i*188 + 188)]
			pes, err := d.DemuxPkg(pKgBuf)

			// 解封装失败 TODO
			if err != nil {
				Log.Error("Demux ts file failed: " + err.Error())
				return nil, err
			}
			if pes != nil {
				indexer.feedFrame(pes.PTS, pes.PkgOffset)
			}

		}
	}

	// 索引对象
	var mediaFileIndex MediaFileIndex
	mediaFileIndex.VideoSize = common.GetFileSize(tsFilePath)
	mediaFileIndex.Duration = uint32(indexer.maxTime-indexer.minTime) / 1000
	mediaFileIndex.BindWidth = uint32(mediaFileIndex.VideoSize / uint64(mediaFileIndex.Duration))
	mediaFileIndex.TimesArray = make([]TimeSlice, 0)

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
			mediaFileIndex.TimesArray = append(mediaFileIndex.TimesArray, slice)

			// 重置分片信息
			lastSliceMaxTime = slice.MaxTime
			sliceMaxTime = -1
			newSlice = true
		}
	}

	// 最后一个分片
	mediaFileIndex.TimesArray = append(mediaFileIndex.TimesArray, slice)

	// 写索引文件
	fileWriteErr := writeIndexFile(&mediaFileIndex, indexFileURI)
	if fileWriteErr != nil {
		return nil, fileWriteErr
	}

	return &mediaFileIndex, nil
}

// min
func min(x int64, y int64) int64 {
	if x < y {
		return x
	}
	return y
}
