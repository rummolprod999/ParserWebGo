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
	case Rusneft:
		FileLog = Filelog(fmt.Sprintf("%s/log_Rusneft_%v.log", dirlog, ft))
	case Phosagro:
		FileLog = Filelog(fmt.Sprintf("%s/log_Phosagro_%v.log", dirlog, ft))
	case Icetrade:
		FileLog = Filelog(fmt.Sprintf("%s/log_Icetrade_%v.log", dirlog, ft))
	case Komtech:
		FileLog = Filelog(fmt.Sprintf("%s/log_Komtech_%v.log", dirlog, ft))
	case Ocontract:
		FileLog = Filelog(fmt.Sprintf("%s/log_Ocontract_%v.log", dirlog, ft))
	case Cpc:
		FileLog = Filelog(fmt.Sprintf("%s/log_Cpc_%v.log", dirlog, ft))
	case Novatek:
		FileLog = Filelog(fmt.Sprintf("%s/log_Novatek_%v.log", dirlog, ft))
	case Azot:
		FileLog = Filelog(fmt.Sprintf("%s/log_Azot_%v.log", dirlog, ft))
	case Uva:
		FileLog = Filelog(fmt.Sprintf("%s/log_Uva_%v.log", dirlog, ft))
	case Salym:
		FileLog = Filelog(fmt.Sprintf("%s/log_Salym_%v.log", dirlog, ft))
	case Monetka:
		FileLog = Filelog(fmt.Sprintf("%s/log_Monetka_%v.log", dirlog, ft))
	case Dtek:
		FileLog = Filelog(fmt.Sprintf("%s/log_Dtek_%v.log", dirlog, ft))
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
	case Rusneft:
		DirLog = LogRusneft
		DirTemp = TempRusneft
	case Phosagro:
		DirLog = LogPhosagro
		DirTemp = TempPhosagro
	case Icetrade:
		DirLog = LogIcetrade
		DirTemp = TempIcetrade
	case Komtech:
		DirLog = LogKomtech
		DirTemp = TempKomtech
	case Ocontract:
		DirLog = LogOcontract
		DirTemp = TempOcontract
	case Cpc:
		DirLog = LogCpc
		DirTemp = TempCpc
	case Novatek:
		DirLog = LogNovatek
		DirTemp = TempNovatek
	case Azot:
		DirLog = LogAzot
		DirTemp = TempAzot
	case Uva:
		DirLog = LogUva
		DirTemp = TempUva
	case Salym:
		DirLog = LogSalym
		DirTemp = TempSalym
	case Monetka:
		DirLog = LogMonetka
		DirTemp = TempMonetka
	case Dtek:
		DirLog = LogDtek
		DirTemp = TempDtek
	}
	CreateLogFile()
	CreateTempDir()
}
