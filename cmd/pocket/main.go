package main

import (
	"flag"
	"fmt"
)

var version = "UNKNOWN"

func main() {
	v := flag.Bool("version", false, "")
	flag.Parse()
	if *v {
		fmt.Println(version)
	}
}