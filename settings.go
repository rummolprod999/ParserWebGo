package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var arguments = "x5, dixy, rusneft, phosagro, komtech, ocontract, cpc, novatek, azot, uva, salym, monetka, dtek, mmk, letoile, sistema, metafrax, ies, uralchem, gosby, apk, aztpa, rosatom, tpsre"
var Prefix string
var DbName string
var UserDb string
var PassDb string
var Server string
var Port int
var PagesIcetrade int
var TempX5Group string
var LogX5Group string
var TempDixy string
var LogDixy string
var TempRusneft string
var LogRusneft string
var TempPhosagro string
var LogPhosagro string
var TempIcetrade string
var LogIcetrade string
var TempKomtech string
var LogKomtech string
var TempOcontract string
var LogOcontract string
var TempCpc string
var LogCpc string
var TempNovatek string
var LogNovatek string
var TempAzot string
var LogAzot string
var TempUva string
var LogUva string
var TempSalym string
var LogSalym string
var TempMonetka string
var LogMonetka string
var TempDtek string
var LogDtek string
var TempMmk string
var LogMmk string
var TempLetoile string
var LogLetoile string
var TempSistema string
var LogSistema string
var TempMetafrax string
var LogMetafrax string
var TempIes string
var LogIes string
var TempUralChem string
var LogUralChem string
var TempGosBy string
var LogGosBy string
var TempApk string
var LogApk string
var TempAztpa string
var LogAztpa string
var TempRosAtom string
var LogRosAtom string
var TempTpsre string
var LogTpsre string
var ArgS string
var A Arg
var Dsn string

type Arg int

const (
	X5Group Arg = iota
	Dixy
	Rusneft
	Phosagro
	Icetrade
	Komtech
	Ocontract
	Cpc
	Novatek
	Azot
	Uva
	Salym
	Monetka
	Dtek
	Mmk
	Letoile
	Sistema
	Metafrax
	Ies
	UralChem
	GosBy
	Apk
	Aztpa
	RosAtom
	Tpsre
)

type Settings struct {
	Prefix        string `xml:"prefix"`
	Db            string `xml:"db"`
	UserDb        string `xml:"userdb"`
	PassDb        string `xml:"passdb"`
	Server        string `xml:"server"`
	Port          int    `xml:"port"`
	PagesIcetrade int    `xml:"pages_icetrade"`
	TempX5Group   string `xml:"tempdir_x5group"`
	LogX5Group    string `xml:"logdir_x5group"`
	TempDixy      string `xml:"tempdir_dixy"`
	LogDixy       string `xml:"logdir_dixy"`
	TempRusneft   string `xml:"tempdir_rusneft"`
	LogRusneft    string `xml:"logdir_rusneft"`
	TempPhosagro  string `xml:"tempdir_phosagro"`
	LogPhosagro   string `xml:"logdir_phosagro"`
	TempIcetrade  string `xml:"tempdir_icetrade"`
	LogIcetrade   string `xml:"logdir_icetrade"`
	TempKomtech   string `xml:"tempdir_komtech"`
	LogKomtech    string `xml:"logdir_komtech"`
	TempOcontract string `xml:"tempdir_onlinecontract"`
	LogOcontract  string `xml:"logdir_onlinecontract"`
	TempCpc       string `xml:"tempdir_cpc"`
	LogCpc        string `xml:"logdir_cpc"`
	TempNovatek   string `xml:"tempdir_novatek"`
	LogNovatek    string `xml:"logdir_novatek"`
	TempAzot      string `xml:"tempdir_azot"`
	LogAzot       string `xml:"logdir_azot"`
	TempUva       string `xml:"tempdir_uva"`
	LogUva        string `xml:"logdir_uva"`
	TempSalym     string `xml:"tempdir_salym"`
	LogSalym      string `xml:"logdir_salym"`
	TempMonetka   string `xml:"tempdir_monetka"`
	LogMonetka    string `xml:"logdir_monetka"`
	TempDtek      string `xml:"tempdir_dtek"`
	LogDtek       string `xml:"logdir_dtek"`
	TempMmk      string `xml:"tempdir_mmk"`
	LogMmk       string `xml:"logdir_mmk"`
	TempLetole   string `xml:"tempdir_letoile"`
	LogLetoile   string `xml:"logdir_letoile"`
	TempSistema  string `xml:"tempdir_sistema"`
	LogSistema   string `xml:"logdir_sistema"`
	TempMetafrax string `xml:"tempdir_metafrax"`
	LogMetafrax  string `xml:"logdir_metafrax"`
	TempIes      string `xml:"tempdir_ies"`
	LogIes       string `xml:"logdir_ies"`
	TempUralChem string `xml:"tempdir_uralchem"`
	LogUralChem  string `xml:"logdir_uralchem"`
	TempGosBy    string `xml:"tempdir_gosby"`
	LogGosBy     string `xml:"logdir_gosby"`
	TempApk      string `xml:"tempdir_apk"`
	LogApk       string `xml:"logdir_apk"`
	TempAztpa    string `xml:"tempdir_aztpa"`
	LogAztpa     string `xml:"logdir_aztpa"`
	TempRosAtom  string `xml:"tempdir_rosatom"`
	LogRosAtom   string `xml:"logdir_rosatom"`
	TempTpsre    string `xml:"tempdir_tpsre"`
	LogTpsre     string `xml:"logdir_tpsre"`
}

func GetSetting() {
	GetArgument()
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	xmlFile, err := os.Open(fmt.Sprintf("%s/settings_tenders.xml", dir))
	defer xmlFile.Close()
	if err != nil {
		println(err)
	}
	byteValue, _ := ioutil.ReadAll(xmlFile)
	var settings Settings
	e := xml.Unmarshal(byteValue, &settings)
	if e != nil {
		println(e)
	}
	Prefix = settings.Prefix
	DbName = settings.Db
	UserDb = settings.UserDb
	PassDb = settings.PassDb
	Server = settings.Server
	Port = settings.Port
	PagesIcetrade = settings.PagesIcetrade
	TempX5Group = settings.TempX5Group
	LogX5Group = settings.LogX5Group
	LogDixy = settings.LogDixy
	TempDixy = settings.TempDixy
	LogRusneft = settings.LogRusneft
	TempRusneft = settings.TempRusneft
	LogPhosagro = settings.LogPhosagro
	TempPhosagro = settings.TempPhosagro
	LogIcetrade = settings.LogIcetrade
	TempIcetrade = settings.TempIcetrade
	LogKomtech = settings.LogKomtech
	TempKomtech = settings.TempKomtech
	LogOcontract = settings.LogOcontract
	TempOcontract = settings.TempOcontract
	LogCpc = settings.LogCpc
	TempCpc = settings.TempCpc
	LogNovatek = settings.LogNovatek
	TempNovatek = settings.TempNovatek
	LogAzot = settings.LogAzot
	TempAzot = settings.TempAzot
	LogUva = settings.LogUva
	TempUva = settings.TempUva
	LogSalym = settings.LogSalym
	TempSalym = settings.TempSalym
	LogMonetka = settings.LogMonetka
	TempMonetka = settings.TempMonetka
	LogDtek = settings.LogDtek
	TempDtek = settings.TempDtek
	LogMmk = settings.LogMmk
	TempMmk = settings.TempMmk
	LogLetoile = settings.LogLetoile
	TempLetoile = settings.TempLetole
	LogSistema = settings.LogSistema
	TempSistema = settings.TempSistema
	LogMetafrax = settings.LogMetafrax
	TempMetafrax = settings.TempMetafrax
	LogIes = settings.LogIes
	TempIes = settings.TempIes
	LogUralChem = settings.LogUralChem
	TempUralChem = settings.TempUralChem
	LogGosBy = settings.LogGosBy
	TempGosBy = settings.TempGosBy
	LogApk = settings.LogApk
	TempApk = settings.TempApk
	LogAztpa = settings.LogAztpa
	TempAztpa = settings.TempAztpa
	LogRosAtom = settings.LogRosAtom
	TempRosAtom = settings.TempRosAtom
	LogTpsre = settings.LogTpsre
	TempTpsre = settings.TempTpsre
	Dsn = fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=true&readTimeout=60m&maxAllowedPacket=0&timeout=60m&writeTimeout=60m&autocommit=true&loc=Local", UserDb, PassDb, DbName)
	checkEmptySettings()
}

func GetArgument() {
	switch ArgS {
	case "x5":
		A = X5Group
	case "dixy":
		A = Dixy
	case "rusneft":
		A = Rusneft
	case "phosagro":
		A = Phosagro
	case "icetrade":
		A = Icetrade
	case "komtech":
		A = Komtech
	case "ocontract":
		A = Ocontract
	case "cpc":
		A = Cpc
	case "novatek":
		A = Novatek
	case "azot":
		A = Azot
	case "uva":
		A = Uva
	case "salym":
		A = Salym
	case "monetka":
		A = Monetka
	case "dtek":
		A = Dtek
	case "mmk":
		A = Mmk
	case "letoile":
		A = Letoile
	case "sistema":
		A = Sistema
	case "metafrax":
		A = Metafrax
	case "ies":
		A = Ies
	case "uralchem":
		A = UralChem
	case "gosby":
		A = GosBy
	case "apk":
		A = Apk
	case "aztpa":
		A = Aztpa
	case "rosatom":
		A = RosAtom
	case "tpsre":
		A = Tpsre
	default:
		fmt.Printf("Bad argument, please use %s", arguments)
		os.Exit(1)
	}
}
func checkEmptySettings() {
	if DbName == "" || UserDb == "" || PassDb == "" || Server == "" || TempX5Group == "" || LogX5Group == "" || TempDixy == "" || LogDixy == "" || TempRusneft == "" || LogRusneft == "" || TempPhosagro == "" || LogPhosagro == "" || TempIcetrade == "" || LogIcetrade == "" || TempKomtech == "" || LogKomtech == "" || TempOcontract == "" || LogOcontract == "" || TempCpc == "" || LogCpc == "" || TempNovatek == "" || LogNovatek == "" || TempAzot == "" || LogAzot == "" || TempUva == "" || LogUva == "" || TempSalym == "" || LogSalym == "" || TempMonetka == "" || LogMonetka == "" || TempDtek == "" || LogDtek == "" || TempMmk == "" || LogMmk == "" || TempLetoile == "" || LogLetoile == "" || TempSistema == "" || LogSistema == "" || TempMetafrax == "" || LogMetafrax == "" || TempIes == "" || LogIes == "" || TempUralChem == "" || LogUralChem == "" || TempGosBy == "" || LogGosBy == "" || TempApk == "" || LogApk == "" || TempAztpa == "" || LogAztpa == "" || TempRosAtom == "" || LogRosAtom == "" || TempTpsre == "" || LogTpsre == "" {
		fmt.Println("bad settings xml")
		os.Exit(1)
	}
}
