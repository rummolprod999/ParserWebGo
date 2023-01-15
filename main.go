package main

import (
	"flag"
	"time"
)

var sleepingTimeM = 7 * time.Second

func init() {
	flag.Parse()
	argS = flag.Arg(0)
	getSetting()
	createEnv()
}
func mainParser(p Parser) {
	p.parsing()
}
func main() {
	defer SaveStack()
	switch a {
	case x5Group:
		p := parserX5Group{TypeFz: 26, Urls: []string{"https://tender.x5.ru/auction/guiding/list_auction/2-start", "https://tender.x5.ru/auction/guiding/list_auction/3-start", "https://tender.x5.ru/auction/guiding/list_auction/1-start", "https://tender.x5.ru/auction/guiding/list_auction/4-start"}}
		mainParser(&p)
	case dixy:
		p := parserDixy{TypeFz: 28, Urls: []string{"http://www.dixygroup.ru/our-partners/our-suppliers/tender-info.aspx?sc_lang=ru-RU"}}
		mainParser(&p)
	case rusneft:
		p := parserRusneft{TypeFz: 29, Urls: []string{"http://www.russneft.ru/tenders/russneft/", "http://www.russneft.ru/tenders/all/zapsibgroop/", "http://www.russneft.ru/tenders/all/centrsibgroop/", "http://www.russneft.ru/tenders/all/volgagroop/", "http://www.russneft.ru/tenders/all/belarus/", "http://www.russneft.ru/tenders/all/overseas/"}}
		mainParser(&p)
	case phosagro:
		p := parserPhosagro{TypeFz: 35, Urls: []string{"https://etpreg.phosagro.ru/tenders/?PAGEN_2=", "https://etpreg.phosagro.ru/services/?PAGEN_2="}}
		mainParser(&p)
	case icetrade:
		p := parserIcetrade{TypeFz: 77, Urls: []string{"http://www.icetrade.by/search/auctions?search_text=&search=%D0%9D%D0%B0%D0%B9%D1%82%D0%B8&zakup_type[1]=1&zakup_type[2]=1&auc_num=&okrb=&company_title=&establishment=0&period=&created_from=&created_to=&request_end_from=&request_end_to=&t[Trade]=1&t[eTrade]=1&t[Request]=1&t[singleSource]=1&t[Auction]=1&t[Other]=1&t[contractingTrades]=1&t[socialOrder]=1&t[negotiations]=1&r[1]=1&r[2]=2&r[7]=7&r[3]=3&r[4]=4&r[6]=6&r[5]=5&sort=num%3Adesc&p="}}
		mainParser(&p)
	case komtech:
		p := parserKomtech{TypeFz: 40, Url: "http://zakupki.kom-tech.ru/main.asp?id="}
		mainParser(&p)
	case ocontract:
		p := parserOcontract{TypeFz: 41}
		mainParser(&p)
	case cpc:
		p := parserCpc{TypeFz: 52, Url: "http://www.cpc.ru/ru/tenders/Pages/default.aspx"}
		mainParser(&p)
	case novatek:
		p := parserNovatek{TypeFz: 60, Urls: []string{"https://www.novatek.ru/ru/about/tenders/supply/", "https://www.novatek.ru/ru/about/tenders/service/"}}
		mainParser(&p)
	case azot:
		p := parserAzot{TypeFz: 61, Url: "http://zakupki.sbu-azot.ru/procurement_requests/open_request_for_proposals/current/?SECTION_CODE=procurement_requests&SECTION_CODE_2=open_request_for_proposals&SECTION_CODE_3=current&PAGEN_1="}
		mainParser(&p)
	case uva:
		p := parserUva{TypeFz: 64}
		mainParser(&p)
	case salym:
		p := parserSalymNew{TypeFz: 101}
		mainParser(&p)
	case monetka:
		p := parserMonetka{TypeFz: 104}
		mainParser(&p)
	case dtek:
		p := parserDtek{TypeFz: 117}
		mainParser(&p)
	case mmk:
		//p := parserMmk{TypeFz: 119, Urls: []string{"http://mmk.ru/for_suppliers/tenders1/", "http://mmk.ru/for_suppliers/tenders2/"}}
		//mainParser(&p)
		m := parserMmk2{TypeFz: 119, Urls: []string{"https://mmk.ru/ru/for-suppliers/auction/"}}
		mainParser(&m)
		t := parserMmk3{TypeFz: 119, Urls: []string{"https://mmk.ru/ru/for-suppliers/auction/purchase-of-services/"}}
		mainParser(&t)
	case letoile:
		p := parserLetoile{TypeFz: 120, Urls: []string{"http://b2b.letoile.ru/company/tenders/current/"}}
		mainParser(&p)
	case sistema:
		p := parserSistema{TypeFz: 121}
		mainParser(&p)
	case metafrax:
		p := parserMetafrax{TypeFz: 122, Urls: []string{"http://metafrax.ru/tender/tendery-na-postavku-tovarov", "http://metafrax.ru/tender/tendery-na-vypolnenie-rabot-i-uslug"}}
		mainParser(&p)
	case ies:
		p := parserIes{TypeFz: 123, Urls: []string{"http://zakupki.ies-holding.com/other/"}}
		mainParser(&p)
	case uralChem:
		d := parserUralChem{TypeFz: 124, Urls: []string{"http://www.uralchem.ru/purchase/tenders_Ariba/?PAGEN_1="}}
		mainParser(&d)
		p := parserUralChem{TypeFz: 124, Urls: []string{"http://www.uralchem.ru/purchase/tenders/?PAGEN_1="}}
		mainParser(&p)
	case gosBy:
		/*p := ParserGosBy{TypeFz: 137, Urls: []string{"http://www.goszakupki.by/search/auctions?auc_num=&search_text=&price_from=&price_to=&created_from=&created_to=&request_end_from=&request_end_to=&auction_date_from=&auction_date_to=&s[a]=1&s[b]=1&s[e]=1&s[p]=1&s[v]=1&s[c]=1&s[w]=1&s[s]=1&s[m]=1&s[ps]="}}*/
		p := parserGosByNew{TypeFz: 137, Url: "http://goszakupki.by/tenders/posted?page="}
		mainParser(&p)
	case apk:
		p := parserApk{TypeFz: 182, Urls: []string{"http://tender-apk.ru/?nav-list-lot=page-"}}
		mainParser(&p)
	case aztpa:
		p := parserAztpa{TypeFz: 188, Urls: []string{"https://zakupki.aztpa.ru/zakupki/list?active=1&type=1"}}
		mainParser(&p)
	case rosAtom:
		p := parserRosAtom{TypeFz: 221, Urls: []string{"http://zakupki.rosatom.ru/Web.aspx?node=currentorders&page="}}
		mainParser(&p)
	case tpsre:
		p := parserTpsre{TypeFz: 240, Urls: []string{"https://www.tpsre.ru/tenders/"}}
		mainParser(&p)
	case tektkp:
		p := parserTekTkp{TypeFz: 259, maxPage: 0}
		mainParser(&p)
	case tekgaz:
		p := parserTektorg{TypeFz: 22, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ТЭК Торг Газпром бурение", EtpUrl: "https://www.tektorg.ru/gazprom/procedures", Section: "8"}
		mainParser(&p)
	case tekmarket:
		p := parserTektorg{TypeFz: 139, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ТЭК Торг Маркет", EtpUrl: "https://www.tektorg.ru/market/procedures", Section: "10"}
		mainParser(&p)
	case tekrao:
		p := parserTektorg{TypeFz: 24, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ТЭК Торг Интер РАО", EtpUrl: "https://www.tektorg.ru/interao/procedures", Section: "9"}
		mainParser(&p)
	case tekmos:
		p := parserTektorg{TypeFz: 140, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ТЭК Торг ТЭК Мосэнерго", EtpUrl: "https://www.tektorg.ru/mosenergo/procedures", Section: "14"}
		mainParser(&p)
	case tekrn:
		p := parserTektorg{TypeFz: 149, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ТЭК Торг ТЭК Роснефть", EtpUrl: "https://www.tektorg.ru/rosneft/procedures", Section: "6"}
		mainParser(&p)
	case tekkom:
		p := parserTektorg{TypeFz: 138, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ТЭК Торг Коммерческие закупки и 223-ФЗ", EtpUrl: "https://www.tektorg.ru/223-fz/procedures", Section: "3"}
		mainParser(&p)
	case tekrusgazbur:
		p := parserTektorg{TypeFz: 325, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ООО «Русгазбурение»", EtpUrl: "https://www.tektorg.ru/rusgazburenie/procedures", Section: "26"}
		mainParser(&p)
	case tekrosimport:
		p := parserTektorg{TypeFz: 326, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ФГУП «Росморпорт»", EtpUrl: "https://www.tektorg.ru/rosmorport/procedures", Section: "29"}
		mainParser(&p)
		m := parserTektorg{TypeFz: 326, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ФГУП «Росморпорт»", EtpUrl: "https://www.tektorg.ru/rosmorport/procedures", Section: "33"}
		mainParser(&m)
	case tektyumen:
		p := parserTektorg{TypeFz: 327, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "Портал поставщиков Тюменской области", EtpUrl: "https://www.tektorg.ru/portal_tyumen/procedures", Section: "19"}
		mainParser(&p)
	case teksil:
		p := parserTektorg{TypeFz: 218, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ТЭК Торг ПАО «Силовые машины»", EtpUrl: "https://www.tektorg.ru/silovyi_machine/procedures", Section: "15"}
		mainParser(&p)
	case tekrzd:
		p := parserTektorg{TypeFz: 25, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ТЭК Торг РЖД", EtpUrl: "https://www.tektorg.ru/rzd/procedures", Section: "2"}
		mainParser(&p)
	case teksibur:
		p := parserTektorg{TypeFz: 361, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "АО «Сибур-Химпром»", EtpUrl: "https://www.tektorg.ru/sibur/procedures", Section: "37"}
		mainParser(&p)
	case tekppk:
		p := parserTektorg{TypeFz: 371, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "АО \"Первая Портовая Компания\"", EtpUrl: "https://www.tektorg.ru/portone/procedures", Section: "44"}
		mainParser(&p)
	case tekspec:
		p := parserTektorg{TypeFz: 372, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "АО  «Спецнефтетранс»", EtpUrl: "https://www.tektorg.ru/portone/procedures", Section: "42"}
		mainParser(&p)
	case grls:
		p := GrlsReader{Url: "https://grls.rosminzdrav.ru/pricelims.aspx", Added: 0, AddedExcept: 0}
		mainParser(&p)
	}

}
