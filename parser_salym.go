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

var AddtenderSalym int
var UpdatetenderSalym int

type ParserSalym struct {
	TypeFz int
}
type TenderSalym struct {
	purName string
	purNum  string
	url     string
}

func (t *ParserSalym) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	t.parsingPageAll()
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderSalym))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderSalym))
}

func (t *ParserSalym) parsingPageAll() {
	t.parsingPage("https://salympetroleum.ru/cp/tenders/")

}

func (t *ParserSalym) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserSalym) parsingTenderList(p string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("div.title-block > h2 > a").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s)
	})
}

func (t *ParserSalym) parsingTenderFromList(p *goquery.Selection) {
	defer SaveStack()
	purName := strings.TrimSpace(p.Text())
	if purName == "" {
		Logging("The element can not have purName", p.Text())
		return
	}
	href, exist := p.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", p.Text())
		return
	}
	href = fmt.Sprintf("https://salympetroleum.ru%s", href)
	purNum := findFromRegExp(href, `/tenders/(.+)/$`)
	tnd := TenderSalym{purName: purName, purNum: purNum, url: href}
	t.Tender(tnd)
}

func (t *ParserSalym) Tender(tn TenderSalym) {
	defer SaveStack()
	r := DownloadPage(tn.url)
	if r == "" {
		Logging("Получили пустую строку", tn.url)
		return
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(r))
	if err != nil {
		Logging(err)
		return
	}
	pubDateTT := strings.TrimSpace(doc.Find("span:contains('Дата и время начала приема заявок:') + p").First().Text())
	if pubDateTT == "" {
		Logging("Can not find pubDateTT", tn.url)
		return
	}
	dateT := findFromRegExp(pubDateTT, `(\d{2}.\d{2}.\d{4})`)
	timeT := findFromRegExp(pubDateTT, `(\d{2}:\d{2}:\d{2})`)
	pubDate := time.Time{}
	if dateT != "" && timeT != "" {
		pubDateS := fmt.Sprintf("%s %s", dateT, timeT)
		pubDate = getTimeMoscowLayout(pubDateS, "02.01.2006 15:04:05")
	} else if dateT != "" {
		pubDate = getTimeMoscowLayout(dateT, "02.01.2006")
	} else {
		pubDate = time.Time{}
	}
	if (pubDate == time.Time{}) {
		Logging("Can not parse pubDate in ", tn.url)
		return
	}
	endDateTT := strings.TrimSpace(doc.Find("span:contains('Дата и время окончания приема заявок:') + p").First().Text())
	if endDateTT == "" {
		Logging("Can not find endDateTT", tn.url)
		return
	}
	dateTend := findFromRegExp(endDateTT, `(\d{2}.\d{2}.\d{4})`)
	timeTend := findFromRegExp(endDateTT, `(\d{2}:\d{2}:\d{2})`)
	endDate := time.Time{}
	if dateTend != "" && timeTend != "" {
		endDateS := fmt.Sprintf("%s %s", dateTend, timeTend)
		endDate = getTimeMoscowLayout(endDateS, "02.01.2006 15:04:05")
	} else if dateTend != "" {
		endDate = getTimeMoscowLayout(dateTend, "02.01.2006")
	} else {
		endDate = time.Time{}
	}
	if (endDate == time.Time{}) {
		Logging("Can not parse endDate in ", tn.url)
		return
	}
	scoringDateT := strings.TrimSpace(doc.Find("span:contains('Дата определения победителя и заключение договора:') + p").First().Text())
	scoringDate := getTimeMoscowLayout(scoringDateT, "02.01.2006")
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
	organizerPostAddress := strings.TrimSpace(doc.Find("span:contains('Адрес:') + p").First().Text())
	idRegion := getRegionId(organizerPostAddress, db)
	orgName := strings.TrimSpace(doc.Find("span:contains('Организатор:') + p").First().Text())
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
			email := strings.TrimSpace(doc.Find("span:contains('Email:') + p > a").First().Text())
			phone := strings.TrimSpace(doc.Find("span:contains('Контактный телефон:') + p").First().Text())
			organizerINN := ""
			contactPerson := strings.TrimSpace(doc.Find("span:contains('Контактное лицо:') + p").First().Text())
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
	etpName := "Компания «Салым Петролеум Девелопмент Н.В.»"
	etpUrl := "https://salympetroleum.ru"
	IdEtp = getEtpId(etpName, etpUrl, db)
	idXml := tn.purNum
	version := 1
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?, bidding_date = ?, scoring_date = ?", Prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, pubDate, tn.url, tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, endDate, cancelStatus, upDate, version, "", printForm, idRegion, "", time.Time{}, scoringDate)
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		UpdatetenderSalym++
	} else {
		AddtenderSalym++
	}
	doc.Find("a[href ^= '/upload/medialibrary/']").Each(func(i int, s *goquery.Selection) {
		t.documents(idTender, s, db)
	})
	var LotNumber = 1
	idLot := 0
	stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, currency = ?", Prefix))
	resl, err := stmtl.Exec(idTender, LotNumber, "")
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

func (t *ParserSalym) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.Text())
	href, exist := doc.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("https://salympetroleum.ru%s", href)
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
