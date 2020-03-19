package routers

import (
	"net/http"

	strings "strings"

	hls "../hls"
)

// Welcome 欢迎页
func Welcome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("OtterHLSServer!"))
}

// GetMainM3U8 M3U8文件获取
func GetMainM3U8(w http.ResponseWriter, r *http.Request) {

	var url = r.URL.Path

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
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(m3u8))
}

// GetSubM3U8 M3U8文件获取
func GetSubM3U8(w http.ResponseWriter, r *http.Request) {

	var url = r.URL.Path

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
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(m3u8))
}

// GetVideo 视频文件获取
func GetVideo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("GetVideo！"))
}
