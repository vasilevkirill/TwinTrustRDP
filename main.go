package main

import "log"

func main() {
	checkErrorFatal(Run())

}
func checkErrorFatal(err error) {
	if err == nil {
		return
	}
	log.Fatalln(err.Error())
}
