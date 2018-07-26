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

var AddtenderAzot int
var UpdatetenderAzot int

type ParserAzot struct {
	TypeFz int
	Url    string
}

type TenderAzot struct {
	purName string
	purNum  string
	status  string
	pWay    string
	noticeV string
	pubDate time.Time
	endDate time.Time
}

func (t *ParserAzot) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	t.parsingPageAll()
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderAzot))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderAzot))
}

func (t *ParserAzot) parsingPageAll() {
	for i := 1; i <= 10; i++ {
		url := fmt.Sprintf("%s%d", t.Url, i)
		t.parsingPage(url)
	}
}

func (t *ParserAzot) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserAzot) parsingTenderList(p string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("div.item-page > div.warpPurc").Each(func(i int, s *goquery.Selection) {
		txt := s.Text()
		if !strings.Contains(txt, "Название конкурса") {
			t.parsingTenderFromList(s)
		}
	})
}

func (t *ParserAzot) parsingTenderFromList(p *goquery.Selection) {
	defer SaveStack()
	purName := strings.TrimSpace(p.Find("div.line div:contains('Предмет') + div").First().Text())
	if purName == "" {
		Logging("The element can not have purNum", p.Text())
		return
	}
	pubDateT := strings.TrimSpace(p.Find("div.line div:contains('Дата объявления') + div").First().Text())
	if pubDateT == "" {
		Logging("Can not find pubDateT", purName)
		return
	}
	pubDate := getTimeMoscowLayout(pubDateT, "02.01.2006")
	if (pubDate == time.Time{}) {
		Logging("Can not parse pubDate in ", purName)
		return
	}
	endDateT := strings.TrimSpace(p.Find("div.line div:contains('Дата вскрытия') + div").First().Text())
	if endDateT == "" {
		Logging("Can not find endDateT in ", purName)
		return
	}
	endDate := getTimeMoscowLayout(endDateT, "02.01.2006")
	if (endDate == time.Time{}) {
		Logging("Can not parse endDate in ", purName)
		return
	}
	status := strings.TrimSpace(p.Find("div.line div:contains('Статус') + div").First().Text())
	pWay := strings.TrimSpace(p.Find("div.line div:contains('Форма торгов') + div").First().Text())
	purNum := strings.TrimSpace(p.Find("div.title b").First().Text())
	noticeV := strings.TrimSpace(p.Find("div.title + div.zamanda").First().Text())
	if purName == "" {
		Logging("The element can not have purNum", purName)
		return
	}
	tnd := TenderAzot{purName: purName, purNum: purNum, status: status, pWay: pWay, pubDate: pubDate, endDate: endDate, noticeV: noticeV}
	t.Tender(tnd)
}

func (t *ParserAzot) Tender(tn TenderAzot) {
	defer SaveStack()
	db, err := sql.Open("mysql", Dsn)
	if err != nil {
		Logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ? AND doc_publish_date = ? AND notice_version = ?", Prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.endDate, tn.pubDate, tn.noticeV)
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
	printForm := ""
	idOrganizer := 0
	orgName := "АО \"СДС Азот\""
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
			contactPerson := ""
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
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?, bidding_date = ?, scoring_date = ?", Prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.pubDate, "", tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, "", printForm, 0, tn.noticeV, time.Time{}, tn.endDate)
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		UpdatetenderAzot++
	} else {
		AddtenderAzot++
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
	stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?", Prefix))
	_, errr := stmtr.Exec(idLot, idCustomer, tn.purName)
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
