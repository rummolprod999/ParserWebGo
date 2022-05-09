package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var arguments = "x5, dixy, rusneft, phosagro, komtech, ocontract, cpc, novatek, azot, uva, salym, monetka, dtek, mmk, letoile, sistema, metafrax, ies, uralchem, gosby, apk, aztpa, rosatom, tpsre, tektkp, tekgaz, tekmarket, tekrao, tekmos, tekrn, tekkom, tekrusgazbur, tekrosimport, tektyumen, teksil, tekrzd, teksibur"
var prefix string
var dbName string
var userDb string
var passDb string
var server string
var port int
var pagesIcetrade int
var tempX5Group string
var logX5Group string
var tempDixy string
var logDixy string
var tempRusneft string
var logRusneft string
var tempPhosagro string
var logPhosagro string
var tempIcetrade string
var logIcetrade string
var tempKomtech string
var logKomtech string
var tempOcontract string
var logOcontract string
var tempCpc string
var logCpc string
var tempNovatek string
var logNovatek string
var tempAzot string
var logAzot string
var tempUva string
var logUva string
var tempSalym string
var logSalym string
var tempMonetka string
var logMonetka string
var tempDtek string
var logDtek string
var tempMmk string
var logMmk string
var tempLetoile string
var logLetoile string
var tempSistema string
var logSistema string
var tempMetafrax string
var logMetafrax string
var tempIes string
var logIes string
var tempUralChem string
var logUralChem string
var tempGosBy string
var logGosBy string
var tempApk string
var logApk string
var tempAztpa string
var logAztpa string
var tempRosAtom string
var logRosAtom string
var tempTpsre string
var logTpsre string
var tempTektkp string
var logTektkp string
var tempTekGaz string
var logTekGaz string
var tempTekMarket string
var logTekMarket string
var tempTekRao string
var logTekRao string
var tempTekMos string
var logTekMos string
var tempTekRn string
var logTekRn string
var tempTekKom string
var logTekKom string
var tempTekRusGazBur string
var logTekRusGazBur string
var tempTekRosImport string
var logTekRosImport string
var tempTekTyumen string
var logTekTyumen string
var tempTekSil string
var logTekSil string
var tempTekRzd string
var logTekRzd string
var tempTekSibur string
var logTekSibur string
var argS string
var a arg
var dsn string

type arg int

const (
	x5Group arg = iota
	dixy
	rusneft
	phosagro
	icetrade
	komtech
	ocontract
	cpc
	novatek
	azot
	uva
	salym
	monetka
	dtek
	mmk
	letoile
	sistema
	metafrax
	ies
	uralChem
	gosBy
	apk
	aztpa
	rosAtom
	tpsre
	tektkp
	tekgaz
	tekmarket
	tekrao
	tekmos
	tekrn
	tekkom
	tekrusgazbur
	tekrosimport
	tektyumen
	teksil
	tekrzd
	teksibur
)

type settings struct {
	Prefix           string `xml:"prefix"`
	Db               string `xml:"db"`
	UserDb           string `xml:"userdb"`
	PassDb           string `xml:"passdb"`
	Server           string `xml:"server"`
	Port             int    `xml:"port"`
	PagesIcetrade    int    `xml:"pages_icetrade"`
	TempX5Group      string `xml:"tempdir_x5group"`
	LogX5Group       string `xml:"logdir_x5group"`
	TempDixy         string `xml:"tempdir_dixy"`
	LogDixy          string `xml:"logdir_dixy"`
	TempRusneft      string `xml:"tempdir_rusneft"`
	LogRusneft       string `xml:"logdir_rusneft"`
	TempPhosagro     string `xml:"tempdir_phosagro"`
	LogPhosagro      string `xml:"logdir_phosagro"`
	TempIcetrade     string `xml:"tempdir_icetrade"`
	LogIcetrade      string `xml:"logdir_icetrade"`
	TempKomtech      string `xml:"tempdir_komtech"`
	LogKomtech       string `xml:"logdir_komtech"`
	TempOcontract    string `xml:"tempdir_onlinecontract"`
	LogOcontract     string `xml:"logdir_onlinecontract"`
	TempCpc          string `xml:"tempdir_cpc"`
	LogCpc           string `xml:"logdir_cpc"`
	TempNovatek      string `xml:"tempdir_novatek"`
	LogNovatek       string `xml:"logdir_novatek"`
	TempAzot         string `xml:"tempdir_azot"`
	LogAzot          string `xml:"logdir_azot"`
	TempUva          string `xml:"tempdir_uva"`
	LogUva           string `xml:"logdir_uva"`
	TempSalym        string `xml:"tempdir_salym"`
	LogSalym         string `xml:"logdir_salym"`
	TempMonetka      string `xml:"tempdir_monetka"`
	LogMonetka       string `xml:"logdir_monetka"`
	TempDtek         string `xml:"tempdir_dtek"`
	LogDtek          string `xml:"logdir_dtek"`
	TempMmk          string `xml:"tempdir_mmk"`
	LogMmk           string `xml:"logdir_mmk"`
	TempLetole       string `xml:"tempdir_letoile"`
	LogLetoile       string `xml:"logdir_letoile"`
	TempSistema      string `xml:"tempdir_sistema"`
	LogSistema       string `xml:"logdir_sistema"`
	TempMetafrax     string `xml:"tempdir_metafrax"`
	LogMetafrax      string `xml:"logdir_metafrax"`
	TempIes          string `xml:"tempdir_ies"`
	LogIes           string `xml:"logdir_ies"`
	TempUralChem     string `xml:"tempdir_uralchem"`
	LogUralChem      string `xml:"logdir_uralchem"`
	TempGosBy        string `xml:"tempdir_gosby"`
	LogGosBy         string `xml:"logdir_gosby"`
	TempApk          string `xml:"tempdir_apk"`
	LogApk           string `xml:"logdir_apk"`
	TempAztpa        string `xml:"tempdir_aztpa"`
	LogAztpa         string `xml:"logdir_aztpa"`
	TempRosAtom      string `xml:"tempdir_rosatom"`
	LogRosAtom       string `xml:"logdir_rosatom"`
	TempTpsre        string `xml:"tempdir_tpsre"`
	LogTpsre         string `xml:"logdir_tpsre"`
	TempTektkp       string `xml:"tempdir_tektkp"`
	LogTektkp        string `xml:"logdir_tektkp"`
	TempTekgaz       string `xml:"tempdir_tekgaz"`
	LogTekgaz        string `xml:"logdir_tekgaz"`
	TempTekmarket    string `xml:"tempdir_tekmarket"`
	LogTekmarket     string `xml:"logdir_tekmarket"`
	TempTekrao       string `xml:"tempdir_tekraoint"`
	LogTekrao        string `xml:"logdir_tekrao"`
	TempTekmos       string `xml:"tempdir_tekmos"`
	LogTekmos        string `xml:"logdir_tekmos"`
	TempTekrn        string `xml:"tempdir_tekrn"`
	LogTekrn         string `xml:"logdir_tekrn"`
	TempTekkom       string `xml:"tempdir_tekkom"`
	LogTekkom        string `xml:"logdir_tekkom"`
	TempTekrusgazbur string `xml:"tempdir_tekrusgazbur"`
	LogTekrusgazbur  string `xml:"logdir_tekrusgazbur"`
	TempTekrosimport string `xml:"tempdir_tekrosimport"`
	LogTekrosimport  string `xml:"logdir_tekrosimport"`
	TempTektyumen    string `xml:"tempdir_tektyumen"`
	LogTektyumen     string `xml:"logdir_tektyumen"`
	TempTeksil       string `xml:"tempdir_teksil"`
	LogTeksil        string `xml:"logdir_teksil"`
	TempTekrzd       string `xml:"tempdir_tekrzd"`
	LogTekrzd        string `xml:"logdir_tekrzd"`
	TempTeksibur     string `xml:"tempdir_teksibur"`
	LogTeksibur      string `xml:"logdir_teksibur"`
}

func getSetting() {
	getArgument()
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	xmlFile, err := os.Open(fmt.Sprintf("%s/settings_tenders.xml", dir))
	defer xmlFile.Close()
	if err != nil {
		println(err)
	}
	byteValue, _ := ioutil.ReadAll(xmlFile)
	var settings settings
	e := xml.Unmarshal(byteValue, &settings)
	if e != nil {
		println(e)
	}
	prefix = settings.Prefix
	dbName = settings.Db
	userDb = settings.UserDb
	passDb = settings.PassDb
	server = settings.Server
	port = settings.Port
	pagesIcetrade = settings.PagesIcetrade
	tempX5Group = settings.TempX5Group
	logX5Group = settings.LogX5Group
	logDixy = settings.LogDixy
	tempDixy = settings.TempDixy
	logRusneft = settings.LogRusneft
	tempRusneft = settings.TempRusneft
	logPhosagro = settings.LogPhosagro
	tempPhosagro = settings.TempPhosagro
	logIcetrade = settings.LogIcetrade
	tempIcetrade = settings.TempIcetrade
	logKomtech = settings.LogKomtech
	tempKomtech = settings.TempKomtech
	logOcontract = settings.LogOcontract
	tempOcontract = settings.TempOcontract
	logCpc = settings.LogCpc
	tempCpc = settings.TempCpc
	logNovatek = settings.LogNovatek
	tempNovatek = settings.TempNovatek
	logAzot = settings.LogAzot
	tempAzot = settings.TempAzot
	logUva = settings.LogUva
	tempUva = settings.TempUva
	logSalym = settings.LogSalym
	tempSalym = settings.TempSalym
	logMonetka = settings.LogMonetka
	tempMonetka = settings.TempMonetka
	logDtek = settings.LogDtek
	tempDtek = settings.TempDtek
	logMmk = settings.LogMmk
	tempMmk = settings.TempMmk
	logLetoile = settings.LogLetoile
	tempLetoile = settings.TempLetole
	logSistema = settings.LogSistema
	tempSistema = settings.TempSistema
	logMetafrax = settings.LogMetafrax
	tempMetafrax = settings.TempMetafrax
	logIes = settings.LogIes
	tempIes = settings.TempIes
	logUralChem = settings.LogUralChem
	tempUralChem = settings.TempUralChem
	logGosBy = settings.LogGosBy
	tempGosBy = settings.TempGosBy
	logApk = settings.LogApk
	tempApk = settings.TempApk
	logAztpa = settings.LogAztpa
	tempAztpa = settings.TempAztpa
	logRosAtom = settings.LogRosAtom
	tempRosAtom = settings.TempRosAtom
	logTpsre = settings.LogTpsre
	tempTpsre = settings.TempTpsre
	logTektkp = settings.LogTektkp
	tempTektkp = settings.TempTektkp
	logTekGaz = settings.LogTekgaz
	tempTekGaz = settings.TempTekgaz
	logTekMarket = settings.LogTekmarket
	tempTekMarket = settings.TempTekmarket
	logTekRao = settings.LogTekrao
	tempTekRao = settings.TempTekrao
	logTekMos = settings.LogTekmos
	tempTekMos = settings.TempTekmos
	logTekRn = settings.LogTekrn
	tempTekRn = settings.TempTekrn
	logTekKom = settings.LogTekkom
	tempTekKom = settings.TempTekkom
	logTekRusGazBur = settings.LogTekrusgazbur
	tempTekRusGazBur = settings.TempTekrusgazbur
	logTekRosImport = settings.LogTekrosimport
	tempTekRosImport = settings.TempTekrosimport
	logTekTyumen = settings.LogTektyumen
	tempTekTyumen = settings.TempTektyumen
	logTekSil = settings.LogTeksil
	tempTekSil = settings.TempTeksil
	logTekRzd = settings.LogTekrzd
	tempTekRzd = settings.TempTekrzd
	logTekSibur = settings.LogTeksibur
	tempTekSibur = settings.TempTeksibur
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&readTimeout=60m&maxAllowedPacket=0&timeout=60m&writeTimeout=60m&autocommit=true&loc=Local", userDb, passDb, server, port, dbName)
	checkEmptySettings()
}

func getArgument() {
	switch argS {
	case "x5":
		a = x5Group
	case "dixy":
		a = dixy
	case "rusneft":
		a = rusneft
	case "phosagro":
		a = phosagro
	case "icetrade":
		a = icetrade
	case "komtech":
		a = komtech
	case "ocontract":
		a = ocontract
	case "cpc":
		a = cpc
	case "novatek":
		a = novatek
	case "azot":
		a = azot
	case "uva":
		a = uva
	case "salym":
		a = salym
	case "monetka":
		a = monetka
	case "dtek":
		a = dtek
	case "mmk":
		a = mmk
	case "letoile":
		a = letoile
	case "sistema":
		a = sistema
	case "metafrax":
		a = metafrax
	case "ies":
		a = ies
	case "uralchem":
		a = uralChem
	case "gosby":
		a = gosBy
	case "apk":
		a = apk
	case "aztpa":
		a = aztpa
	case "rosatom":
		a = rosAtom
	case "tpsre":
		a = tpsre
	case "tektkp":
		a = tektkp
	case "tekgaz":
		a = tekgaz
	case "tekmarket":
		a = tekmarket
	case "tekrao":
		a = tekrao
	case "tekmos":
		a = tekmos
	case "tekrn":
		a = tekrn
	case "tekkom":
		a = tekkom
	case "tekrusgazbur":
		a = tekrusgazbur
	case "tekrosimport":
		a = tekrosimport
	case "tektyumen":
		a = tektyumen
	case "teksil":
		a = teksil
	case "tekrzd":
		a = tekrzd
	case "teksibur":
		a = teksibur
	default:
		fmt.Printf("Bad argument, please use %s", arguments)
		os.Exit(1)
	}
}
func checkEmptySettings() {
	if dbName == "" || userDb == "" || passDb == "" || server == "" || tempX5Group == "" || logX5Group == "" || tempDixy == "" || logDixy == "" || tempRusneft == "" || logRusneft == "" || tempPhosagro == "" || logPhosagro == "" || tempIcetrade == "" || logIcetrade == "" || tempKomtech == "" || logKomtech == "" || tempOcontract == "" || logOcontract == "" || tempCpc == "" || logCpc == "" || tempNovatek == "" || logNovatek == "" || tempAzot == "" || logAzot == "" || tempUva == "" || logUva == "" || tempSalym == "" || logSalym == "" || tempMonetka == "" || logMonetka == "" || tempDtek == "" || logDtek == "" || tempMmk == "" || logMmk == "" || tempLetoile == "" || logLetoile == "" || tempSistema == "" || logSistema == "" || tempMetafrax == "" || logMetafrax == "" || tempIes == "" || logIes == "" || tempUralChem == "" || logUralChem == "" || tempGosBy == "" || logGosBy == "" || tempApk == "" || logApk == "" || tempAztpa == "" || logAztpa == "" || tempRosAtom == "" || logRosAtom == "" || tempTpsre == "" || logTpsre == "" || tempTektkp == "" || logTektkp == "" || tempTekGaz == "" || logTekGaz == "" || tempTekMarket == "" || logTekMarket == "" || tempTekRao == "" || logTekRao == "" || tempTekMos == "" || logTekMos == "" || tempTekRn == "" || logTekRn == "" || tempTekKom == "" || logTekKom == "" || tempTekRusGazBur == "" || logTekRusGazBur == "" || tempTekRosImport == "" || logTekRosImport == "" || tempTekTyumen == "" || logTekTyumen == "" || tempTekSil == "" || logTekSil == "" || tempTekRzd == "" || logTekRzd == "" || tempTekSibur == "" || logTekSibur == "" {
		fmt.Println("bad settings xml")
		os.Exit(1)
	}
}
