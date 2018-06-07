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
	case Phosagro:
		p := ParserPhosagro{TypeFz: 35, Urls: []string{"https://etpreg.phosagro.ru/tenders/?PAGEN_1="}}
		MainParser(&p)
	case Icetrade:
		p := ParserIcetrade{TypeFz: 36, Urls: []string{"http://www.icetrade.by/search/auctions?search_text=&search=%D0%9D%D0%B0%D0%B9%D1%82%D0%B8&zakup_type[1]=1&zakup_type[2]=1&auc_num=&okrb=&company_title=&establishment=0&period=&created_from=&created_to=&request_end_from=&request_end_to=&t[Trade]=1&t[eTrade]=1&t[Request]=1&t[singleSource]=1&t[Auction]=1&t[Other]=1&t[contractingTrades]=1&t[socialOrder]=1&t[negotiations]=1&r[1]=1&r[2]=2&r[7]=7&r[3]=3&r[4]=4&r[6]=6&r[5]=5&sort=num%3Adesc&p="}}
		MainParser(&p)
	case Komtech:
		p := ParserKomtech{TypeFz: 40, Url: "http://zakupki.kom-tech.ru/main.asp?id="}
		MainParser(&p)
	case Ocontract:
		p := ParserOcontract{TypeFz: 41}
		MainParser(&p)

	}
}
