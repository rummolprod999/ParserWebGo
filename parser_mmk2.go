package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
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
	println(val1, val2)
}
