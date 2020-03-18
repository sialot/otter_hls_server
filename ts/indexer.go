package ts

import (
	"fmt"
	"io"
	"os"
	"strings"
	"bytes"
	"math"
	"encoding/binary"

	common "../common"
	errors "../errors"
)

// 索引速度优化、区分文件的索引锁
// 索引计算更新时间，对比ts文件防止索引失效

// Indexer TS文件索引创建器
type Indexer struct {
	indexFilePath string            // 索引文件路径
	frameArray    []TimeSlice       // 帧时间片集合列表
	minTime       int               // 最小显示时间戳
	maxTime       int               // 最大显示时间戳
}

// TsIndex ts文件索引
type TsIndex struct {
	BindWidth  uint32      // 带宽(比特率)
	Duration   uint32      // 总时长
	TimesArray []TimeSlice // 时间片集合列表
}

// TimeSlice 以秒为单位的时间片
type TimeSlice struct {
	time        float32
	startOffset uint64 // 开始偏移量
}

// GetTsIndex 获取ts文件索引
func GetTsIndex(indexFilePath string) (*TsIndex, error) {
	
	var tsIndex *TsIndex
	var err error

	// 判断是否已存在二进制索引
	if !common.FileExists(indexFilePath) {

		// fmt.Println("tsidx file not exist,now try to build one.")

		// 索引器 TODO 需要加锁
		var indexer Indexer
		err := indexer.createIndex(indexFilePath)

		// 创建索引失败
		if err != nil {
			err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file get failed!")
			return nil, err
		}

	} else {

		// fmt.Println("tsidx file exists.")

		tsIndex, err = readIndexFile(indexFilePath)
		if err != nil {
			return nil, err
		}
	}
	return tsIndex, nil
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

	var t TimeSlice
	t.time = float32(pts / 90)
	t.startOffset = offset

	indexer.frameArray = append(indexer.frameArray, t)
}

// writeIndexFile 将索引文件写入硬盘
//
// 索引文件构成 
// HEADER[0x12F(12bit),type(4bit)],CONTENT(96bit),0xffff(16bit)
// type = 0 : CONTENT[bindWidth(32bit),zero(64bit)]]
// type = 1 : CONTENT[duration(32bit),zero(64bit)]]
// type = 2 : CONTENT[time(32bit),startOffset(64bit)]]
//
// bindWidth    |码率(32bit)
// duration		|时长秒(32bit)
// time			|分片时间（单位秒）(32bit)
// startOffset	|分片偏移量(64bit)
func writeIndexFile(pTsIndex *TsIndex, idxFilePath string) error {

	var file *os.File
	var err error

	// 已存在索引清空文件
	if common.FileExists(idxFilePath) {
		file, err = os.Open(idxFilePath)
		if err != nil {
			return err
		} else {

			// 清空文件
			file.Truncate(0)
		}

		defer file.Close()

	// 文件不存在创建文件
	}  else {
		file, err = os.Create(idxFilePath)
		defer file.Close()
	}

	var bin_buf bytes.Buffer

	// 写入带宽
	binary.Write(&bin_buf, binary.BigEndian, uint16(0x12F0))
	binary.Write(&bin_buf, binary.BigEndian, pTsIndex.BindWidth)
	binary.Write(&bin_buf, binary.BigEndian, uint64(0x00))
	binary.Write(&bin_buf, binary.BigEndian, uint16(0xFFFF))

	// 写入时长
	binary.Write(&bin_buf, binary.BigEndian, uint16(0x12F1))
	binary.Write(&bin_buf, binary.BigEndian, pTsIndex.Duration)
	binary.Write(&bin_buf, binary.BigEndian, uint64(0x00))
	binary.Write(&bin_buf, binary.BigEndian, uint16(0xFFFF))

	// 写入时间片信息
	var i int
	for i = 0; i < len(pTsIndex.TimesArray) ; i++ {
		slice := pTsIndex.TimesArray[i]
		binary.Write(&bin_buf, binary.BigEndian, uint16(0x12F2))
		bits := math.Float32bits(slice.time)
		binary.Write(&bin_buf, binary.BigEndian, bits)
		binary.Write(&bin_buf, binary.BigEndian, slice.startOffset)
		binary.Write(&bin_buf, binary.BigEndian, uint16(0xFFFF))
	}

	_, err = file.Write(bin_buf.Bytes())
	if err != nil {
		fmt.Println("write index file failed, ", err.Error())
		return err
	}

	return nil
}

// readIndexFile 从磁盘读取索引文件
func readIndexFile(idxFilePath string) (*TsIndex, error) {

	var file *os.File
	var err error
	var tsIndex TsIndex
	tsIndex.TimesArray = make([]TimeSlice, 0)

	// 打开文件
	file, err = os.Open(idxFilePath)
	if err != nil {
		err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file read failed!")
		return nil, err
	}
	defer file.Close()

	// 预加载包字节
	data := make([]byte, 16)

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
		var syncData uint8 = data[0]
		if syncData != 0x12 {
			err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file read failed!")
			return nil, err
		}

		// 检验结束位
		var endSyncData uint16 = uint16(data[14]) << 8 | uint16(data[15])
		if endSyncData != 0xFFFF {
			err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file read failed!")
			return nil, err
		}

		var dataType uint8 = data[1] & 0x0F
		switch dataType {
		case 0: 
			tsIndex.BindWidth =  uint32(data[2]) << 24 | uint32(data[3]) << 16 | uint32(data[4]) << 8 | uint32(data[5])
		case 1:	
			tsIndex.Duration =  uint32(data[2]) << 24 | uint32(data[3]) << 16 | uint32(data[4]) << 8 | uint32(data[5])
		case 2:
			var slice TimeSlice
			slice.time = math.Float32frombits(uint32(data[2]) << 24 | uint32(data[3]) << 16 | uint32(data[4]) << 8 | uint32(data[5]))
			slice.startOffset = uint64(data[6]) << 56 | uint64(data[7]) << 48 | uint64(data[9]) << 40 | uint64(data[9]) << 32|
				uint64(data[10]) << 24 | uint64(data[11]) << 16 | uint64(data[12]) << 8 | uint64(data[13])

			tsIndex.TimesArray = append(tsIndex.TimesArray, slice)
		}
	}

	return &tsIndex, nil
}

// createIndex Info
func (indexer *Indexer) createIndex(idxFilePath string) error {

	// 初始化成员变量
	indexer.minTime = -1
	indexer.maxTime = -1
	indexer.frameArray = make([]TimeSlice, 0)

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
	var tsIndex TsIndex
	tsIndex.Duration = uint32(indexer.maxTime - indexer.minTime) / 1000
	tsIndex.BindWidth = uint32(common.GetFileSize(tsFilePath) / uint64(tsIndex.Duration)) 
	tsIndex.TimesArray = make([]TimeSlice, 0)
	
	// 整理切片时间,time单位为秒，改为每秒一个切片
	var i int
	var cursecond int = -1
	var second float32
	for i = 0; i < len(indexer.frameArray); i++ {

		second = (indexer.frameArray[i].time - float32(indexer.minTime)) / 1000

		if int(second) > cursecond {
			cursecond = int(second)
			indexer.frameArray[i].time = second
			tsIndex.TimesArray = append(tsIndex.TimesArray, indexer.frameArray[i])
		}
	}

	// 写索引文件
	fileWriteErr := writeIndexFile(&tsIndex, idxFilePath)
	if fileWriteErr != nil {
		return fileWriteErr
	}

	fmt.Printf("Indexing finish\n")

	return nil
}