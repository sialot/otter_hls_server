package ts

import "fmt"

// Indexer TS文件索引创建器
type Indexer struct {
	timesMap   map[int]TimeSlice // 时间片列表,key为秒数
	timesArray []TimeSlice       // 时间片集合列表
	minTime    int               // 最小显示时间戳
	maxTime    int               // 最大显示时间戳
}

// TimeSlice 以秒为单位的时间片
type TimeSlice struct {
	time        int
	startOffset uint64 // 开始偏移量
	endOffset   uint64 // 结束偏移量
}

// Init 初始化索引创建器
func (indexer *Indexer) Init() {
	indexer.minTime = -1
	indexer.maxTime = -1
	indexer.timesMap = make(map[int]TimeSlice)
	indexer.timesArray = make([]TimeSlice, 0)
}

// FeedFrame 输入帧数据
func (indexer *Indexer) FeedFrame(pts int64, offset uint64) {

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
	t.endOffset = 0

	indexer.timesArray = append(indexer.timesArray, t)
}

// CreateIndex Info
func (indexer *Indexer) CreateIndex() {
	fmt.Println(indexer.maxTime)
	fmt.Println(indexer.minTime)
}
