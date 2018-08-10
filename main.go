package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"restful/server"
)

var (
	logFileBase string
	frontPort   int
	restPort    int
)

func main() {
	flag.StringVar(&logFileBase, "logFile", "/tmp/log_url_shortener", "The base full filename for logs.")
	flag.IntVar(&frontPort, "frontPort", 80, "The front server port.")
	flag.IntVar(&restPort, "restPort", 8080, "The RESTful API server port.")
	logFile := logFileBase + time.Now().Local().Format("2006-01-02")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		fmt.Printf("Can't open the logging file %s", logFile)
		fmt.Println(err)
		fmt.Println("Aborting")
		os.Exit(1)
	}
	log.SetOutput(f)
	server := server.NewURLServer(server.Routes)
	server.Start(frontPort, restPort)
}
