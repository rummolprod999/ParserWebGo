package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Filelog string

var DirLog string
var DirTemp string
var FileLog Filelog

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
	switch A {
	case X5Group:
		FileLog = Filelog(fmt.Sprintf("%s/log_X5Group_%v.log", dirlog, ft))
	case Dixy:
		FileLog = Filelog(fmt.Sprintf("%s/log_Dixy_%v.log", dirlog, ft))
	}

}

func CreateTempDir() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	dirtemp := fmt.Sprintf("%s/%s", dir, DirTemp)
	if _, err := os.Stat(dirtemp); os.IsNotExist(err) {
		err := os.MkdirAll(dirtemp, 0711)

		if err != nil {
			fmt.Println("Не могу создать папку для временных файлов")
			os.Exit(1)
		}
	} else {
		err = os.RemoveAll(dirtemp)
		if err != nil {
			fmt.Println("Не могу удалить папку для временных файлов")
		}
		err := os.MkdirAll(dirtemp, 0711)
		if err != nil {
			fmt.Println("Не могу создать папку для временных файлов")
			os.Exit(1)
		}
	}
}

func CreateEnv() {
	switch A {
	case X5Group:
		DirLog = LogX5Group
		DirTemp = TempX5Group
	case Dixy:
		DirLog = LogDixy
		DirTemp = TempDixy
	}
	CreateLogFile()
	CreateTempDir()
}
