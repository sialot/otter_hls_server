package routers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	strings "strings"

	hls "../hls"
	logger "../log"
	"github.com/sialot/ezlog"
)

// Log 系统日志
var Log *ezlog.Log

// Init 初始化
func Init() {
	Log = logger.Log
}

// Welcome 欢迎页
func Welcome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("OtterHLSServer!"))
}

// GetMainM3U8 M3U8文件获取
func GetMainM3U8(w http.ResponseWriter, r *http.Request) {

	var url = r.URL.Path
	Log.Debug(">>>>>>>>>>> Request url:" + url)

	// 非m3u8请求，返回404
	if !(strings.HasSuffix(url, ".m3u8") || strings.HasSuffix(url, ".M3U8")) {
		w.WriteHeader(404)
		w.Write([]byte("ERROR 404: The file requested is not exist!"))
		return
	}

	// 获取m3u8文件
	m3u8, err := hls.GetM3U8(strings.Replace(r.URL.Path, "/hls/", "", 1), true)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("ERROR 404: The file requested is not exist!"))
		return
	}

	// 返回m3u8文件内容
	w.Header().Set("Content-Type", "application/x-mpegURL")
	w.Write([]byte(m3u8))
}

// GetSubM3U8 M3U8文件获取
func GetSubM3U8(w http.ResponseWriter, r *http.Request) {

	var url = r.URL.Path
	Log.Debug(">>>>>>>>>>> Request url:" + url)

	// 非m3u8请求，返回404
	if !(strings.HasSuffix(url, ".m3u8") || strings.HasSuffix(url, ".M3U8")) {
		w.WriteHeader(404)
		w.Write([]byte("ERROR 404: The file requested is not exist!"))
		return
	}

	// 获取m3u8文件
	m3u8, err := hls.GetM3U8(strings.Replace(r.URL.Path, "/hls_sub/", "", 1), false)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("ERROR 404: The file requested is not exist!"))
		return
	}

	// 返回m3u8文件内容
	w.Header().Set("Content-Type", "application/x-mpegURL")
	w.Write([]byte(m3u8))
}

// GetVideoStream 视频文件获取
func GetVideoStream(w http.ResponseWriter, r *http.Request) {

	var url = r.URL.Path
	Log.Debug(">>>>>>>>>>> Request url:" + url)

	// 非m3u8请求，返回404
	if !(strings.HasSuffix(url, ".ts")) {
		w.WriteHeader(404)
		w.Write([]byte("ERROR 404: unsupport file type!"))
		return
	}

	// 获取视频文件信息
	videoInfo, realTsFileURINoSuffix, err := hls.GetVideoStream(strings.Replace(r.URL.Path, "/video/", "", 1))

	// 打开文件
	file, err := os.Open(hls.MediaRootPath + realTsFileURINoSuffix + ".ts")
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("ERROR 404: The file requested is not exist!"))
		file.Close()
		return
	}

	// 获取文件状态
	fileStat, err := file.Stat()
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("ERROR 404: The file requested is not exist!"))
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
				w.Write([]byte("ERROR 404: unsupport file type!"))
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

func min(x int64, y int64) int64 {
	if x < y {
		return x
	}
	return y
}
