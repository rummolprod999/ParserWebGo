package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var arguments = "x5"
var Prefix string
var DbName string
var UserDb string
var PassDb string
var Server string
var Port int
var TempX5Group string
var LogX5Group string
var ArgS string
var A Arg
var Dsn string

type Arg int

const (
	X5Group Arg = iota
)

type Settings struct {
	Prefix      string `xml:"prefix"`
	Db          string `xml:"db"`
	UserDb      string `xml:"userdb"`
	PassDb      string `xml:"passdb"`
	Server      string `xml:"server"`
	Port        int    `xml:"port"`
	TempX5Group string `xml:"tempdir_x5group"`
	LogX5Group  string `xml:"logdir_x5group"`
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
	TempX5Group = settings.TempX5Group
	LogX5Group = settings.LogX5Group
	Dsn = fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=true&readTimeout=60m&maxAllowedPacket=0&timeout=60m&writeTimeout=60m&autocommit=true&loc=Local", UserDb, PassDb, DbName)
}

func GetArgument() {
	switch ArgS {
	case "x5":
		A = X5Group
	default:
		fmt.Printf("Bad argument, please use %s", arguments)
		os.Exit(1)
	}
}
