package main

import "flag"

func init() {
	flag.Parse()
	ArgS = flag.Arg(0)
	GetSetting()
	CreateLogFile()
}

func main() {

}
