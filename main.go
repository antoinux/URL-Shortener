package restful

import (
	"fmt"
	"log"
	"os"
	"time"

	"restful/server"
)

var logFileBase string = "/home/lc22073/go/src/restful/log/log_url_shortener"

func main() {
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
	server.Start(8080, 8081)
}
