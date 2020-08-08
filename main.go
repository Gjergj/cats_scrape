package main

import (
	"fmt"
	"os"
	"runtime"
	"scrape/cats"
	"strings"

	log "github.com/sirupsen/logrus"
)

func formatFilePath(path string) string {
	arr := strings.Split(path, "/")
	return arr[len(arr)-1]
}
func initLogging(file *os.File) {
	log.SetOutput(file)
	log.SetFormatter(&log.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			// this function is required when you want to introduce your custom format.
			// In my case I wanted file and line to look like this `file="engine.go:141`
			// but f.File provides a full path along with the file name.
			// So in `formatFilePath()` function I just trimmet everything before the file name
			// and added a line number in the end
			return "", fmt.Sprintf("%s:%d", formatFilePath(f.File), f.Line)
		},
	})
	log.SetLevel(log.DebugLevel)

	log.SetReportCaller(true)
}

func main() {
	file, err := os.OpenFile("log.json", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	initLogging(file)

	catsData := cats.NewCatsUserData()
	catsData.Username = os.Getenv("CATS_VAT_NUMBER")
	catsData.Password = os.Getenv("CATS_PASSWORD")
	err = catsData.Login()
	if err != nil {
		panic(err)
	}
	err = catsData.AccountingRealTimeDebits()
	if err != nil {
		panic(err)
	}
	fmt.Println(catsData.AccountingFields.RealTimeDebitsResponse.ToJSON())
}
