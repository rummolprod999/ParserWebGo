package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var Prefix string
var DbName string
var UserDb string
var PassDb string
var Server string
var Port int
var TempTransneft string
var LogTransneft string
var ArgS string

type Arg int

const (
	Transneft Arg = iota

)
type Settings struct {
	Prefix string `xml:"prefix"`
	Db     string `xml:"db"`
	UserDb string `xml:"userdb"`
	PassDb string `xml:"passdb"`
	Server string `xml:"server"`
	Port   int    `xml:"port"`
	TempTransneft string `xml:"tempdir_transneft"`
	LogTransneft string `xml:"logdir_transneft"`
}

func GetSetting() {
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
	TempTransneft = settings.TempTransneft
	LogTransneft = settings.LogTransneft
}

func GetArgument(){

}