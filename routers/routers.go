package routers

import (
	"net/http"
)

// Welcome 欢迎页
func Welcome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write( []byte("OtterHLSServer Welcome!"))
}

// Hls M3U8文件获取
func Hls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write( []byte("OtterHLSServer:HLS >>> " + r.URL.Path + " | "+ r.URL.Host))
}

