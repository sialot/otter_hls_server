package main

import (
	"fmt"
	"os"
	"log"
)

func main()  {

	file,err := os.Open("/Volumes/user/2.ts")
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	ProcessFile(file);
	fmt.Printf("OK! \n")
}