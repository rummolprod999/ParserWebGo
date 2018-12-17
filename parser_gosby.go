package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var AddtenderGosBy int
var UpdatetenderGosBy int

type ParserGosBy struct {
	TypeFz int
	Urls   []string
}

type TenderGosBy struct {
	purNum     string
	url        string
	orgName    string
	lotPrice   string
	purObjInfo string
	currency   string
	status     string
	endDate    time.Time
}

func (t *ParserGosBy) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	t.parsingLastday()
	t.parsingToday()
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderGosBy))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderGosBy))
}
func (t *ParserGosBy) getCountPage(url string) int {
	c := 10
	u := fmt.Sprintf("%s%d", url, 1)
	defer SaveStack()
	r := DownloadPageWithUAIceTrade(u)
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
func (t *ParserGosBy) parsingLastday() {
	var today = time.Now()
	oneDay := time.Hour * -24
	yesterday := today.Add(oneDay)
	yesterdayS := yesterday.Format("02.01.2006")
	startUrl := fmt.Sprintf("http://www.goszakupki.by/search/auctions?auc_num=&search_text=&industries=&search_company=&search_participant=&price_from=&price_to=&created_from=%s&created_to=%s&request_end_from=&request_end_to=&auction_date_from=&auction_date_to=&auc_type[]=&auc_type[]=Auction&auc_type[]=Request&r[]=1&r[]=2&r[]=3&r[]=4&r[]=6&r[]=5&s[a]=1&s[b]=1&s[e]=1&s[p]=1&s[v]=1&s[c]=1&s[w]=1&s[s]=1&s[m]=1&s[ps]=1&searchBtn=Искать&goSearch=1&p=", yesterdayS, yesterdayS)
	countPage := t.getCountPage(startUrl)
	for i := 1; i <= countPage; i++ {
		ul := fmt.Sprintf("%s%d", startUrl, i)
		t.parsingPage(ul)
		time.Sleep(SleepingTime)
	}

}
func (t *ParserGosBy) parsingToday() {
	var today = time.Now()
	yesterday := today
	yesterdayS := yesterday.Format("02.01.2006")
	startUrl := fmt.Sprintf("http://www.goszakupki.by/search/auctions?auc_num=&search_text=&industries=&search_company=&search_participant=&price_from=&price_to=&created_from=%s&created_to=%s&request_end_from=&request_end_to=&auction_date_from=&auction_date_to=&auc_type[]=&auc_type[]=Auction&auc_type[]=Request&r[]=1&r[]=2&r[]=3&r[]=4&r[]=6&r[]=5&s[a]=1&s[b]=1&s[e]=1&s[p]=1&s[v]=1&s[c]=1&s[w]=1&s[s]=1&s[m]=1&s[ps]=1&searchBtn=Искать&goSearch=1&p=", yesterdayS, yesterdayS)
	countPage := t.getCountPage(startUrl)
	for i := 1; i <= countPage; i++ {
		ul := fmt.Sprintf("%s%d", startUrl, i)
		t.parsingPage(ul)
		time.Sleep(SleepingTime)
	}
}

func (t *ParserGosBy) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPageWithUAIceTrade(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserGosBy) parsingTenderList(p string, url string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("#auctions-list tbody tr[class^='rw']:nth-child(3n+1)").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)

	})

}
func (t *ParserGosBy) parsingTenderFromList(p *goquery.Selection, url string) {
	defer SaveStack()
	orgName := strings.TrimSpace(p.Find("td:nth-child(2)").First().Text())
	lotPrice := strings.TrimSpace(p.Find("td:nth-child(4) span").First().Text())
	lotPrice = delallwhitespace(lotPrice)
	currency := strings.TrimSpace(p.Find("td:nth-child(4)").First().Text())
	//fmt.Println(currency)
	currency = strings.TrimSpace(findFromRegExp(currency, `([^\d\s]+)$`))
	if lotPrice == "" {
		currency = ""
	}
	nextNode := p.Next()
	purNum := strings.TrimSpace(nextNode.Find("td:nth-child(2) strong").First().Text())
	if purNum == "" {
		Logging("Can not find purnum in ", url)
		return
	}
	if strings.HasPrefix(purNum, "№") {
		purNum = strings.Replace(purNum, "№", "", -1)
	}
	purObjInfo := strings.TrimSpace(nextNode.Find("td:nth-child(2) a").First().Text())
	hrefT := nextNode.Find("td:nth-child(2) a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", hrefT.Text())
		return
	}
	endDateTT := cleanString(strings.TrimSpace(nextNode.Find("td:nth-child(4)").First().Text()))
	endDateT := findFromRegExp(endDateTT, `(\d{2}\.\d{2}\.\d{4})`)
	endDate := getTimeMoscowLayoutIceTrade(endDateT, "02.01.2006")
	if (endDate == time.Time{}) {
		Logging("Can not find end_date in ", href, purNum)
		return
	}
	nextNode = nextNode.Next()
	status := strings.TrimSpace(nextNode.Find("td:nth-child(2)").First().Text())
	tnd := TenderGosBy{purNum: purNum, url: href, orgName: orgName, lotPrice: lotPrice, purObjInfo: purObjInfo, currency: currency, status: status, endDate: endDate}
	t.Tender(tnd)
}
func (t *ParserGosBy) Tender(tn TenderGosBy) {
	defer SaveStack()
	db, err := sql.Open("mysql", Dsn)
	if err != nil {
		Logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ? AND notice_version = ?", Prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.endDate, tn.status)
	stmt.Close()
	if err != nil {
		Logging("Ошибка выполения запроса", err)
		return
	}
	if res.Next() {
		//Logging("Такой тендер уже есть", TradeId)
		res.Close()
		return
	}
	res.Close()
	upDate := time.Now()
	var cancelStatus = 0
	var updated = false
	if tn.purNum != "" {
		stmt, err := db.Prepare(fmt.Sprintf("SELECT id_tender, date_version FROM %stender WHERE purchase_number = ? AND cancel=0 AND type_fz = ?", Prefix))
		rows, err := stmt.Query(tn.purNum, t.TypeFz)
		stmt.Close()
		if err != nil {
			Logging("Ошибка выполения запроса", err)
			return
		}
		for rows.Next() {
			updated = true
			var idTender int
			var dateVersion time.Time
			err = rows.Scan(&idTender, &dateVersion)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
				return
			}
			//fmt.Println(DateUpdated.Sub(dateVersion))
			if dateVersion.Sub(upDate) <= 0 {
				stmtupd, _ := db.Prepare(fmt.Sprintf("UPDATE %stender SET cancel=1 WHERE id_tender = ?", Prefix))
				_, err = stmtupd.Exec(idTender)
				stmtupd.Close()

			} else {
				cancelStatus = 1
			}

		}
		rows.Close()
	}
	time.Sleep(SleepingTime)
	r := DownloadPageWithUAIceTrade(tn.url)
	if r == "" {
		Logging("Получили пустую строку", tn.url)
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r))
	if err != nil {
		Logging(err)
		return
	}
	pubDateT := cleanString(strings.TrimSpace(doc.Find("table td:contains('Дата размещения приглашения') + td").First().Text()))
	pubDate := getTimeMoscowLayoutIceTrade(pubDateT, "02.01.2006")
	if (pubDate == time.Time{}) {
		pubDate = getTimeMoscowLayoutIceTrade(pubDateT, "02.01.2006")
	}
	if (pubDate == time.Time{}) {
		Logging("Can not find startdate in ", tn.url, tn.purNum)
		return
	}
	printForm := tn.url
	idOrganizer := 0
	organizerINN := ""
	if tn.orgName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_organizer FROM %sorganizer WHERE full_name = ?", Prefix))
		rows, err := stmt.Query(tn.orgName)
		stmt.Close()
		if err != nil {
			Logging("Ошибка выполения запроса", err)
			return
		}
		if rows.Next() {
			err = rows.Scan(&idOrganizer)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
				return
			}
			rows.Close()
		} else {
			rows.Close()
			email := ""
			phone := ""
			organizerINN = findFromRegExp(cleanString(strings.TrimSpace(doc.Find("table td:contains('УНП') + td").First().Text())), `(\d{9})`)
			organizerPostAddress := cleanString(strings.TrimSpace(doc.Find("table td:contains('Место нахождения') + td").First().Text()))
			contactPerson := cleanString(strings.TrimSpace(doc.Find("table td:contains('Фамилии') + td").First().Text()))
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, inn = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?, contact_person = ?", Prefix))
			res, err := stmt.Exec(tn.orgName, organizerINN, organizerPostAddress, organizerPostAddress, email, phone, contactPerson)
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки организатора", err)
				return
			}
			id, err := res.LastInsertId()
			idOrganizer = int(id)
		}
	}
	idPlacingWay := 0
	IdEtp := 0
	etpName := "РУП «НАЦИОНАЛЬНЫЙ ЦЕНТР МАРКЕТИНГА И КОНЪЮНКТУРЫ ЦЕН»"
	etpUrl := "http://www.goszakupki.by"
	IdEtp = getEtpId(etpName, etpUrl, db)
	idXml := tn.purNum
	version := 1
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?, bidding_date = ?, scoring_date = ?", Prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, pubDate, tn.url, tn.purObjInfo, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, "", printForm, 0, tn.status, time.Time{}, time.Time{})
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		UpdatetenderGosBy++
	} else {
		AddtenderGosBy++
	}
	var requirement []string
	qualrec := cleanString(strings.TrimSpace(doc.Find("table td:contains('Квалификационные') + td").First().Text()))
	if qualrec != "" {
		requirement = append(requirement, qualrec)
	}
	rec := cleanString(strings.TrimSpace(doc.Find("table td:contains('Требования к составу') + td").First().Text()))
	if rec != "" {
		requirement = append(requirement, rec)
	}
	doc.Find("td.af-files a").Each(func(i int, s *goquery.Selection) {
		t.documents(idTender, s, db)
	})
	idCustomer := 0
	if tn.orgName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name = ?", Prefix))
		rows, err := stmt.Query(tn.orgName)
		stmt.Close()
		if err != nil {
			Logging("Ошибка выполения запроса", err)
			return
		}
		if rows.Next() {
			err = rows.Scan(&idCustomer)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
				return
			}
			rows.Close()
		} else {
			rows.Close()
			out, err := exec.Command("uuidgen").Output()
			if err != nil {
				Logging("Ошибка генерации UUID", err)
				return
			}
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, reg_num = ?, is223=1, inn = ?", Prefix))
			res, err := stmt.Exec(tn.orgName, out, organizerINN)
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки заказчика", err)
				return
			}
			id, err := res.LastInsertId()
			idCustomer = int(id)
		}
	}
	doc.Find("#lots_list tr[id^='lotRow']").Each(t.parsingLots(tn, doc, db, idTender, requirement, idCustomer))
	e := TenderKwords(db, idTender)
	if e != nil {
		Logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, tn.purNum, t.TypeFz)
	if e1 != nil {
		Logging("Ошибка обработки AddVerNumber", e1)
	}
}

func (t *ParserGosBy) parsingLots(tn TenderGosBy, doc *goquery.Document, db *sql.DB, idTender int, requirement []string, idCustomer int) func(i int, s *goquery.Selection) {
	return func(i int, s *goquery.Selection) {
		defer SaveStack()
		lotId, exist := s.Attr("id")
		if !exist {
			Logging("The element can not have lotId attribute", s.Text())
			return
		}
		lotNum, err := strconv.Atoi(cleanString(strings.TrimSpace(s.Find("td:nth-child(1)").First().Text())))
		if err != nil {
			Logging("can not find lotNum", err)
			return
		}
		lotPriceT := strings.TrimSpace(s.Find("td:nth-child(3)").First().Text())
		formatter := fmt.Sprintf(",([\\d\\.]+)[ //s]+%s", tn.currency)
		lt := delallwhitespace(lotPriceT)
		lotPrice := findFromRegExp(lt, formatter)
		if lotPrice == "" {
			lotPrice = tn.lotPrice
		}
		finSource := strings.TrimSpace(doc.Find(fmt.Sprintf("#%s ~ tr th:contains('Источник финансирования') + td div", lotId)).First().Text())
		idLot := 0
		stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, max_price = ?, currency = ?, finance_source = ?", Prefix))
		resl, err := stmtl.Exec(idTender, lotNum, lotPrice, tn.currency, finSource)
		stmtl.Close()
		if err != nil {
			Logging("Ошибка вставки lot", err)
			return
		}
		id, _ := resl.LastInsertId()
		idLot = int(id)
		for _, rd := range requirement {
			stmtreq, _ := db.Prepare(fmt.Sprintf("INSERT INTO %srequirement SET id_lot =?, content = ?", Prefix))
			_, err := stmtreq.Exec(idLot, rd)
			stmtreq.Close()
			if err != nil {
				Logging("Ошибка вставки lot", err)
				return
			}
		}
		okpd2 := strings.TrimSpace(doc.Find(fmt.Sprintf("#%s ~ tr th:contains('ОКРБ') + td", lotId)).First().Text())
		quantv := delallwhitespace(strings.TrimSpace(s.Find("td:nth-child(3) span").First().Text()))
		nameL := cleanString(strings.TrimSpace(s.Find("td:nth-child(2)").First().Text()))
		stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?, okpd2_code = ?, sum = ?, quantity_value = ?, customer_quantity_value = ?", Prefix))
		_, errr := stmtr.Exec(idLot, idCustomer, nameL, okpd2, lotPrice, quantv, quantv)
		stmtr.Close()
		if errr != nil {
			Logging("Ошибка вставки purchase_object", errr)
			return
		}
		delivterm := strings.TrimSpace(doc.Find(fmt.Sprintf("#%s ~ tr th:contains('Срок поставки') + td", lotId)).First().Text())
		delivplace := strings.TrimSpace(doc.Find(fmt.Sprintf("#%s ~ tr th:contains('Место поставки товара') + td div", lotId)).First().Text())
		if delivterm != "" || delivplace != "" {
			stmtcr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer_requirement SET id_lot = ?, id_customer = ?, delivery_term = ?, delivery_place = ?, max_price = ?", Prefix))
			_, err := stmtcr.Exec(idLot, idCustomer, delivterm, delivplace, lotPrice)
			stmtcr.Close()
			if err != nil {
				Logging("Ошибка вставки purchase_object", errr)
				return
			}
		}

	}

}
func (t *ParserGosBy) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.First().Text())
	href, exist := doc.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("%s", href)
	if nameF != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", Prefix))
		_, err := stmt.Exec(idTender, nameF, href)
		stmt.Close()
		if err != nil {
			Logging("Ошибка вставки attachment", err)
			return
		}
	}
}
