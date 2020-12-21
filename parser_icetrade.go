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

var AddtenderIcetrade int
var UpdatetenderIcetrade int
var SleepingTime = 7 * time.Second

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
	currency   string
}

func (t *ParserIcetrade) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	t.parsingLastday()
	t.parsingToday()
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderIcetrade))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderIcetrade))
}

func (t *ParserIcetrade) parsingLastday() {
	var today = time.Now()
	oneDay := time.Hour * -24
	yesterday := today.Add(oneDay)
	yesterdayS := yesterday.Format("02.01.2006")
	startUrl := fmt.Sprintf("http://www.icetrade.by/search/auctions?search_text=&zakup_type[1]=1&zakup_type[2]=1&auc_num=&okrb=&company_title=&establishment=0&industries=&period=&created_from=%s&created_to=%s&request_end_from=&request_end_to=&t[Trade]=1&t[eTrade]=1&t[socialOrder]=1&t[singleSource]=1&t[Auction]=1&t[Request]=1&t[contractingTrades]=1&t[negotiations]=1&t[Other]=1&r[1]=1&r[2]=2&r[7]=7&r[3]=3&r[4]=4&r[6]=6&r[5]=5&sort=num%%3Adesc&sbm=1&onPage=20&p=", yesterdayS, yesterdayS)
	countPage := t.getCountPage(startUrl)
	for i := 1; i <= countPage; i++ {
		ul := fmt.Sprintf("%s%d", startUrl, i)
		t.parsingPage(ul)
		time.Sleep(SleepingTime)
	}
}

func (t *ParserIcetrade) parsingToday() {
	var today = time.Now()
	yesterday := today
	yesterdayS := yesterday.Format("02.01.2006")
	startUrl := fmt.Sprintf("http://www.icetrade.by/search/auctions?search_text=&zakup_type[1]=1&zakup_type[2]=1&auc_num=&okrb=&company_title=&establishment=0&industries=&period=&created_from=%s&created_to=%s&request_end_from=&request_end_to=&t[Trade]=1&t[eTrade]=1&t[socialOrder]=1&t[singleSource]=1&t[Auction]=1&t[Request]=1&t[contractingTrades]=1&t[negotiations]=1&t[Other]=1&r[1]=1&r[2]=2&r[7]=7&r[3]=3&r[4]=4&r[6]=6&r[5]=5&sort=num%%3Adesc&sbm=1&onPage=20&p=", yesterdayS, yesterdayS)
	countPage := t.getCountPage(startUrl)
	for i := 1; i <= countPage; i++ {
		ul := fmt.Sprintf("%s%d", startUrl, i)
		t.parsingPage(ul)
		time.Sleep(SleepingTime)
	}
}
func (t *ParserIcetrade) getCountPage(url string) int {
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
func (t *ParserIcetrade) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPageWithUAIceTrade(p)
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
		time.Sleep(SleepingTime)
	})
}

func (t *ParserIcetrade) parsingTenderFromList(p *goquery.Selection, url string) {
	defer SaveStack()
	purNum := strings.TrimSpace(p.Find("td:nth-child(4)").First().Text())
	if purNum == "" {
		Logging("cannot find purnum in ", url)
		return
	}
	orgName := strings.TrimSpace(p.Find("td:nth-child(2)").First().Text())
	purObjInfo := strings.TrimSpace(p.Find("td:nth-child(1) a").First().Text())
	hrefT := p.Find("td:nth-child(1) a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	lotPrice := strings.TrimSpace(p.Find("td:nth-child(5) span").First().Text())
	lotPrice = delallwhitespace(lotPrice)
	currency := strings.TrimSpace(p.Find("td:nth-child(5)").First().Text())
	currency = strings.TrimSpace(findFromRegExp(currency, `([^\d\s]+)$`))
	tnd := TenderIcetrade{purNum: purNum, url: href, orgName: orgName, lotPrice: lotPrice, purObjInfo: purObjInfo, currency: currency}
	t.Tender(tnd)
}

func (t *ParserIcetrade) Tender(tn TenderIcetrade) {
	defer SaveStack()
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
		pubDate = getTimeMoscowLayoutIceTrade(pubDateT, "02.01.2006 15:04")
	}
	if (pubDate == time.Time{}) {
		Logging("cannot find startdate in ", tn.url, tn.purNum)
		return
	}
	endDateTT := cleanString(strings.TrimSpace(doc.Find("table td:contains('Дата и время окончания приема предложений') + td").First().Text()))
	endDateT := findFromRegExpDixy(endDateTT, `(\d{2}\.\d{2}\.\d{4}).+(\d{2}:\d{2})`)
	if endDateT == "" {
		endDateT = findFromRegExp(endDateTT, `(\d{2}\.\d{2}\.\d{4})`)
	}
	endDate := getTimeMoscowLayoutIceTrade(endDateT, "02.01.2006 15:04")
	if (endDate == time.Time{}) {
		endDate = getTimeMoscowLayoutIceTrade(endDateT, "02.01.2006")
	}
	if (endDate == time.Time{}) {
		Logging("cannot find enddate in ", tn.url, tn.purNum)
		return
	}
	db, err := sql.Open("mysql", Dsn)
	if err != nil {
		Logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ? AND doc_publish_date = ?", Prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, endDate, pubDate)
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
			organizerPostAddress := ""
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
	noticeVer := cleanString(strings.TrimSpace(doc.Find("table td:contains('Иные сведения') + td").First().Text()))
	idPlacingWay := 0
	IdEtp := 0
	etpName := "ICETRADE"
	etpUrl := "http://www.icetrade.by"
	IdEtp = getEtpId(etpName, etpUrl, db)
	idXml := tn.purNum
	version := 1
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?, bidding_date = ?, scoring_date = ?", Prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, pubDate, tn.url, tn.purObjInfo, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, endDate, cancelStatus, upDate, version, "", printForm, 0, noticeVer, time.Time{}, time.Time{})
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		UpdatetenderIcetrade++
	} else {
		AddtenderIcetrade++
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

func (t *ParserIcetrade) parsingLots(tn TenderIcetrade, doc *goquery.Document, db *sql.DB, idTender int, requirement []string, idCustomer int) func(i int, s *goquery.Selection) {
	return func(i int, s *goquery.Selection) {
		defer SaveStack()
		lotId, exist := s.Attr("id")
		if !exist {
			Logging("The element cannot have lotId attribute", s.Text())
			return
		}
		lotNum, err := strconv.Atoi(cleanString(strings.TrimSpace(s.Find("td:nth-child(1)").First().Text())))
		if err != nil {
			Logging("cannot find lotNum", err)
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
		okpd2 := strings.TrimSpace(doc.Find(fmt.Sprintf("#%s ~ tr th:contains('Код ОКРБ') + td div", lotId)).First().Text())
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

func (t *ParserIcetrade) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.First().Text())
	href, exist := doc.Attr("href")
	if !exist {
		Logging("The element cannot have href attribute", doc.Text())
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
