package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"os/exec"
	"strings"
	"time"
)

var AddtenderKomtech int

type ParserKomtech struct {
	TypeFz int
	Url    string
}

type TenderKomtech struct {
	url        string
	purObjInfo string
	datePub    time.Time
	dateEnd    time.Time
	status     string
}

func (t *ParserKomtech) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	for i := 1; i <= 2; i++ {
		urllist := fmt.Sprintf("%s%d", t.Url, i)
		t.parsingPage(urllist)
	}
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderKomtech))
}
func (t *ParserKomtech) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage1251(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Получили пустую строку", p)
	}
}
func (t *ParserKomtech) parsingTenderList(p string, url string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("form[action='main.asp'] > table[cellpadding='4'] td").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)

	})
}

func (t *ParserKomtech) parsingTenderFromList(p *goquery.Selection, url string) {
	defer SaveStack()
	fullString := strings.TrimSpace(p.Find("span.txt_rem").First().Text())
	if fullString == "" {
		Logging("Can not find fullString in ", url)
		return
	}
	pubDateT := strings.TrimSpace(findFromRegExp(fullString, `^(\d{2}.\d{2}.\d{4})`))
	pubDate := getTimeMoscowLayout(pubDateT, "02.01.2006")
	endDateT := strings.TrimSpace(findFromRegExp(fullString, `Дата окончания приема заявок: (\d{2}.\d{2}.\d{4})`))
	endDate := getTimeMoscowLayout(endDateT, "02.01.2006")
	status := strings.TrimSpace(findFromRegExp(fullString, `Статус заявки: (.+)`))
	if (pubDate == time.Time{} || endDate == time.Time{}) {
		Logging("Can not find enddate or startdate in ", fullString)
		return
	}
	hrefT := p.Find("a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://zakupki.kom-tech.ru/%s", href)
	purName := strings.TrimSpace(p.Find("a").First().Text())
	tnd := TenderKomtech{url: href, purObjInfo: purName, datePub: pubDate, dateEnd: endDate, status: status}
	t.Tender(tnd)
}

func (t *ParserKomtech) Tender(tn TenderKomtech) {
	defer SaveStack()
	db, err := sql.Open("mysql", Dsn)
	if err != nil {
		Logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	r := DownloadPage1251(tn.url)
	if r == "" {
		Logging("Получили пустую строку", tn.url)
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r))
	if err != nil {
		Logging(err)
		return
	}
	purNum := strings.TrimSpace(doc.Find("table span.ttl2").First().Text())
	if purNum == "" {
		Logging("Can not find purnum in ", tn.url)
		return
	}
	upDate := time.Now()
	idXml := purNum
	version := 1
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND doc_publish_date = ? AND type_fz = ? AND end_date = ? AND notice_version = ?", Prefix))
	res, err := stmt.Query(purNum, tn.datePub, t.TypeFz, tn.dateEnd, tn.status)
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
	var cancelStatus = 0
	if purNum != "" {
		stmt, err := db.Prepare(fmt.Sprintf("SELECT id_tender, date_version FROM %stender WHERE purchase_number = ? AND cancel=0 AND type_fz = ?", Prefix))
		rows, err := stmt.Query(purNum, t.TypeFz)
		stmt.Close()
		if err != nil {
			Logging("Ошибка выполения запроса", err)
			return
		}
		for rows.Next() {
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
	orgName := "ООО \"Коммунальные технологии\""
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
			email := "adm@kom-tech.ru"
			phone := "(8352) 39-25-42"
			organizerINN := "2128051193"
			organizerKPP := "213050001"
			organizerPostAddress := "г. Чебоксары: Гаражный пр-д, д. 6/40"
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, inn = ?, kpp = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?", Prefix))
			res, err := stmt.Exec(orgName, organizerINN, organizerKPP, organizerPostAddress, organizerPostAddress, email, phone)
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
	etpUrl := "http://zakupki.kom-tech.ru"
	if true {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_etp FROM %setp WHERE name = ? AND url = ? LIMIT 1", Prefix))
		rows, err := stmt.Query(etpName, etpUrl)
		stmt.Close()
		if err != nil {
			Logging("Ошибка выполения запроса", err)
			return
		}
		if rows.Next() {
			err = rows.Scan(&IdEtp)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
				return
			}
			rows.Close()
		} else {
			rows.Close()
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %setp SET name = ?, url = ?, conf=0", Prefix))
			res, err := stmt.Exec(etpName, etpUrl)
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки etp", err)
				return
			}
			id, err := res.LastInsertId()
			IdEtp = int(id)
		}
	}
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?", Prefix))
	rest, err := stmtt.Exec(idXml, purNum, tn.datePub, tn.url, tn.purObjInfo, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.dateEnd, cancelStatus, upDate, version, tn.url, printForm, 0, tn.status)
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	AddtenderKomtech++
	doc.Find("b:contains('Основная информация') ~ b > a").Each(func(i int, s *goquery.Selection) {
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
			res, err := stmt.Exec(orgName, out, "2128051193")
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки заказчика", err)
				return
			}
			id, err := res.LastInsertId()
			idCustomer = int(id)
		}
	}
	var LotNumber = 1
	idLot := 0
	stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?", Prefix))
	resl, err := stmtl.Exec(idTender, LotNumber)
	stmtl.Close()
	if err != nil {
		Logging("Ошибка вставки lot", err)
		return
	}
	id, _ := resl.LastInsertId()
	idLot = int(id)
	stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?", Prefix))
	_, errr := stmtr.Exec(idLot, idCustomer, tn.purObjInfo)
	stmtr.Close()
	if errr != nil {
		Logging("Ошибка вставки purchase_object", errr)
		return
	}
	e := TenderKwords(db, idTender)
	if e != nil {
		Logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, purNum, t.TypeFz)
	if e1 != nil {
		Logging("Ошибка обработки AddVerNumber", e1)
	}
}

func (t *ParserKomtech) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.First().Text())
	href, exist := doc.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", doc.Text())
		return
	}
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
