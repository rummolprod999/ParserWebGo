package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type filelog string

var dirLog string
var dirTemp string
var fileLog filelog

func logging(args ...interface{}) {
	file, err := os.OpenFile(string(fileLog), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		fmt.Println("Ошибка записи в файл лога", err)
		return
	}
	fmt.Fprintf(file, "%v  ", time.Now())
	for _, v := range args {

		fmt.Fprintf(file, " %v", v)
	}
	fmt.Fprintln(file, "")

}

func createLogFile() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	dirlog := fmt.Sprintf("%s/%s", dir, dirLog)
	if _, err := os.Stat(dirlog); os.IsNotExist(err) {
		err := os.MkdirAll(dirlog, 0711)

		if err != nil {
			fmt.Println("Не могу создать папку для лога")
			os.Exit(1)
		}
	}
	t := time.Now()
	ft := t.Format("2006-01-02")
	switch a {
	case x5Group:
		fileLog = filelog(fmt.Sprintf("%s/log_X5Group_%v.log", dirlog, ft))
	case dixy:
		fileLog = filelog(fmt.Sprintf("%s/log_Dixy_%v.log", dirlog, ft))
	case rusneft:
		fileLog = filelog(fmt.Sprintf("%s/log_Rusneft_%v.log", dirlog, ft))
	case phosagro:
		fileLog = filelog(fmt.Sprintf("%s/log_Phosagro_%v.log", dirlog, ft))
	case icetrade:
		fileLog = filelog(fmt.Sprintf("%s/log_Icetrade_%v.log", dirlog, ft))
	case komtech:
		fileLog = filelog(fmt.Sprintf("%s/log_Komtech_%v.log", dirlog, ft))
	case ocontract:
		fileLog = filelog(fmt.Sprintf("%s/log_Ocontract_%v.log", dirlog, ft))
	case cpc:
		fileLog = filelog(fmt.Sprintf("%s/log_Cpc_%v.log", dirlog, ft))
	case novatek:
		fileLog = filelog(fmt.Sprintf("%s/log_Novatek_%v.log", dirlog, ft))
	case azot:
		fileLog = filelog(fmt.Sprintf("%s/log_Azot_%v.log", dirlog, ft))
	case uva:
		fileLog = filelog(fmt.Sprintf("%s/log_Uva_%v.log", dirlog, ft))
	case salym:
		fileLog = filelog(fmt.Sprintf("%s/log_Salym_%v.log", dirlog, ft))
	case monetka:
		fileLog = filelog(fmt.Sprintf("%s/log_Monetka_%v.log", dirlog, ft))
	case dtek:
		fileLog = filelog(fmt.Sprintf("%s/log_Dtek_%v.log", dirlog, ft))
	case mmk:
		fileLog = filelog(fmt.Sprintf("%s/log_Mmk_%v.log", dirlog, ft))
	case letoile:
		fileLog = filelog(fmt.Sprintf("%s/log_Letoile_%v.log", dirlog, ft))
	case sistema:
		fileLog = filelog(fmt.Sprintf("%s/log_Sistema_%v.log", dirlog, ft))
	case metafrax:
		fileLog = filelog(fmt.Sprintf("%s/log_Metafrax_%v.log", dirlog, ft))
	case ies:
		fileLog = filelog(fmt.Sprintf("%s/log_Ies_%v.log", dirlog, ft))
	case uralChem:
		fileLog = filelog(fmt.Sprintf("%s/log_UralChem_%v.log", dirlog, ft))
	case gosBy:
		fileLog = filelog(fmt.Sprintf("%s/log_GosBy_%v.log", dirlog, ft))
	case apk:
		fileLog = filelog(fmt.Sprintf("%s/log_Apk_%v.log", dirlog, ft))
	case aztpa:
		fileLog = filelog(fmt.Sprintf("%s/log_Aztpa_%v.log", dirlog, ft))
	case rosAtom:
		fileLog = filelog(fmt.Sprintf("%s/log_RosAtom_%v.log", dirlog, ft))
	case tpsre:
		fileLog = filelog(fmt.Sprintf("%s/log_Tpsre_%v.log", dirlog, ft))
	case tektkp:
		fileLog = filelog(fmt.Sprintf("%s/log_TekTkp_%v.log", dirlog, ft))
	case tekgaz:
		fileLog = filelog(fmt.Sprintf("%s/log_TekGaz_%v.log", dirlog, ft))
	case tekmarket:
		fileLog = filelog(fmt.Sprintf("%s/log_TekMarket_%v.log", dirlog, ft))
	case tekrao:
		fileLog = filelog(fmt.Sprintf("%s/log_TekRao_%v.log", dirlog, ft))
	case tekmos:
		fileLog = filelog(fmt.Sprintf("%s/log_TekMos_%v.log", dirlog, ft))
	case tekrn:
		fileLog = filelog(fmt.Sprintf("%s/log_TekRn_%v.log", dirlog, ft))
	case tekkom:
		fileLog = filelog(fmt.Sprintf("%s/log_TekKom_%v.log", dirlog, ft))
	case tekrusgazbur:
		fileLog = filelog(fmt.Sprintf("%s/log_TekRusGazBur_%v.log", dirlog, ft))
	case tekrosimport:
		fileLog = filelog(fmt.Sprintf("%s/log_TekRosImport_%v.log", dirlog, ft))
	case tektyumen:
		fileLog = filelog(fmt.Sprintf("%s/log_TekTyumen_%v.log", dirlog, ft))
	case teksil:
		fileLog = filelog(fmt.Sprintf("%s/log_TekSil_%v.log", dirlog, ft))
	case tekrzd:
		fileLog = filelog(fmt.Sprintf("%s/log_TekRzd_%v.log", dirlog, ft))
	case teksibur:
		fileLog = filelog(fmt.Sprintf("%s/log_TekSibur_%v.log", dirlog, ft))
	}

}

func createTempDir() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	dirtemp := fmt.Sprintf("%s/%s", dir, dirTemp)
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

func createEnv() {
	switch a {
	case x5Group:
		dirLog = logX5Group
		dirTemp = tempX5Group
	case dixy:
		dirLog = logDixy
		dirTemp = tempDixy
	case rusneft:
		dirLog = logRusneft
		dirTemp = tempRusneft
	case phosagro:
		dirLog = logPhosagro
		dirTemp = tempPhosagro
	case icetrade:
		dirLog = logIcetrade
		dirTemp = tempIcetrade
	case komtech:
		dirLog = logKomtech
		dirTemp = tempKomtech
	case ocontract:
		dirLog = logOcontract
		dirTemp = tempOcontract
	case cpc:
		dirLog = logCpc
		dirTemp = tempCpc
	case novatek:
		dirLog = logNovatek
		dirTemp = tempNovatek
	case azot:
		dirLog = logAzot
		dirTemp = tempAzot
	case uva:
		dirLog = logUva
		dirTemp = tempUva
	case salym:
		dirLog = logSalym
		dirTemp = tempSalym
	case monetka:
		dirLog = logMonetka
		dirTemp = tempMonetka
	case dtek:
		dirLog = logDtek
		dirTemp = tempDtek
	case mmk:
		dirLog = logMmk
		dirTemp = tempMmk
	case letoile:
		dirLog = logLetoile
		dirTemp = tempLetoile
	case sistema:
		dirLog = logSistema
		dirTemp = tempSistema
	case metafrax:
		dirLog = logMetafrax
		dirTemp = tempMetafrax
	case ies:
		dirLog = logIes
		dirTemp = tempIes
	case uralChem:
		dirLog = logUralChem
		dirTemp = tempUralChem
	case gosBy:
		dirLog = logGosBy
		dirTemp = tempGosBy
	case apk:
		dirLog = logApk
		dirTemp = tempApk
	case aztpa:
		dirLog = logAztpa
		dirTemp = tempAztpa
	case rosAtom:
		dirLog = logRosAtom
		dirTemp = tempRosAtom
	case tpsre:
		dirLog = logTpsre
		dirTemp = tempTpsre
	case tektkp:
		dirLog = logTektkp
		dirTemp = tempTektkp
	case tekgaz:
		dirLog = logTekGaz
		dirTemp = tempTekGaz
	case tekmarket:
		dirLog = logTekMarket
		dirTemp = tempTekMarket
	case tekrao:
		dirLog = logTekRao
		dirTemp = tempTekRao
	case tekmos:
		dirLog = logTekMos
		dirTemp = tempTekMos
	case tekrn:
		dirLog = logTekRn
		dirTemp = tempTekRn
	case tekkom:
		dirLog = logTekKom
		dirTemp = tempTekKom
	case tekrusgazbur:
		dirLog = logTekRusGazBur
		dirTemp = tempTekRusGazBur
	case tekrosimport:
		dirLog = logTekRosImport
		dirTemp = tempTekRosImport
	case tektyumen:
		dirLog = logTekTyumen
		dirTemp = tempTekTyumen
	case teksil:
		dirLog = logTekSil
		dirTemp = tempTekSil
	case tekrzd:
		dirLog = logTekRzd
		dirTemp = tempTekRzd
	case teksibur:
		dirLog = logTekSibur
		dirTemp = tempTekSibur
	}
	createLogFile()
	createTempDir()
}
