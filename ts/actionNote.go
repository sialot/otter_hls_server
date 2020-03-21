package ts

import "sync"

// 存放当前处理的文件路径
var processMap map[string]bool

// 操作锁
var mutex sync.Mutex

// init
func init() {
	processMap = make(map[string]bool)
}

// startProcess 开始出楼
func startProcess(filePath string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	if processMap[filePath] {
		return true
	}

	processMap[filePath] = true
	Log.Debug("[ActionNote]StartProcess:" + filePath)
	return false
}

// finishProcess 结束处理
func finishProcess(filePath string) {
	mutex.Lock()
	Log.Debug("[ActionNote]finishProcess:" + filePath)
	defer mutex.Unlock()
	delete(processMap, filePath)
}
