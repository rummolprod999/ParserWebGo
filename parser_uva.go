package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"os/exec"
	"strings"
	"time"
)

var AddtenderUva int
var UpdatetenderUva int

type ParserUva struct {
	TypeFz int
}

type TenderUva struct {
	purName string
	purNum  string
	url     string
	pubDate time.Time
	endDate time.Time
}

func (t *ParserUva) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	t.parsingPageAll()
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderUva))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderUva))
}

func (t *ParserUva) parsingPageAll() {
	t.parsingPage("http://tender.uva-moloko.ru/")
	for i := 20; i <= 100; i += 20 {
		t.parsingPage(fmt.Sprintf("http://tender.uva-moloko.ru/?view=list&layout=list&listtype=0&start=%d", i))
	}
}
func (t *ParserUva) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserUva) parsingTenderList(p string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("table.tenders_table tbody tr").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s)
	})
}

func (t *ParserUva) parsingTenderFromList(p *goquery.Selection) {
	defer SaveStack()
	purName := strings.TrimSpace(p.Find("td:nth-child(3) a").First().Text())
	if purName == "" {
		Logging("The element cannot have purName", p.Text())
		return
	}
	pubDateT := strings.TrimSpace(p.Find("td:nth-child(5)").First().Text())
	if pubDateT == "" {
		Logging("cannot find pubDateT", purName)
		return
	}
	pubDate := getTimeMoscowLayout(pubDateT, "02.01.2006 15:04:05")
	if (pubDate == time.Time{}) {
		Logging("cannot parse pubDate in ", purName)
		return
	}
	endDateT := strings.TrimSpace(p.Find("td:nth-child(6)").First().Text())
	if endDateT == "" {
		Logging("cannot find endDateT in ", purName)
		return
	}
	endDate := getTimeMoscowLayout(endDateT, "02.01.2006 15:04:05")
	if (endDate == time.Time{}) {
		Logging("cannot parse endDate in ", purName)
		return
	}
	hrefT := p.Find("td:nth-child(3) a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://tender.uva-moloko.ru%s", href)
	purNum := findFromRegExp(href, `id=(\d+)$`)
	if purNum == "" {
		Logging("The element cannot have purNum", href)
		return
	}
	tnd := TenderUva{purName: purName, purNum: purNum, url: href, pubDate: pubDate, endDate: endDate}
	t.Tender(tnd)

}
func (t *ParserUva) Tender(tn TenderUva) {
	defer SaveStack()
	db, err := sql.Open("mysql", Dsn)
	if err != nil {
		Logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ? AND doc_publish_date = ?", Prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.endDate, tn.pubDate)
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
	orgName := "ООО «Ува-молоко»"
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
			organizerINN := ""
			organizerPostAddress := ""
			contactPerson := strings.TrimSpace(strings.Replace(doc.Find("p:contains('Ответственный менеджер:')").First().Text(), "Ответственный менеджер:", "", -1))
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
	etpName := orgName
	etpUrl := "http://tender.uva-moloko.ru"
	IdEtp = getEtpId(etpName, etpUrl, db)
	idXml := tn.purNum
	version := 1
	noticeV := strings.TrimSpace(strings.Replace(doc.Find("p:contains('Описание лота:')").First().Text(), "Описание лота:", "", -1))
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?, bidding_date = ?, scoring_date = ?", Prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.pubDate, tn.url, tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, "", printForm, 0, noticeV, time.Time{}, time.Time{})
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		UpdatetenderUva++
	} else {
		AddtenderUva++
	}
	currency := strings.TrimSpace(strings.Replace(doc.Find("p:contains('Валюта:')").First().Text(), "Валюта:", "", -1))
	delivTerm1 := strings.TrimSpace(strings.Replace(doc.Find("p:contains('Условия поставки:')").First().Text(), "Условия поставки:", "", -1))
	delivTerm2 := strings.TrimSpace(strings.Replace(doc.Find("p:contains('Условия оплаты:')").First().Text(), "Условия оплаты:", "", -1))
	delivTerm := strings.TrimSpace(fmt.Sprintf("%s %s", delivTerm1, delivTerm2))
	hrefT := doc.Find("b:contains('Приложенные файлы:') ~ a")
	href, exist := hrefT.Attr("href")
	if exist {
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", Prefix))
		_, err := stmt.Exec(idTender, "Приложенные файлы:", href)
		stmt.Close()
		if err != nil {
			Logging("Ошибка вставки attachment", err)
			return
		}
	}
	var LotNumber = 1
	idLot := 0
	stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, currency = ?", Prefix))
	resl, err := stmtl.Exec(idTender, LotNumber, currency)
	stmtl.Close()
	if err != nil {
		Logging("Ошибка вставки lot", err)
		return
	}
	id, _ := resl.LastInsertId()
	idLot = int(id)
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
			res, err := stmt.Exec(orgName, out, "")
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки заказчика", err)
				return
			}
			id, err := res.LastInsertId()
			idCustomer = int(id)
		}
	}
	if delivTerm != "" {
		stmtcr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer_requirement SET id_lot = ?, id_customer = ?, delivery_term = ?", Prefix))
		_, errc := stmtcr.Exec(idLot, idCustomer, delivTerm)
		stmtcr.Close()
		if err != nil {
			Logging("Ошибка вставки purchase_object", errc)
			return
		}
	}
	doc.Find("table.tenders_table tbody tr").Each(func(i int, s *goquery.Selection) {
		pName := strings.TrimSpace(s.Find("td:nth-child(1)").First().Text())
		quantity := strings.TrimSpace(s.Find("td:nth-child(2)").First().Text())
		okei := strings.TrimSpace(s.Find("td:nth-child(3)").First().Text())
		price := strings.TrimSpace(s.Find("td:nth-child(4) span").First().Text())
		sum := strings.TrimSpace(s.Find("td:nth-child(5)").First().Text())
		stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?, quantity_value = ?, okei = ?, customer_quantity_value = ?, price = ?, sum = ?", Prefix))
		_, errr := stmtr.Exec(idLot, idCustomer, pName, quantity, okei, quantity, price, sum)
		stmtr.Close()
		if errr != nil {
			Logging("Ошибка вставки purchase_object", errr)
			return
		}

	})
	e := TenderKwords(db, idTender)
	if e != nil {
		Logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, tn.purNum, t.TypeFz)
	if e1 != nil {
		Logging("Ошибка обработки AddVerNumber", e1)
	}
}
