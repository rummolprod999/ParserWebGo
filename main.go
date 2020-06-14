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
		p := ParserIcetrade{TypeFz: 77, Urls: []string{"http://www.icetrade.by/search/auctions?search_text=&search=%D0%9D%D0%B0%D0%B9%D1%82%D0%B8&zakup_type[1]=1&zakup_type[2]=1&auc_num=&okrb=&company_title=&establishment=0&period=&created_from=&created_to=&request_end_from=&request_end_to=&t[Trade]=1&t[eTrade]=1&t[Request]=1&t[singleSource]=1&t[Auction]=1&t[Other]=1&t[contractingTrades]=1&t[socialOrder]=1&t[negotiations]=1&r[1]=1&r[2]=2&r[7]=7&r[3]=3&r[4]=4&r[6]=6&r[5]=5&sort=num%3Adesc&p="}}
		MainParser(&p)
	case Komtech:
		p := ParserKomtech{TypeFz: 40, Url: "http://zakupki.kom-tech.ru/main.asp?id="}
		MainParser(&p)
	case Ocontract:
		p := ParserOcontract{TypeFz: 41}
		MainParser(&p)
	case Cpc:
		p := ParserCpc{TypeFz: 52, Url: "http://www.cpc.ru/ru/tenders/Pages/default.aspx"}
		MainParser(&p)
	case Novatek:
		p := ParserNovatek{TypeFz: 60, Urls: []string{"http://www.novatek.ru/ru/about/tenders/supply/", "http://www.novatek.ru/ru/about/tenders/service/"}}
		MainParser(&p)
	case Azot:
		p := ParserAzot{TypeFz: 61, Url: "http://zakupki.sbu-azot.ru/?PAGEN_1="}
		MainParser(&p)
	case Uva:
		p := ParserUva{TypeFz: 64}
		MainParser(&p)
	case Salym:
		p := ParserSalymNew{TypeFz: 101}
		MainParser(&p)
	case Monetka:
		p := ParserMonetka{TypeFz: 104}
		MainParser(&p)
	case Dtek:
		p := ParserDtek{TypeFz: 117}
		MainParser(&p)
	case Mmk:
		p := ParserMmk{TypeFz: 119, Urls: []string{"http://mmk.ru/for_suppliers/tenders1/", "http://mmk.ru/for_suppliers/tenders2/"}}
		MainParser(&p)
	case Letoile:
		p := ParserLetoile{TypeFz: 120, Urls: []string{"http://b2b.letoile.ru/company/tenders/current/"}}
		MainParser(&p)
	case Sistema:
		p := ParserSistema{TypeFz: 121}
		MainParser(&p)
	case Metafrax:
		p := ParserMetafrax{TypeFz: 122, Urls: []string{"http://metafrax.ru/ru/p/181"}}
		MainParser(&p)
	case Ies:
		p := ParserIes{TypeFz: 123, Urls: []string{"http://zakupki.ies-holding.com/other/"}}
		MainParser(&p)
	case UralChem:
		d := ParserUralChem{TypeFz: 124, Urls: []string{"http://www.uralchem.ru/purchase/tenders_Ariba/?PAGEN_1="}}
		MainParser(&d)
		p := ParserUralChem{TypeFz: 124, Urls: []string{"http://www.uralchem.ru/purchase/tenders/?PAGEN_1="}}
		MainParser(&p)
	case GosBy:
		/*p := ParserGosBy{TypeFz: 137, Urls: []string{"http://www.goszakupki.by/search/auctions?auc_num=&search_text=&price_from=&price_to=&created_from=&created_to=&request_end_from=&request_end_to=&auction_date_from=&auction_date_to=&s[a]=1&s[b]=1&s[e]=1&s[p]=1&s[v]=1&s[c]=1&s[w]=1&s[s]=1&s[m]=1&s[ps]="}}*/
		p := ParserGosByNew{TypeFz: 137, Url: "http://goszakupki.by/tenders/posted?page="}
		MainParser(&p)
	case Apk:
		p := ParserApk{TypeFz: 182, Urls: []string{"http://tender-apk.ru/?nav-list-lot=page-"}}
		MainParser(&p)
	case Aztpa:
		p := ParserAztpa{TypeFz: 188, Urls: []string{"https://zakupki.aztpa.ru/zakupki/list?active=1&type=1"}}
		MainParser(&p)
	case RosAtom:
		p := ParserRosAtom{TypeFz: 221, Urls: []string{"http://zakupki.rosatom.ru/Web.aspx?node=currentorders&page="}}
		MainParser(&p)
	case Tpsre:
		p := ParserTpsre{TypeFz: 240, Urls: []string{"https://www.tpsre.ru/tenders/"}}
		MainParser(&p)
	case Tektkp:
		p := ParserTekTkp{TypeFz: 259, maxPage: 0}
		MainParser(&p)
	case Tekgaz:
		p := ParserTektorg{TypeFz: 22, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ТЭК Торг Газпром бурение", EtpUrl: "https://www.tektorg.ru/gazprom/procedures", Section: "8"}
		MainParser(&p)
	case Tekmarket:
		p := ParserTektorg{TypeFz: 139, maxPage: 0, Addtender: 0, Updatetender: 0, EtpName: "ТЭК Торг Маркет", EtpUrl: "https://www.tektorg.ru/market/procedures", Section: "10"}
		MainParser(&p)
	}

}
