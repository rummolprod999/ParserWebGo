package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"time"
)

var AddtenderIcetrade int
var UpdatetenderIcetrade int

type ParserIcetrade struct {
	TypeFz int
	Urls   []string
}

type TenderIcetrade struct {
	purNum     string
	url        string
	orgName    string
	lotPrice   string
	purObjInfo string
}

func (t *ParserIcetrade) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	for _, p := range t.Urls {
		for i := 1; i <= PagesIcetrade; i++ {
			urllist := fmt.Sprintf("%s%s", p, fmt.Sprintf("%d&onPage=100", i))
			t.parsingPage(urllist)
		}
	}

	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderPhosagro))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderPhosagro))
}

func (t *ParserIcetrade) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserIcetrade) parsingTenderList(p string, url string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("#auctions-list tbody tr[class^='rw']").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)

	})
}

func (t *ParserIcetrade) parsingTenderFromList(p *goquery.Selection, url string) {
	defer SaveStack()
	purNum := strings.TrimSpace(p.Find("td:nth-child(4)").First().Text())
	if purNum == "" {
		Logging("Can not find purnum in ", url)
		return
	}
	orgName := strings.TrimSpace(p.Find("td:nth-child(2)").First().Text())
	purObjInfo := strings.TrimSpace(p.Find("td:nth-child(1) a").First().Text())
	hrefT := p.Find("td:nth-child(1) a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", hrefT.Text())
		return
	}
	lotPrice := strings.TrimSpace(p.Find("td:nth-child(5) a").First().Text())
	tnd := TenderIcetrade{purNum: purNum, url: href, orgName: orgName, lotPrice: lotPrice, purObjInfo: purObjInfo}
	t.Tender(tnd)
}

func (t *ParserIcetrade) Tender(tn TenderIcetrade) {
	defer SaveStack()
	r := DownloadPage(tn.url)
	if r == "" {
		Logging("Получили пустую строку", tn.url)
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r))
	if err != nil {
		Logging(err)
		return
	}
	pubDateT := strings.TrimSpace(doc.Find("table td:contains('Дата размещения приглашения') + td").First().Text())
	pubDate := getTimeMoscowLayout(pubDateT, "02.01.2006")
	if (pubDate == time.Time{}) {
		pubDate = getTimeMoscowLayout(pubDateT, "02.01.2006 15:04")
	}
	endDateT := strings.TrimSpace(doc.Find("table td:contains('Дата и время окончания приема предложений') + td").First().Text())
	endDate := getTimeMoscowLayout(endDateT, "02.01.2006 15:04")
	if (endDate == time.Time{}) {
		endDate = getTimeMoscowLayout(endDateT, "02.01.2006")
	}
	if (pubDate == time.Time{} || endDate == time.Time{}) {
		Logging("Can not find enddate or startdate in ", tn.url, tn.purNum)
		return
	}
	fmt.Println(pubDate)
	fmt.Println(endDate)
}
