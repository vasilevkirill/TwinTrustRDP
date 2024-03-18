package main

import (
	"log"
	"os"
)

func fileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func closeFileForce(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Println(err)
	}
}
