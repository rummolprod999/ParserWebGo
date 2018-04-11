package main

import "flag"

func init() {
	flag.Parse()
	ArgS = flag.Arg(0)
	GetSetting()
	CreateEnv()
}
func MainParser(p Parser) {
	p.parsing()
}
func main() {
	defer SaveStack()
	switch A {
	case X5Group:
		p := ParserX5Group{TypeFz: 26, Urls: []string{"https://tender.x5.ru/auction/guiding/list_auction/2-start", "https://tender.x5.ru/auction/guiding/list_auction/3-start", "https://tender.x5.ru/auction/guiding/list_auction/1-start", "https://tender.x5.ru/auction/guiding/list_auction/4-start"}}
		MainParser(&p)
	case Dixy:
		p := ParserDixy{TypeFz: 28, Urls: []string{"http://www.dixygroup.ru/our-partners/our-suppliers/tender-info.aspx?sc_lang=ru-RU"}}
		MainParser(&p)
	case Rusneft:
		p := ParserRusneft{TypeFz: 29, Urls: []string{"http://www.russneft.ru/tenders/russneft/", "http://www.russneft.ru/tenders/all/zapsibgroop/", "http://www.russneft.ru/tenders/all/centrsibgroop/", "http://www.russneft.ru/tenders/all/volgagroop/", "http://www.russneft.ru/tenders/all/belarus/", "http://www.russneft.ru/tenders/all/overseas/"}}
		MainParser(&p)

	}
}
