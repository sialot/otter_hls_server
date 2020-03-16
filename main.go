package main

import (
	"fmt"
	"io"
	"log"
	"os"

	ts "./ts"
)

func main() {

	file, err := os.Open("/Volumes/user/2.ts")
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	var d ts.Demuxer
	d.Init()

	var indexer ts.Indexer
	indexer.Init()

	fmt.Printf("ts_demuxer.processFile start ! \n")

	// 预加载ts包字节 切片
	preLoadData := make([]byte, ts.TsPkgSize*ts.TsReloadNum)

	// 取ts文件
	for {
		_, err := file.Read(preLoadData)
		//fmt.Printf("LoadData, size：%d \n", len(preLoadData))

		// 读取文件失败
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}

		// 解封装
		var i int
		for i = 0; i < ts.TsReloadNum; i++ {
			var pKgBuf []byte = preLoadData[i*188 : (i*188 + 188)]
			pes, err := d.DemuxPkg(pKgBuf)

			if err != nil {
				fmt.Println(err)
			}
			if pes != nil {
				//fmt.Printf("PTS:%d,PTS:%d,DTS:%d,PkgOffset:%d \n",	pes.PID, pes.PTS, pes.DTS, pes.PkgOffset)

				indexer.FeedFrame(pes.PTS, pes.PkgNum)

	// 声明路由
 	mux := http.NewServeMux()
	mux.HandleFunc("/", routers.Welcome)
 	mux.HandleFunc("/hls/", routers.GetM3U8)
	mux.HandleFunc("/hls/video/", routers.GetVideo)

		}

	}

	indexer.CreateIndex()

	fmt.Printf("ts_demuxer.processFile finish ! \n")

	fmt.Printf("OK! \n")
}
