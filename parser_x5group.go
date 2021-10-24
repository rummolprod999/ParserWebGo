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

var addtenderX5Group int
var updatetenderX5Group int

type parserX5Group struct {
	TypeFz int
	Urls   []string
}

func (t *parserX5Group) parsing() {
	defer SaveStack()
	logging("Start parsing")
	t.parsingPageAll()
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderX5Group))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderX5Group))
}

func (t *parserX5Group) parsingPageAll() {
	for _, p := range t.Urls {
		t.parsingPage(p)
	}
}
func (t *parserX5Group) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage1251(p)
	if r != "" {
		t.parsingTenderList(r)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserX5Group) parsingTenderList(p string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(p))
	if err != nil {
		logging(err)
		return
	}
	doc.Find("h2:contains('Объявленные') + #mytable tbody tr").Each(func(i int, s *goquery.Selection) {
		txt := s.Text()
		if !strings.Contains(txt, "Название конкурса") {
			t.parsingTenderFromList(s)
		}
	})
}

func (t *parserX5Group) parsingTenderFromList(p *goquery.Selection) {
	defer SaveStack()
	hrefT := p.Find("td[align='left'] a")
	href, exist := hrefT.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("https://tender.x5.ru%s", href)
	t.downloadTender(href)
}

func (t *parserX5Group) downloadTender(page string) {
	defer SaveStack()
	r := DownloadPage1251(page)
	if r != "" {
		doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(r))
		if err != nil {
			logging(err)
			return
		}
		t.parsingTender(doc, page)

	} else {
		logging("Получили пустую строку", page)
	}
}

func (t *parserX5Group) parsingTender(doc *goquery.Document, page string) {
	defer SaveStack()
	purNum := doc.Find("td:contains('Номер конкурса') + td").First().Text()
	purNum = strings.TrimSpace(purNum)
	if purNum == "" {
		logging("cannot find purchase number in ", page)
		return
	}
	pubDateT := strings.TrimSpace(doc.Find("td:contains('Дата и время начала подачи заявок на участие в конкурсе') + td").First().Text())
	endDateT := strings.TrimSpace(doc.Find("td:contains('Дата и время начала конкурса') + td").First().Text())
	if pubDateT == "" || endDateT == "" {
		logging("cannot find dates in ", page)
		return
	}
	pubDate := getTimeMoscow(pubDateT)
	endDate := getTimeMoscow(endDateT)
	if (pubDate == time.Time{}) || (endDate == time.Time{}) {
		logging("cannot parse dates in ", page)
		return
	}
	t.tender(purNum, page, pubDate, endDate, doc)
}
func (t *parserX5Group) tender(purNum string, page string, pubDate time.Time, endDate time.Time, doc *goquery.Document) {
	defer SaveStack()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	status := strings.TrimSpace(doc.Find("td:contains('Статус конкурса') + td").First().Text())
	upDate := time.Now()
	idXml := purNum
	version := 1
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND doc_publish_date = ? AND end_date = ? AND type_fz = ? AND notice_version = ?", prefix))
	res, err := stmt.Query(purNum, pubDate, endDate, t.TypeFz, status)
	stmt.Close()
	if err != nil {
		logging("Ошибка выполения запроса", err)
		return
	}
	if res.Next() {
		//Logging("Такой тендер уже есть", TradeId)
		res.Close()
		return
	}
	res.Close()
	var cancelStatus = 0
	var updated = false
	if purNum != "" {
		stmt, err := db.Prepare(fmt.Sprintf("SELECT id_tender, date_version FROM %stender WHERE purchase_number = ? AND cancel=0 AND type_fz = ?", prefix))
		rows, err := stmt.Query(purNum, t.TypeFz)
		stmt.Close()
		if err != nil {
			logging("Ошибка выполения запроса", err)
			return
		}
		for rows.Next() {
			updated = true
			var idTender int
			var dateVersion time.Time
			err = rows.Scan(&idTender, &dateVersion)
			if err != nil {
				logging("Ошибка чтения результата запроса", err)
				return
			}
			//fmt.Println(DateUpdated.Sub(dateVersion))
			if dateVersion.Sub(upDate) <= 0 {
				stmtupd, _ := db.Prepare(fmt.Sprintf("UPDATE %stender SET cancel=1 WHERE id_tender = ?", prefix))
				_, err = stmtupd.Exec(idTender)
				stmtupd.Close()

			} else {
				cancelStatus = 1
			}

		}
		rows.Close()

	}
	purchaseObjectInfo := strings.TrimSpace(doc.Find("td:contains('Название конкурса') + td").First().Text())
	printForm := page
	idOrganizer := 0
	orgFullName := "X5 Retail Group"
	if true {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_organizer FROM %sorganizer WHERE full_name = ?", prefix))
		rows, err := stmt.Query(orgFullName)
		stmt.Close()
		if err != nil {
			logging("Ошибка выполения запроса", err)
			return
		}
		if rows.Next() {
			err = rows.Scan(&idOrganizer)
			if err != nil {
				logging("Ошибка чтения результата запроса", err)
				return
			}
			rows.Close()
		} else {
			rows.Close()
			email := "tender.support@x5.ru"
			phone := "+7 (495) 502-91-37"
			organizerINN := "7733571872"
			organizerKPP := "773301001"
			organizerPostAddress := "РФ, 109029, Москва, ул. Средняя Калитниковская, д.28 стр.4"
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, inn = ?, kpp = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?", prefix))
			res, err := stmt.Exec(orgFullName, organizerINN, organizerKPP, organizerPostAddress, organizerPostAddress, email, phone)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки организатора", err)
				return
			}
			id, err := res.LastInsertId()
			idOrganizer = int(id)
		}
	}
	idPlacingWay := 0
	IdEtp := 0
	etpName := orgFullName
	etpUrl := "https://tender.x5.ru"
	if true {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_etp FROM %setp WHERE name = ? AND url = ? LIMIT 1", prefix))
		rows, err := stmt.Query(etpName, etpUrl)
		stmt.Close()
		if err != nil {
			logging("Ошибка выполения запроса", err)
			return
		}
		if rows.Next() {
			err = rows.Scan(&IdEtp)
			if err != nil {
				logging("Ошибка чтения результата запроса", err)
				return
			}
			rows.Close()
		} else {
			rows.Close()
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %setp SET name = ?, url = ?, conf=0", prefix))
			res, err := stmt.Exec(etpName, etpUrl)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки etp", err)
				return
			}
			id, err := res.LastInsertId()
			IdEtp = int(id)
		}
	}
	biddingDate := endDate
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, notice_version = ?, xml = ?, print_form = ?, bidding_date = ?, id_region = ?", prefix))
	rest, err := stmtt.Exec(idXml, purNum, pubDate, page, purchaseObjectInfo, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, endDate, cancelStatus, upDate, version, status, page, printForm, biddingDate, 0)
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderX5Group++
	} else {
		addtenderX5Group++
	}
	/*doc.Find("h2:contains('Конкурсная документация') ~ table:contains('Название файла') tbody tr").Each(func(i int, s *goquery.Selection) {
		txt := s.Text()
		if !strings.Contains(txt, "Название файла") {
			t.documents(idTender, s)
		}
	})*/
	idCustomer := 0
	if orgFullName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name = ?", prefix))
		rows, err := stmt.Query(orgFullName)
		stmt.Close()
		if err != nil {
			logging("Ошибка выполения запроса", err)
			return
		}
		if rows.Next() {
			err = rows.Scan(&idCustomer)
			if err != nil {
				logging("Ошибка чтения результата запроса", err)
				return
			}
			rows.Close()
		} else {
			rows.Close()
			out, err := exec.Command("uuidgen").Output()
			if err != nil {
				logging("Ошибка генерации UUID", err)
				return
			}
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, inn = ?, reg_num = ?, is223=1", prefix))
			res, err := stmt.Exec(orgFullName, "7733571872", out)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки заказчика", err)
				return
			}
			id, err := res.LastInsertId()
			idCustomer = int(id)
		}
	}
	var LotNumber = 1
	idLot := 0
	stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?", prefix))
	resl, err := stmtl.Exec(idTender, LotNumber)
	stmtl.Close()
	if err != nil {
		logging("Ошибка вставки lot", err)
		return
	}
	id, _ := resl.LastInsertId()
	idLot = int(id)
	stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?", prefix))
	_, errr := stmtr.Exec(idLot, idCustomer, purchaseObjectInfo)
	stmtr.Close()
	if errr != nil {
		logging("Ошибка вставки purchase_object", errr)
		return
	}
	e := TenderKwords(db, idTender)
	if e != nil {
		logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, purNum, t.TypeFz)
	if e1 != nil {
		logging("Ошибка обработки AddVerNumber", e1)
	}
}

func (t *parserX5Group) documents(idTender int, doc *goquery.Selection) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.Find("td[align='left']").First().Text())
	println(nameF)
}
