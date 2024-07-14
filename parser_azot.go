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

var addtenderAzot int
var updatetenderAzot int

type parserAzot struct {
	TypeFz int
	Url    string
}

type tenderAzot struct {
	purName string
	purNum  string
	status  string
	pWay    string
	noticeV string
	pubDate time.Time
	endDate time.Time
}

func (t *parserAzot) parsing() {
	defer SaveStack()
	logging("Start parsing")
	t.parsingPageAll()
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderAzot))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderAzot))
}

func (t *parserAzot) parsingPageAll() {
	for i := 1; i <= 30; i++ {
		url := fmt.Sprintf("%s%d", t.Url, i)
		t.parsingPage(url)
	}
}

func (t *parserAzot) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPageAzot(p)
	if r != "" {
		t.parsingTenderList(r)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserAzot) parsingTenderList(p string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(p))
	if err != nil {
		logging(err)
		return
	}
	doc.Find("div.item-page > div.warpPurc").Each(func(i int, s *goquery.Selection) {
		txt := s.Text()
		if !strings.Contains(txt, "Название конкурса") {
			t.parsingTenderFromList(s)
		}
	})
}

func (t *parserAzot) parsingTenderFromList(p *goquery.Selection) {
	defer SaveStack()
	purName := strings.TrimSpace(p.Find("div.line div:contains('Предмет') + div").First().Text())
	if purName == "" {
		logging("The element cannot have purNum", p.Text())
		return
	}
	pubDateT := strings.TrimSpace(p.Find("div.line div:contains('Дата объявления') + div").First().Text())
	if pubDateT == "" {
		logging("cannot find pubDateT", purName)
		return
	}
	pubDate := getTimeMoscowLayout(pubDateT, "02.01.2006")
	if (pubDate == time.Time{}) {
		logging("cannot parse pubDate in ", purName)
		return
	}
	endDateT := strings.TrimSpace(p.Find("div.line div:contains('Дата вскрытия') + div").First().Text())
	if endDateT == "" {
		logging("cannot find endDateT in ", purName)
		return
	}
	endDate := getTimeMoscowLayout(endDateT, "02.01.2006")
	if (endDate == time.Time{}) {
		endDate = pubDate.AddDate(0, 0, 2)
	}
	status := strings.TrimSpace(p.Find("div.line div:contains('Статус') + div").First().Text())
	pWay := strings.TrimSpace(p.Find("div.line div:contains('Форма торгов') + div").First().Text())
	purNum := strings.TrimSpace(p.Find("div.title b").First().Text())
	noticeV := strings.TrimSpace(p.Find("div.title + div.zamanda").First().Text())
	if purName == "" {
		logging("The element cannot have purNum", purName)
		return
	}
	tnd := tenderAzot{purName: purName, purNum: purNum, status: status, pWay: pWay, pubDate: pubDate, endDate: endDate, noticeV: noticeV}
	t.tender(tnd, p)
}

func (t *parserAzot) tender(tn tenderAzot, p *goquery.Selection) {
	defer SaveStack()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ? AND doc_publish_date = ? AND notice_version = ?", prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.endDate, tn.pubDate, tn.noticeV)
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
	upDate := time.Now()
	var cancelStatus = 0
	var updated = false
	if tn.purNum != "" {
		stmt, err := db.Prepare(fmt.Sprintf("SELECT id_tender, date_version FROM %stender WHERE purchase_number = ? AND cancel=0 AND type_fz = ?", prefix))
		rows, err := stmt.Query(tn.purNum, t.TypeFz)
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
	printForm := ""
	idOrganizer := 0
	orgName := "АО \"СДС Азот\""
	if orgName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_organizer FROM %sorganizer WHERE full_name = ?", prefix))
		rows, err := stmt.Query(orgName)
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
			email := ""
			phone := ""
			organizerINN := ""
			organizerPostAddress := ""
			contactPerson := ""
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, inn = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?, contact_person = ?", prefix))
			res, err := stmt.Exec(orgName, organizerINN, organizerPostAddress, organizerPostAddress, email, phone, contactPerson)
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
	if tn.pWay != "" {
		idPlacingWay = getPlacingWayId(tn.pWay, db)
	}
	IdEtp := 0
	etpName := orgName
	etpUrl := "http://zakupki.sbu-azot.ru/"
	IdEtp = getEtpId(etpName, etpUrl, db)
	idXml := tn.purNum
	version := 1
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?, bidding_date = ?, scoring_date = ?", prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.pubDate, "", tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, "", printForm, 0, tn.noticeV, time.Time{}, tn.endDate)
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderAzot++
	} else {
		addtenderAzot++
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
	idCustomer := 0
	if orgName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name = ?", prefix))
		rows, err := stmt.Query(orgName)
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
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, reg_num = ?, is223=1, inn = ?", prefix))
			res, err := stmt.Exec(orgName, out, "")
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки заказчика", err)
				return
			}
			id, err := res.LastInsertId()
			idCustomer = int(id)
		}
	}
	stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?", prefix))
	_, errr := stmtr.Exec(idLot, idCustomer, tn.purName)
	stmtr.Close()
	if errr != nil {
		logging("Ошибка вставки purchase_object", errr)
		return
	}
	p.Find("div.textLine > a").Each(func(i int, s *goquery.Selection) {
		t.documents(idTender, s, db)
	})
	e := TenderKwords(db, idTender)
	if e != nil {
		logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, tn.purNum, t.TypeFz)
	if e1 != nil {
		logging("Ошибка обработки AddVerNumber", e1)
	}
}

func (t *parserAzot) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.First().Text())
	href, exist := doc.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("http://zakupki.sbu-azot.ru%s", href)
	if nameF != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", prefix))
		_, err := stmt.Exec(idTender, nameF, href)
		stmt.Close()
		if err != nil {
			logging("Ошибка вставки attachment", err)
			return
		}
	}
}
