package ts

import (
	"strconv"
	"sync"
)

type ProcessInfo struct {
	Progress int
	FileSize int64
	FilePath string
}

// 存放当前处理的文件路径
var processMap map[string]*ProcessInfo
var keySlice []string

// 操作锁
var mutex sync.Mutex

// init
func init() {
	processMap = make(map[string]*ProcessInfo)
	keySlice = make([]string, 0)
}

// startProcess 开始出楼
func startProcess(filePath string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := processMap[filePath]; ok {
		return true
	}

	var p ProcessInfo
	p.FilePath = filePath
	p.FileSize = -1
	p.Progress = -1

	processMap[filePath] = &p
	keySlice = append(keySlice, filePath)
	Log.Debug("[ActionNote]StartProcess:" + filePath)
	return false
}

// updateProcess 更新进度
func updateProcess(filePath string, curOffset int64, fileSize int64) {
	var progress = int(curOffset * 100 / fileSize)
	if progress > processMap[filePath].Progress {
		processMap[filePath].FileSize = fileSize
		processMap[filePath].Progress = progress
		Log.Debug("[ActionNote]updateProcess:" + filePath + ", progress:" + strconv.Itoa(processMap[filePath].Progress) + "%")
	}
}

// finishProcess 结束处理
func finishProcess(filePath string) {
	mutex.Lock()
	Log.Debug("[ActionNote]finishProcess:" + filePath)
	defer mutex.Unlock()
	delete(processMap, filePath)

	var index int = -1
	for i, key := range keySlice {
		if key == filePath {
			index = i
			break
		}
	}
	if index >= 0 {
		keySlice = append(keySlice[:index], keySlice[index+1:]...)
	}
}

// GetProgressInfo 查询所有当前索引进度
func GetProgressInfo() []*ProcessInfo {
	var processList = make([]*ProcessInfo, 0)
	for _, key := range keySlice {
		processList = append(processList, processMap[key])
	}
	return processList
}
