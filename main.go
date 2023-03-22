package main

import (
	"latale_tool/ldt"
	"log"
	"os"
)

func main() {
	l := &ldt.LDT{}
	err := l.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
}
