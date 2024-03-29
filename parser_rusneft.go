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

var addtenderRusneft int
var updatetenderRusneft int

type parserRusneft struct {
	TypeFz int
	Urls   []string
}

func (t *parserRusneft) parsing() {
	defer SaveStack()
	logging("Start parsing")
	t.parsingPageAll()
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderRusneft))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderRusneft))
}

func (t *parserRusneft) parsingPageAll() {
	for _, p := range t.Urls {
		for i := 1; i < 3; i++ {
			urllist := fmt.Sprintf("%s%d", p, i)
			t.parsingPage(urllist)
		}
	}
}

func (t *parserRusneft) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPageInsecure(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserRusneft) parsingTenderList(p string, url string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		logging(err)
		return
	}
	doc.Find("div.tenders-list table.table tbody tr").Each(func(i int, s *goquery.Selection) {
		txt := s.Text()
		if !strings.Contains(txt, "Дата публикации приглашения") {
			t.parsingTenderFromList(s, url)
		}

	})
}

func (t *parserRusneft) parsingTenderFromList(p *goquery.Selection, url string) {
	defer SaveStack()
	purNum := strings.TrimSpace(p.Find("td:nth-child(1)").First().Text())
	if purNum == "" {
		logging("cannot find purnum in ", url)
		return
	}
	if strings.Contains(url, "tenders/russneft") {
		purNum = fmt.Sprintf("%s_russneft", purNum)
	} else if strings.Contains(url, "tenders/all/zapsibgroop") {
		purNum = fmt.Sprintf("%s_zapsibgroop", purNum)
	} else if strings.Contains(url, "tenders/all/centrsibgroop") {
		purNum = fmt.Sprintf("%s_centrsibgroop", purNum)
	} else if strings.Contains(url, "tenders/all/volgagroop") {
		purNum = fmt.Sprintf("%s_volgagroop", purNum)
	} else if strings.Contains(url, "tenders/all/belarus") {
		purNum = fmt.Sprintf("%s_belarus", purNum)
	} else if strings.Contains(url, "tenders/all/overseas") {
		purNum = fmt.Sprintf("%s_overseas", purNum)
	}
	pubDateT := strings.TrimSpace(p.Find("td:nth-child(4)").First().Text())
	pubDate := getTimeMoscowLayout(pubDateT, "02.01.2006")
	if (pubDate == time.Time{}) {
		pubDate = getTimeMoscowLayout(pubDateT, "02.01.2006")
	}
	endDateT := strings.TrimSpace(p.Find("td:nth-child(5)").First().Text())
	endDate := getTimeMoscowLayout(endDateT, "02.01.2006")
	if (endDate == time.Time{}) {
		endDate = getTimeMoscowLayout(endDateT, "02.01.2006")
	}
	if (pubDate == time.Time{} || endDate == time.Time{}) {
		logging("cannot find enddate or startdate in ", url, purNum)
		return
	}
	if (pubDate == time.Time{} || endDate == time.Time{}) {
		logging("cannot find enddate or startdate in ", url, purNum)
		return
	}
	purNum = fmt.Sprintf("%s_%s", purNum, pubDate.Format("2006-01-02"))
	t.tender(purNum, url, pubDate, endDate, p)
}

func (t *parserRusneft) tender(purNum string, page string, pubDate time.Time, endDate time.Time, p *goquery.Selection) {
	defer SaveStack()
	purName := strings.TrimSpace(p.Find("td:nth-child(7) a").First().Text())
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	upDate := time.Now()
	idXml := purNum
	version := 1
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND doc_publish_date = ? AND type_fz = ?", prefix))
	res, err := stmt.Query(purNum, pubDate, t.TypeFz)
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
	hrefT := p.Find("td:nth-child(7) a").First()
	href, exist := hrefT.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("https://tender.russneft.ru%s", href)
	printForm := href
	idOrganizer := 0
	orgFullName := "ПАО НК «РуссНефть»"
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
			email := "russneft@russneft.ru"
			phone := "(495) 411-6309"
			organizerINN := "7717133960"
			organizerKPP := "772901001"
			organizerPostAddress := "115054, г. Москва, ул. Пятницкая, д. 69"
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
	pw := strings.TrimSpace(p.Find("td:nth-child(8)").First().Text())
	if pw != "" {
		idPlacingWay = getPlacingWayId(pw, db)
	}
	IdEtp := 0
	etpName := orgFullName
	etpUrl := "http://www.russneft.ru/tenders"
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
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?", prefix))
	rest, err := stmtt.Exec(idXml, purNum, pubDate, href, purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, endDate, cancelStatus, upDate, version, href, printForm, 0)
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderRusneft++
	} else {
		addtenderRusneft++
	}
	p.Find("td:nth-child(9) > a").Each(func(i int, s *goquery.Selection) {
		t.documents(idTender, s, db)
	})
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
			res, err := stmt.Exec(orgFullName, "7717133960", out)
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
	_, errr := stmtr.Exec(idLot, idCustomer, purName)
	stmtr.Close()
	if errr != nil {
		logging("Ошибка вставки purchase_object", errr)
		return
	}
	delivplace := strings.TrimSpace(p.Find("td:nth-child(3)").First().Text())
	if delivplace != "" {
		stmtcr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer_requirement SET id_lot = ?, id_customer = ?, delivery_term = ?, delivery_place = ?, max_price = ?", prefix))
		_, err := stmtcr.Exec(idLot, idCustomer, "", delivplace, "")
		stmtcr.Close()
		if err != nil {
			logging("Ошибка вставки purchase_object", errr)
			return
		}
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

func (t *parserRusneft) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.First().Text())
	href, exist := doc.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("https://tender.russneft.ru%s", href)
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
