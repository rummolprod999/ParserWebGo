package main

import (
	"path/filepath"
	"os"
	"fmt"
	"time"
)

type Filelog string
type FileTemp string

var DirLog string
var DirTemp string
var FileLog Filelog
var FileTmp FileTemp

func Logging(args ...interface{}) {
	file, err := os.OpenFile(string(FileLog), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		fmt.Println("Ошибка записи в файл лога", err)
		return
	}
	fmt.Fprintf(file, "%v  ", time.Now())
	for _, v := range args {

		fmt.Fprintf(file, " %v", v)
	}
	//fmt.Fprintf(file, " %s", UrlXml)
	fmt.Fprintln(file, "")

}

func CreateLogFile() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Println()
	dirlog := fmt.Sprintf("%s/%s", dir, DirLog)
	if _, err := os.Stat(dirlog); os.IsNotExist(err) {
		err := os.MkdirAll(dirlog, 0711)

		if err != nil {
			fmt.Println("Не могу создать папку для лога")
			os.Exit(1)
		}
	}
	t := time.Now()
	ft := t.Format("2006-01-02")
	FileLog = Filelog(fmt.Sprintf("%s/log_fabrikant_%v.log", dirlog, ft))

}

