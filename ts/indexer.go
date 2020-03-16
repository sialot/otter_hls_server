package ts

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	errors "../errors"
)

// Indexer TS文件索引创建器
type Indexer struct {
	indexFilePath string            // 索引文件路径
	timesMap      map[int]TimeSlice // 时间片列表,key为秒数
	timesArray    []TimeSlice       // 时间片集合列表
	minTime       int               // 最小显示时间戳
	maxTime       int               // 最大显示时间戳
}

// TsIndex ts文件索引
type TsIndex struct {
	bindWidth  uint64      // 带宽(比特率)
	duration   int         // 总时长
	timesArray []TimeSlice // 时间片集合列表
}

// TimeSlice 以秒为单位的时间片
type TimeSlice struct {
	time        int
	startOffset uint64 // 开始偏移量
}

// GetTsIndex 获取ts文件索引
func GetTsIndex(indexFilePath string) (*TsIndex, error) {

	fmt.Println("finding indexFilePath:" + indexFilePath)

	// 定义返回值
	var pTsIndex *TsIndex 
	var err error

	// 判断是否已存在二进制索引
	if !fileExists(indexFilePath) {

		fmt.Println("tsidx file not exist,now try to build one.")

		// 索引器
		var indexer Indexer
		pTsIndex, err = indexer.createIndex(indexFilePath)

		// 创建索引失败
		if err != nil {
			err := errors.NewError(errors.ErrorCodeGetIndexFailed, "Ts index file get failed!")
			return nil, err
		}

	} else {

		fmt.Println("tsidx file exists.")

	}

	return pTsIndex, nil
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
	t.time = int(pts / 90)
	t.startOffset = offset

	indexer.timesArray = append(indexer.timesArray, t)
}

// writeIndexFile 将索引文件写入硬盘
func writeIndexFile(pTsIndex *TsIndex) error {

	// TODO
	return nil
}

// writeIndexFile 将索引文件写入硬盘
func readIndexFile(tsIndexFilePath string) (*TsIndex, error) {

	// TODO
	return nil, nil
}

// createIndex Info
func (indexer *Indexer) createIndex(idxFilePath string) (*TsIndex, error) {

	// 初始化成员变量
	indexer.minTime = -1
	indexer.maxTime = -1
	indexer.timesMap = make(map[int]TimeSlice)
	indexer.timesArray = make([]TimeSlice, 0)

	// 获取ts文件路径
	var tsFilePath = strings.TrimSuffix(idxFilePath, ".tsidx") + ".ts"

	fmt.Println("try to open ts file:" + tsFilePath)

	// 打开ts文件
	file, err := os.Open(tsFilePath)
	if err != nil {

		fmt.Println("TsFile not exist!")
		err := errors.NewError(errors.ErrorCodeDemuxFailed, "TsFile not exist!")
		return nil, err
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
				return nil, err
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
				fmt.Printf(err.Error())
				return nil, err
			}
			if pes != nil {
				indexer.feedFrame(pes.PTS, pes.PkgOffset)
			}

		}
	}

	// 创建索引结果对象
	var tsIndex TsIndex
	tsIndex.duration = (indexer.maxTime - indexer.minTime) / 1000
	tsIndex.bindWidth = getFileSize(tsFilePath) / uint64(tsIndex.duration) 
	tsIndex.timesArray = indexer.timesArray
	return &tsIndex, nil
}

// fileExists 判断所给路径文件/文件夹是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// getFileSize 获取文件大小
func getFileSize(filename string) uint64 {
	var result uint64
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		result = uint64(f.Size())
		return nil
	})
	return result
}
