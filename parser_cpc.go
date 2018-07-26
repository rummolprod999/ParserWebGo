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

var AddtenderCpc int
var UpdatetenderCpc int

type ParserCpc struct {
	TypeFz int
	Url    string
}

type TenderCpc struct {
	url    string
	purNum string
}

func (t *ParserCpc) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	r := DownloadPage(t.Url)
	if r != "" {
		t.parsingTenderList(r, t.Url)
	} else {
		Logging("Получили пустую строку", t.Url)
	}
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderCpc))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderCpc))
}

func (t *ParserCpc) parsingTenderList(p string, url string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("ul.tenders-list li").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)

	})
}

func (t *ParserCpc) parsingTenderFromList(p *goquery.Selection, url string) {
	purNumT := strings.TrimSpace(p.Find("a").First().Text())
	if purNumT == "" {
		Logging("Can not find purnumT in ", url)
		return
	}
	purNum := findFromRegExp(purNumT, `Тендер №\s(.+)`)
	if purNum == "" {
		Logging("Can not find purnum in ", url)
		return
	}
	hrefT := p.Find("a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://www.cpc.ru%s", href)
	tnd := TenderCpc{url: href, purNum: purNum}
	t.Tender(tnd)
}

func (t *ParserCpc) Tender(tn TenderCpc) {
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
	purObjInfo := strings.TrimSpace(doc.Find("div.cbq-layout-main p.first-paragraph").First().Text())
	if purObjInfo == "" {
		Logging("Can not find purObjInfo in ", tn.url)
		return
	}
	pubDateT := strings.TrimSpace(doc.Find("div:contains('Размещён:') + div").First().Text())
	if pubDateT == "" {
		Logging("Can not find pubDateT in ", tn.url)
		return
	}
	pubDate := getDateCpc(pubDateT)
	if (pubDate == time.Time{}) {
		Logging("Can not parse pubDate in ", tn.url)
		return
	}

	endDateT := strings.TrimSpace(doc.Find("div:contains('Приём заявок до:') + div").First().Text())
	if endDateT == "" {
		Logging("Can not find endDateT in ", tn.url)
		return
	}
	endDate := getDateCpc(endDateT)
	if (endDate == time.Time{}) {
		Logging("Can not parse endDate in ", tn.url)
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
	orgFullName := strings.TrimSpace(doc.Find("div:contains('Заказчик:') + div").First().Text())
	if orgFullName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_organizer FROM %sorganizer WHERE full_name = ?", Prefix))
		rows, err := stmt.Query(orgFullName)
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
			contactPerson := strings.TrimSpace(doc.Find("div:contains('Контактная информация:') + div").First().Text())
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, inn = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?, contact_person = ?", Prefix))
			res, err := stmt.Exec(orgFullName, organizerINN, organizerPostAddress, organizerPostAddress, email, phone, contactPerson)
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
	placinWayName := strings.TrimSpace(doc.Find("div:contains('Форма открытого тендера:') + div").First().Text())
	if placinWayName != "" {
		idPlacingWay = getPlacingWayId(placinWayName, db)
	}
	IdEtp := 0
	etpName := "Каспийский трубопроводный консорциум (КТК)"
	etpUrl := "http://www.cpc.ru/ru/tenders/Pages/default.aspx"
	IdEtp = getEtpId(etpName, etpUrl, db)
	idXml := tn.purNum
	version := 1
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?, bidding_date = ?", Prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, pubDate, tn.url, purObjInfo, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, endDate, cancelStatus, upDate, version, tn.url, printForm, 0, "", time.Time{})
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		UpdatetenderCpc++
	} else {
		AddtenderCpc++
	}
	doc.Find("div:contains('Приложения:') + div a").Each(func(i int, s *goquery.Selection) {
		t.documents(idTender, s, db)
	})
	idCustomer := 0
	if orgFullName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name = ?", Prefix))
		rows, err := stmt.Query(orgFullName)
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
			res, err := stmt.Exec(orgFullName, out, "")
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
	_, errr := stmtr.Exec(idLot, idCustomer, purObjInfo)
	stmtr.Close()
	if errr != nil {
		Logging("Ошибка вставки purchase_object", errr)
		return
	}
	e := TenderKwords(db, idTender)
	if e != nil {
		Logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, tn.purNum, t.TypeFz)
	if e1 != nil {
		Logging("Ошибка обработки AddVerNumber", e1)
	}
}

func (t *ParserCpc) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.First().Text())
	href, exist := doc.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("http://www.cpc.ru%s", href)
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
