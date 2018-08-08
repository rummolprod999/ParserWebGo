package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
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
	var today = time.Now()
	oneDay := time.Hour * -24
	yesterday := today.Add(oneDay)
	yesterdayS := yesterday.Format("02.01.2006")
	startUrl := fmt.Sprintf("http://www.icetrade.by/search/auctions?search_text=&zakup_type[1]=1&zakup_type[2]=1&auc_num=&okrb=&company_title=&establishment=0&industries=&period=&created_from=%s&created_to=%s&request_end_from=&request_end_to=&t[Trade]=1&t[eTrade]=1&t[socialOrder]=1&t[singleSource]=1&t[Auction]=1&t[Request]=1&t[contractingTrades]=1&t[negotiations]=1&t[Other]=1&r[1]=1&r[2]=2&r[7]=7&r[3]=3&r[4]=4&r[6]=6&r[5]=5&sort=num%%3Adesc&sbm=1&onPage=20&p=", yesterdayS, yesterdayS)
	countPage := t.getCountPage(startUrl)
	for i := 1; i <= countPage; i++ {
		ul := fmt.Sprintf("%s%d", startUrl, i)
		t.parsingPage(ul)
		time.Sleep(5 * time.Second)
	}
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderPhosagro))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderPhosagro))
}
func (t *ParserIcetrade) getCountPage(url string) int {
	c := 10
	u := fmt.Sprintf("%s%d", url, 1)
	defer SaveStack()
	r := DownloadPage(u)
	if r != "" {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(r))
		if err != nil {
			Logging(err)
			return c
		}
		numP := strings.TrimSpace(doc.Find("div.paging a").Last().Text())
		i, err := strconv.Atoi(numP)
		if err != nil {
			return c
		}
		return i
	} else {
		Logging("Получили пустую строку", u)
	}
	return c
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
		time.Sleep(5 * time.Second)
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
