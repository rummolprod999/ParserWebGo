package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var arguments = "x5, dixy, rusneft, pphosagro, komtech, ocontract, cpc"
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
	default:
		fmt.Printf("Bad argument, please use %s", arguments)
		os.Exit(1)
	}
}
func checkEmptySettings() {
	if DbName == "" || UserDb == "" || PassDb == "" || Server == "" || TempX5Group == "" || LogX5Group == "" || TempDixy == "" || LogDixy == "" || TempRusneft == "" || LogRusneft == "" || TempPhosagro == "" || LogPhosagro == "" || TempIcetrade == "" || LogIcetrade == "" || TempKomtech == "" || LogKomtech == "" || TempOcontract == "" || LogOcontract == "" || TempCpc == "" || LogCpc == "" {
		fmt.Println("bad settings xml")
		os.Exit(1)
	}
}
