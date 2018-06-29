package main

import (
	"fmt"
	"log"
	"os"
)

var logFile string = "/home/lc22073/go/src/restful/log/log_url_shortener"

func main() {
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		fmt.Printf("Can't open the logging file %s", logFile)
		fmt.Println(err)
		fmt.Println("Aborting")
		os.Exit(1)
	}
	log.SetOutput(f)
	server := NewURLServer(routes)
	server.Start(8080, 8081)
}
