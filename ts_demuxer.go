package main
import(
	"os"
	"fmt"
	"io"
	"log"

)
const TS_PKG_SIZE int = 188;

/**
 * 处理文件
 */
func ProcessFile(file *os.File)  {
	
	fmt.Printf("ts_demuxer.processFile start ! \n")
	
	tsPkgBytes := make([]byte,TS_PKG_SIZE)
	var i int
	for {
		i++
		_,err := file.Read(tsPkgBytes)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}

		DemuxPkg(tsPkgBytes)
	}
	fmt.Printf("ts_demuxer.processFile finish ! \n")
}

func DemuxPkg(tsPkg []byte){

}