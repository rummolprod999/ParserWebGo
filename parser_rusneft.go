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

var AddtenderRusneft int
var UpdatetenderRusneft int

type ParserRusneft struct {
	TypeFz int
	Urls   []string
}

func (t *ParserRusneft) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	t.parsingPageAll()
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderRusneft))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderRusneft))
}

func (t *ParserRusneft) parsingPageAll() {
	for _, p := range t.Urls {
		t.parsingPage(p)
	}
}

func (t *ParserRusneft) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserRusneft) parsingTenderList(p string, url string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("table.tender-table tbody tr").Each(func(i int, s *goquery.Selection) {
		txt := s.Text()
		if !strings.Contains(txt, "Дата публикации приглашения") {
			t.parsingTenderFromList(s, url)
		}

	})
}

func (t *ParserRusneft) parsingTenderFromList(p *goquery.Selection, url string) {
	defer SaveStack()
	purNum := strings.TrimSpace(p.Find("span.tender-table__info-item_title").First().Text())
	if purNum == "" {
		Logging("Can not find purnum in ", url)
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
	dates := p.Find("span.tender-table__info-item")
	if len(dates.Nodes) < 3 {
		Logging("Can not find 3 dates in ", url)
		return
	}
	pubDateT := strings.TrimSpace(dates.Eq(0).Text())
	pubDate := getTimeMoscowLayout(pubDateT, "02.01.2006")
	endDateT := strings.TrimSpace(dates.Eq(1).Text())
	endDate := getTimeMoscowLayout(endDateT, "02.01.2006")
	endDateT = strings.TrimSpace(dates.Eq(2).Text())
	if endDateT != "" {
		endDate = getTimeMoscowLayout(endDateT, "02.01.2006")
	}
	if (pubDate == time.Time{} || endDate == time.Time{}) {
		Logging("Can not find enddate or startdate in ", url, purNum)
		return
	}
	purNum = fmt.Sprintf("%s_%s", purNum, pubDate.Format("2006-01-02"))
	t.Tender(purNum, url, pubDate, endDate, p)
}

func (t *ParserRusneft) Tender(purNum string, page string, pubDate time.Time, endDate time.Time, p *goquery.Selection) {
	defer SaveStack()
	purName := strings.TrimSpace(p.Find("td a").First().Text())
	db, err := sql.Open("mysql", Dsn)
	if err != nil {
		Logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	upDate := time.Now()
	idXml := purNum
	version := 1
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND doc_publish_date = ? AND type_fz = ?", Prefix))
	res, err := stmt.Query(purNum, pubDate, t.TypeFz)
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
	var updated = false
	if purNum != "" {
		stmt, err := db.Prepare(fmt.Sprintf("SELECT id_tender, date_version FROM %stender WHERE purchase_number = ? AND cancel=0 AND type_fz = ?", Prefix))
		rows, err := stmt.Query(purNum, t.TypeFz)
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
	hrefT := p.Find("td a").First()
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://www.russneft.ru%s", href)
	printForm := href
	idOrganizer := 0
	orgFullName := "ПАО НК «РуссНефть»"
	if true {
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
			email := "russneft@russneft.ru"
			phone := "(495) 411-6309"
			organizerINN := "7717133960"
			organizerKPP := "772901001"
			organizerPostAddress := "115054, г. Москва, ул. Пятницкая, д. 69"
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, inn = ?, kpp = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?", Prefix))
			res, err := stmt.Exec(orgFullName, organizerINN, organizerKPP, organizerPostAddress, organizerPostAddress, email, phone)
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
	etpName := orgFullName
	etpUrl := "http://www.russneft.ru/tenders"
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
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?", Prefix))
	rest, err := stmtt.Exec(idXml, purNum, pubDate, href, purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, endDate, cancelStatus, upDate, version, href, printForm, 0)
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		UpdatetenderRusneft++
	} else {
		AddtenderRusneft++
	}
	p.Find("td.tender-table__download-block > a").Each(func(i int, s *goquery.Selection) {
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
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, inn = ?, reg_num = ?, is223=1", Prefix))
			res, err := stmt.Exec(orgFullName, "7717133960", out)
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
	_, errr := stmtr.Exec(idLot, idCustomer, purName)
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

func (t *ParserRusneft) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.First().Text())
	href, exist := doc.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("http://www.russneft.ru%s", href)
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
