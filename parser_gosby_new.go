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

type ParserGosByNew struct {
	TypeFz int
	Url    string
}

type TenderGosByNew struct {
	purNum     string
	url        string
	purObjInfo string
	status     string
	endDate    time.Time
}

func (t *ParserGosByNew) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	for i := 1; i <= 20; i++ {
		url := fmt.Sprintf("%s%d", t.Url, i)
		t.parsingPage(url)
	}

	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderGosBy))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderGosBy))
}

func (t *ParserGosByNew) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPageWithUAIceTrade(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserGosByNew) parsingTenderList(p string, url string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("#w0 table tbody > tr").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)

	})
}

func (t *ParserGosByNew) parsingTenderFromList(p *goquery.Selection, url string) {
	defer SaveStack()
	purNum := strings.TrimSpace(p.Find("td:nth-child(1)").First().Text())
	if purNum == "" {
		Logging("Can not find purnum in ", url)
		return
	}
	purObjInfo := strings.TrimSpace(p.Find("td:nth-child(2) a").First().Text())
	hrefT := p.Find("td:nth-child(2) a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://goszakupki.by%s", href)
	endDateTT := cleanString(strings.TrimSpace(p.Find("td:nth-child(5)").First().Text()))
	endDateT := findFromRegExp(endDateTT, `(\d{2}\.\d{2}\.\d{4})`)
	endDate := getTimeMoscowLayoutIceTrade(endDateT, "02.01.2006")
	if (endDate == time.Time{}) {
		Logging("Can not find end_date in ", href, purNum)
		return
	}
	status := strings.TrimSpace(p.Find("td:nth-child(4) span").First().Text())
	tnd := TenderGosByNew{purNum: purNum, url: href, purObjInfo: purObjInfo, status: status, endDate: endDate}
	t.Tender(tnd)
}

func (t *ParserGosByNew) Tender(tn TenderGosByNew) {
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
	pubDateT := cleanString(strings.TrimSpace(doc.Find("table th:contains('Дата размещения приглашения') + td").First().Text()))
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
	orgName := cleanString(strings.TrimSpace(doc.Find("table th:contains('Наименование организации') + td").First().Text()))
	if orgName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_organizer FROM %sorganizer WHERE full_name = ?", Prefix))
		rows, err := stmt.Query(orgName)
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
			organizerINN = findFromRegExp(cleanString(strings.TrimSpace(doc.Find("table th:contains('УНП') + td").First().Text())), `(\d{9})`)
			organizerPostAddress := cleanString(strings.TrimSpace(doc.Find("table th:contains('Место нахождения') + td").First().Text()))
			contactPerson := cleanString(strings.TrimSpace(doc.Find("table th:contains('Фамилии') + td").First().Text()))
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, inn = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?, contact_person = ?", Prefix))
			res, err := stmt.Exec(orgName, organizerINN, organizerPostAddress, organizerPostAddress, email, phone, contactPerson)
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
	qualrec := cleanString(strings.TrimSpace(doc.Find("table th:contains('Квалификационные') + td").First().Text()))
	if qualrec != "" {
		requirement = append(requirement, qualrec)
	}
	rec := cleanString(strings.TrimSpace(doc.Find("table th:contains('Требования к участникам') + td").First().Text()))
	if rec != "" {
		requirement = append(requirement, rec)
	}
	doc.Find("a.modal-link").Each(func(i int, s *goquery.Selection) {
		t.documents(idTender, s, db)
	})
	idCustomer := 0
	if orgName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name = ?", Prefix))
		rows, err := stmt.Query(orgName)
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
			res, err := stmt.Exec(orgName, out, organizerINN)
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки заказчика", err)
				return
			}
			id, err := res.LastInsertId()
			idCustomer = int(id)
		}
	}
	LotPT := cleanString(strings.TrimSpace(doc.Find("table th:contains('Общая ориентировочная стоимость') + td").First().Text()))
	doc.Find("#lotsList tbody").Each(t.parsingLots(tn, doc, db, idTender, requirement, idCustomer, LotPT))
	e := TenderKwords(db, idTender)
	if e != nil {
		Logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, tn.purNum, t.TypeFz)
	if e1 != nil {
		Logging("Ошибка обработки AddVerNumber", e1)
	}
}

func (t *ParserGosByNew) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.First().Text())
	href, exist := doc.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("http://goszakupki.by%s", href)
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
func (t *ParserGosByNew) parsingLots(tn TenderGosByNew, doc *goquery.Document, db *sql.DB, idTender int, requirement []string, idCustomer int, LotPT string) func(i int, s *goquery.Selection) {
	return func(i int, s *goquery.Selection) {
		defer SaveStack()
		lotNum, err := strconv.Atoi(cleanString(strings.TrimSpace(s.Find("th.lot-num").First().Text())))
		if err != nil {
			Logging("can not find lotNum", err)
			return
		}

		lotPriceG := findFromRegExp(LotPT, `([\d ,.]+)`)
		lotPriceG = delallwhitespace(lotPriceG)
		currency := findFromRegExp(LotPT, `([^\d ,.]+)`)
		lotPriceT := strings.TrimSpace(s.Find("td.lot-count-price").First().Text())
		formatter := fmt.Sprintf(`,([\d \.]+)%s`, currency)
		lt := delallwhitespace(lotPriceT)
		lotPrice := findFromRegExp(lt, formatter)
		if lotPrice == "" {
			lotPrice = lotPriceG
		}
		finSource := strings.TrimSpace(doc.Find("b:contains('Источник финансирования:') + span").First().Text())
		idLot := 0
		stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, max_price = ?, currency = ?, finance_source = ?", Prefix))
		resl, err := stmtl.Exec(idTender, lotNum, lotPrice, currency, finSource)
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
		okpd2 := strings.TrimSpace(doc.Find("b:contains('Код предмета закупки по ОКРБ:') + span").First().Text())
		quantv := strings.TrimSpace(s.Find("td.lot-count-price").First().Text())
		nameL := cleanString(strings.TrimSpace(s.Find("td.lot-description").First().Text()))
		stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?, okpd2_code = ?, sum = ?, quantity_value = ?, customer_quantity_value = ?", Prefix))
		_, errr := stmtr.Exec(idLot, idCustomer, nameL, okpd2, lotPrice, quantv, quantv)
		stmtr.Close()
		if errr != nil {
			Logging("Ошибка вставки purchase_object", errr)
			return
		}
		delivterm := strings.TrimSpace(doc.Find("b:contains('Срок поставки:') + span").First().Text())
		delivplace := strings.TrimSpace(doc.Find("b:contains('Место поставки товара') + span").First().Text())
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
