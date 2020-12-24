package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/buger/jsonparser"
	_ "github.com/go-sql-driver/mysql"
	strip "github.com/grokify/html-strip-tags-go"
	"net/url"
	"strings"
	"time"
)

type ParserMmk2 struct {
	TypeFz int
	Urls   []string
}

type TenderMmk2 struct {
	purName       string
	purNum        string
	url           string
	pubDate       time.Time
	endDate       time.Time
	status        string
	contactPerson string
	email         string
}

func (t *ParserMmk2) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	t.parsingPageAll()
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderMmk))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderMmk))
}

func (t *ParserMmk2) parsingPageAll() {
	for _, p := range t.Urls {
		t.parsingPage(p)
	}
}

func (t *ParserMmk2) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserMmk2) parsingTenderList(p string, url string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("a[href = '/for_suppliers/auction/index.php']").Each(func(i int, s *goquery.Selection) {
		t.parsingCategoryFromList(s, url)
	})
}

func (t *ParserMmk2) parsingCategoryFromList(p *goquery.Selection, _ string) {
	onclick, exist := p.Attr("onclick")
	if !exist {
		return
	}
	val1, val2 := findFromRegExpTwoValue(onclick, `'(.+)','(.+)'`)
	val1 = strings.Replace(val1, "\\", "", -1)
	val2 = strings.Replace(val2, "\\", "", -1)
	urlP := url.QueryEscape(fmt.Sprintf("\\'%s\\'", val1))
	urlT := fmt.Sprintf("http://mmk.ru/for_suppliers/auction/source.php?LOCATION_CODE=%s", urlP)
	r := DownloadPage(urlT)
	if r != "" {
		t.parsingTenderFromJsonList(r)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserMmk2) parsingTenderFromJsonList(page string) {
	defer SaveStack()
	_, err := jsonparser.ArrayEach([]byte(page), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			Logging(err, "callback workWithResponse")
			return
		}
		purNum, err := jsonparser.GetString(value, "[0]")
		if err != nil {
			Logging(err)
			return
		}
		hrefT, err := jsonparser.GetString(value, "[1]")
		if err != nil {
			Logging(err)
			return
		}
		href, purName := findTwoFromRegExp(hrefT, `href=(.+?)>(.+)<`)
		href = fmt.Sprintf("http://mmk.ru%s", href)
		pubDateT, err := jsonparser.GetString(value, "[2]")
		if err != nil {
			Logging(err)
			return
		}
		pubDateT = strings.Replace(pubDateT, "(MSK)", "", -1)
		pubDate := getTimeMoscowLayout(pubDateT, "02.01.2006 15:04:05")

		endDateT, err := jsonparser.GetString(value, "[3]")
		if err != nil {
			Logging(err)
			return
		}
		endDateT = strings.Replace(endDateT, "(MSK)", "", -1)
		endDate := getTimeMoscowLayout(endDateT, "02.01.2006 15:04:05")

		person, err := jsonparser.GetString(value, "[4]")
		if err != nil {
			Logging(err)
			return
		}
		person = strip.StripTags(person)
		person = strings.Replace(person, "E-mail: ", " E-mail: ", 1)
		email := findFromRegExp(person, `E-mail: (.+)`)
		status, err := jsonparser.GetString(value, "[5]")
		if err != nil {
			Logging(err)
			return
		}
		tnd := TenderMmk2{purName: purName, purNum: purNum, url: href, pubDate: pubDate, endDate: endDate, status: status, contactPerson: person, email: email}
		t.Tender(tnd)
	}, "aaData")
	if err != nil {
		Logging(err, "sportItem tournaments")
		return
	}
}

func (t *ParserMmk2) Tender(tn TenderMmk2) {
	defer SaveStack()
}
