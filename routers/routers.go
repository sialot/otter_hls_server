package routers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"fmt"

	strings "strings"
	config "../config"
	hls "../hls"
	logger "../log"
	ts "../ts"
	"github.com/sialot/ezlog"
)

// Log 系统日志
var Log *ezlog.Log

// M3u8Host
var M3u8Host string

// Init 初始化
func Init() {
	Log = logger.Log

	host, err := config.SysConfig.Get("m3u8host")
	if err != nil {
		panic(err.Error())
	}
	M3u8Host = host
}

// GetM3U8 M3U8文件获取
func GetM3U8(w http.ResponseWriter, r *http.Request) {

	var url = r.URL.Path
	Log.Debug(">>>>>>>>>>> Request url:" + url)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 非m3u8请求，返回404
	if !(strings.HasSuffix(url, ".m3u8") || strings.HasSuffix(url, ".M3U8")) {
		w.WriteHeader(404)
		w.Write([]byte("ERROR 404: The file requested is not exist!"))
		return
	}

	// 获取m3u8文件
	m3u8, err := hls.GetM3U8(strings.Replace(r.URL.Path, "/hls/", "", 1), M3u8Host)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("ERROR 404: The file requested is not exist!\n"))
		w.Write([]byte(err.Error()))
		return
	}

	// 返回m3u8文件内容
	w.Header().Set("Content-Type", "application/x-mpegURL;charset=UTF-8")
	w.Write([]byte(m3u8))
}

// GetVideoStream 视频文件获取
func GetVideoStream(w http.ResponseWriter, r *http.Request) {

	var url = r.URL.Path
	Log.Debug(">>>>>>>>>>> Request url:" + url)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 非m3u8请求，返回404
	if !(strings.HasSuffix(url, ".ts")) {
		w.WriteHeader(404)
		w.Write([]byte("ERROR 404: Unsurported file type!"))
		return
	}

	// 获取视频文件信息
	videoInfo, realMediaLocalPath, err := hls.GetVideoStream(strings.Replace(r.URL.Path, "/video/", "", 1))

	// 打开文件
	file, err := os.Open(realMediaLocalPath)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("ERROR 404: The file requested is not exist!\n"))
		w.Write([]byte(err.Error()))
		file.Close()
		return
	}

	// 获取文件状态
	fileStat, err := file.Stat()
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("ERROR 404: The file requested is not exist!\n"))
		w.Write([]byte(err.Error()))
		file.Close()
		return
	}

	_, fileName := filepath.Split(r.URL.Path)
	w.Header().Set("Last-Modified", fileStat.ModTime().Format(http.TimeFormat))
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "video/MP2T")
	w.Header().Set("Content-Length", strconv.FormatUint(videoInfo.Size, 10))

	file.Seek(int64(videoInfo.StartOffset), 0)

	// Stream data out !
	var tranceSize uint64

	buf := make([]byte, min(1024*1024*10, fileStat.Size()))
	n := 0

	for err == nil {

		n, err = file.Read(buf)

		// 读取文件失败
		if err != nil {
			if err != io.EOF {
				w.WriteHeader(404)
				w.Write([]byte("ERROR 404: Unsurported file type!\n"))
				w.Write([]byte(err.Error()))
			}
			break
		}

		if tranceSize+uint64(n) > videoInfo.Size {
			w.Write(buf[0 : videoInfo.Size-tranceSize])
			tranceSize += videoInfo.Size - tranceSize
			break
		} else {
			w.Write(buf[0:n])
			tranceSize += uint64(n)
		}
	}

	file.Close()
}

// CreateIndex 主动创建索引
func CreateIndex(w http.ResponseWriter, r *http.Request) {

	var url = r.URL.Path
	Log.Debug(">>>>>>>>>>> Request url:" + r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	// 非m3u8请求，返回404
	if !(strings.HasSuffix(url, ".ts") || strings.HasSuffix(url, ".Ts")) {
		w.Write([]byte("{\"code\":\"-1\",\"msg\":\"Unsurported file type!\"}"))
		return
	}

	m3u8FileURI := strings.Replace(r.URL.Path, "/createIndex/", "", 1)
	baseFileURINoSuffix := strings.TrimSuffix(strings.TrimSuffix(m3u8FileURI, ".ts"), ".Ts")
	indexFileURI := baseFileURINoSuffix + ".tsidx"

	// 获取m3u8文件
	err := ts.CreateMediaFileIndex(indexFileURI)
	if err != nil {
		w.Write([]byte("{\"code\":\"-1\",\"msg\":\"Create index failed! ,erros: " + err.Error() + "\"}"))
		return
	}

	w.Write([]byte("{\"code\":\"1\",\"msg\":\"\"}"))
}

// GetProcessInfo 获取索引进度
func GetProcessInfo(w http.ResponseWriter, r *http.Request) {

	var resultJson string

	resultJson += "{\"code\":\"1\","
	resultJson += "\"list\":["

	processList := ts.GetProgressInfo()

	for i, info := range processList {

		resultJson += "{"
		resultJson += "\"filePath\":\"" + info.FilePath + "\","
		resultJson += "\"fileSize\":\"" + formatFileSize(info.FileSize) + "\","
		resultJson += "\"progress\":" + strconv.Itoa(info.Progress)
		resultJson += "}"

		if i < (len(processList) - 1) {
			resultJson += ","
		}
	}
	resultJson += "]}"
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(resultJson))
}

// formatFileSize 字节的单位转换 保留两位小数
func formatFileSize(fileSize int64) (size string) {
	if fileSize < 1024 {
	   return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
	   return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
	   return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
	   return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
	   return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else {
	   return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
 }

func min(x int64, y int64) int64 {
	if x < y {
		return x
	}
	return y
}
