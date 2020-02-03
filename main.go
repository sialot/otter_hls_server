package main

import (
	"fmt"
	"log"
	"os"

	ts "./ts"
)

func main() {

	file, err := os.Open("/Volumes/user/1.ts")
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	var d ts.TsDemuxer
	d.Init()
	d.ProcessFile(file)
	fmt.Printf("OK! \n")
}
