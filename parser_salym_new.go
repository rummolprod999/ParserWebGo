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

type parserSalymNew struct {
	TypeFz int
}
type tenderSalymNew struct {
	purName string
	purNum  string
	url     string
}

func (t *parserSalymNew) parsing() {
	defer SaveStack()
	logging("Start parsing")
	t.parsingPageAll()
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderSalym))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderSalym))
}

func (t *parserSalymNew) parsingPageAll() {
	t.parsingPage("https://salympetroleum.ru/cp/tenders/")

}

func (t *parserSalymNew) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserSalymNew) parsingTenderList(p string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(p))
	if err != nil {
		logging(err)
		return
	}
	doc.Find("div.tender-item").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s)
	})
}

func (t *parserSalymNew) parsingTenderFromList(p *goquery.Selection) {
	defer SaveStack()
	purName := strings.TrimSpace(p.Find("span:contains('Наименование тендера:') + p").First().Text())
	if purName == "" {
		logging("cannot find purName in ")
		return
	}
	href := "https://salympetroleum.ru/cp/tenders/"
	purNum := GetMd5(purName)
	tnd := tenderSalymNew{purName: purName, purNum: purNum, url: href}
	t.tender(tnd, p)
}
func (t *parserSalymNew) tender(tn tenderSalymNew, doc *goquery.Selection) {
	pubDateTT := strings.TrimSpace(doc.Find("span:contains('Дата и время начала приема заявок:') + p").First().Text())
	if pubDateTT == "" {
		logging("cannot find pubDateTT", tn.url)
		return
	}
	dateT := findFromRegExp(pubDateTT, `(\d{2}.\d{2}.\d{4})`)
	timeT := findFromRegExp(pubDateTT, `(\d{2}:\d{2})`)
	pubDate := time.Time{}
	if dateT != "" && timeT != "" {
		pubDateS := fmt.Sprintf("%s %s", dateT, timeT)
		pubDate = getTimeMoscowLayout(pubDateS, "02.01.2006 15:04")
	} else if dateT != "" {
		pubDate = getTimeMoscowLayout(dateT, "02.01.2006")
	} else {
		pubDate = time.Time{}
	}
	if (pubDate == time.Time{}) {
		logging("cannot parse pubDate in ", tn.url)
		return
	}
	endDateTT := strings.TrimSpace(doc.Find("span:contains('Дата и время окончания приема заявок:') + p").First().Text())
	if endDateTT == "" {
		logging("cannot find endDateTT", tn.url)
		return
	}
	dateTend := findFromRegExp(endDateTT, `(\d{2}.\d{2}.\d{4})`)
	timeTend := findFromRegExp(endDateTT, `(\d{2}:\d{2})`)
	endDate := time.Time{}
	if dateTend != "" && timeTend != "" {
		endDateS := fmt.Sprintf("%s %s", dateTend, timeTend)
		endDate = getTimeMoscowLayout(endDateS, "02.01.2006 15:04")
	} else if dateTend != "" {
		endDate = getTimeMoscowLayout(dateTend, "02.01.2006")
	} else {
		endDate = time.Time{}
	}
	if (endDate == time.Time{}) {
		logging("cannot parse endDate in ", tn.url)
		return
	}
	scoringDateT := strings.TrimSpace(doc.Find("span:contains('Дата определения победителя и заключение договора:') + p").First().Text())
	scoringDate := getTimeMoscowLayout(scoringDateT, "02.01.2006")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ? AND doc_publish_date = ?", prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, endDate, pubDate)
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
	printForm := tn.url
	idOrganizer := 0
	organizerPostAddress := strings.TrimSpace(doc.Find("span:contains('Адрес:') + p").First().Text())
	idRegion := getRegionId(organizerPostAddress, db)
	orgName := strings.TrimSpace(doc.Find("span:contains('Организатор:') + p").First().Text())
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
			email := strings.TrimSpace(doc.Find("span:contains('Email:') + p > a").First().Text())
			phone := strings.TrimSpace(doc.Find("span:contains('Контактный телефон:') + p").First().Text())
			organizerINN := ""
			contactPerson := strings.TrimSpace(doc.Find("span:contains('Контактное лицо:') + p").First().Text())
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
	IdEtp := 0
	etpName := "Компания «Салым Петролеум Девелопмент Н.В.»"
	etpUrl := "https://salympetroleum.ru"
	IdEtp = getEtpId(etpName, etpUrl, db)
	idXml := tn.purNum
	version := 1
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?, bidding_date = ?, scoring_date = ?", prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, pubDate, tn.url, tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, endDate, cancelStatus, upDate, version, "", printForm, idRegion, "", time.Time{}, scoringDate)
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderSalym++
	} else {
		addtenderSalym++
	}
	doc.Find("a[href ^= '/upload/medialibrary/']").Each(func(i int, s *goquery.Selection) {
		t.documents(idTender, s, db)
	})
	var LotNumber = 1
	idLot := 0
	stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, currency = ?", prefix))
	resl, err := stmtl.Exec(idTender, LotNumber, "")
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
	e := TenderKwords(db, idTender)
	if e != nil {
		logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, tn.purNum, t.TypeFz)
	if e1 != nil {
		logging("Ошибка обработки AddVerNumber", e1)
	}
}

func (t *parserSalymNew) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.Text())
	href, exist := doc.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("https://salympetroleum.ru%s", href)
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
