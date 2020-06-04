package main

import (
	"fmt"
	"log"
)

func checkErrorFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var (
	Green  = Color("\033[1;32m%s\033[0m")
	Yellow = Color("\033[1;33m%s\033[0m")
)

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}
