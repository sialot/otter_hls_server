package ts

import "fmt"

// Indexer TS文件索引创建器
type Indexer struct {
	timesMap map[int]*TimeSlice // 时间片列表,key为秒数
	minTime  int                // 最小显示时间戳
	maxTime  int                // 最大显示时间戳
}

// TimeSlice 以秒为单位的时间片
type TimeSlice struct {
	time   int
	pkgNum int // 开始偏移量
}

// Init 初始化索引创建器
func (indexer *Indexer) Init() {
	indexer.minTime = -1
	indexer.maxTime = -1
	indexer.timesMap = make(map[int]*TimeSlice)
}

// FeedFrame 输入帧数据
func (indexer *Indexer) FeedFrame(pts int64, pkgNum int) {

	// 记录早和最晚的时间
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

	// 向map中添加时间片
	timeInMap := indexer.timesMap[int(pts/90)/1000]

	if timeInMap == nil {
		var t *TimeSlice = &TimeSlice{}
		t.time = int(pts / 90)
		t.pkgNum = pkgNum
		indexer.timesMap[t.time/1000] = t

	} else {

		// 向前推最早偏移量
		if timeInMap.pkgNum > pkgNum {
			timeInMap.pkgNum = pkgNum
		}
	}
}

// CreateIndex Info
func (indexer *Indexer) CreateIndex() {
	var duration int = int(indexer.maxTime-indexer.minTime) / 1000
	var i int
	for i = 0; i < duration; i++ {
		t := indexer.timesMap[int(indexer.minTime/1000)+i]

		fmt.Printf("time:%d, offset:%d \n", i, t.pkgNum*188)
	}

}
