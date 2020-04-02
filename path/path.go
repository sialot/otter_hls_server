package path

import (
	"strconv"
	config "../config"
	logger "../log"
	"github.com/sialot/ezlog"
)

// Folder 媒体本地文件夹爱
type Folder struct {
	LocalPath string // 媒体本地文件夹
	GroupName   string // 映射到url的文件夹
}

// Log 系统日志
var Log *ezlog.Log

// IndexFileFolder 索引文件存放目录
var IndexFileFolder string

// 媒体文件存放目录集合
var MediaFileFolders map[string]Folder

// 初始化方法
func init() {
	MediaFileFolders = make(map[string]Folder)
}

// LoadPath 初始化方法
func LoadPath() {

	Log = logger.Log

	Log.Info("Loading index_file_folder")
	var err error
	IndexFileFolder, err = config.SysConfig.Get("path.index_file_folder")
	if err != nil {
		panic(err.Error())
	}
	Log.Info("Load index_file_folder complete! index_file_folder: "+ IndexFileFolder)
	
	Log.Info("Loading media_file_folders")

	count, err := config.SysConfig.Count("path.media_file_folders")

	if err != nil {
		panic(err.Error())
	}
	var i int
	for i = 0; i < count; i++ {

		localPath, err := config.SysConfig.Get("path.media_file_folders[" + strconv.Itoa(i) +"].local_path")
		if err != nil {
			panic(err.Error())
		}
		
		groupName, err := config.SysConfig.Get("path.media_file_folders[" + strconv.Itoa(i) +"].group_name")
		if err != nil {
			panic(err.Error())
		}

		var f Folder
		f.LocalPath = localPath
		f.GroupName = groupName

		MediaFileFolders[groupName] = f

		Log.Info("Watch localPath: " + localPath + ", group_name:" + groupName)
	}

	Log.Info("Load media_file_folders complete!")
}