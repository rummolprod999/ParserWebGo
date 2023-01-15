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

var addtenderPhosagro int
var updatetenderPhosagro int

type parserPhosagro struct {
	TypeFz int
	Urls   []string
}

type tenderPhosagro struct {
	purNum     string
	url        string
	datePub    time.Time
	dateEnd    time.Time
	orgName    string
	status     string
	purObjInfo string
}

func (t *parserPhosagro) parsing() {
	defer SaveStack()
	logging("Start parsing")
	t.parsingPage("https://etpreg.phosagro.ru/tenders/?PAGEN_1=")
	t.parsingPage("https://etpreg.phosagro.ru/services?PAGEN_1=")
	for _, p := range t.Urls {
		for i := 2; i < 6; i++ {
			urllist := fmt.Sprintf("%s%d", p, i)
			t.parsingPage(urllist)
		}
	}

	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderPhosagro))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderPhosagro))
}

func (t *parserPhosagro) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPageWithUA(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserPhosagro) parsingTenderList(p string, url string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		logging(err)
		return
	}
	doc.Find("table.b-tbl-tenders tbody tr").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)

	})
}

func (t *parserPhosagro) parsingTenderFromList(p *goquery.Selection, url string) {
	purNum := strings.TrimSpace(p.Find("td:nth-child(1)").First().Text())
	if purNum == "" {
		logging("cannot find purnum in ", url)
		return
	}
	pubDateT := strings.TrimSpace(p.Find("td:nth-child(4)").First().Text())
	pubDate := getTimeMoscowLayout(pubDateT, "02.01.2006 15:04:05")
	if (pubDate == time.Time{}) {
		pubDate = getTimeMoscowLayout(pubDateT, "02.01.2006")
	}
	endDateT := strings.TrimSpace(p.Find("td:nth-child(5)").First().Text())
	endDate := getTimeMoscowLayout(endDateT, "02.01.2006 15:04:05")
	if (endDate == time.Time{}) {
		endDate = getTimeMoscowLayout(endDateT, "02.01.2006")
	}
	if (pubDate == time.Time{} || endDate == time.Time{}) {
		logging("cannot find enddate or startdate in ", url, purNum)
		return
	}
	orgName := strings.TrimSpace(p.Find("td:nth-child(3)").First().Text())
	purObjInfo := strings.TrimSpace(p.Find("td:nth-child(2) a").First().Text())
	status := strings.TrimSpace(p.Find("td:nth-child(7)").First().Text())
	hrefT := p.Find("td:nth-child(2) a")
	href, exist := hrefT.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("https://etpreg.phosagro.ru/tenders/%s", href)
	tnd := tenderPhosagro{purNum: purNum, url: href, datePub: pubDate, dateEnd: endDate, orgName: orgName, status: status, purObjInfo: purObjInfo}
	t.tender(tnd)

}
func (t *parserPhosagro) tender(tn tenderPhosagro) {
	defer SaveStack()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	upDate := time.Now()
	idXml := tn.purNum
	version := 1
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND doc_publish_date = ? AND type_fz = ? AND end_date = ? AND notice_version = ?", prefix))
	res, err := stmt.Query(tn.purNum, tn.datePub, t.TypeFz, tn.dateEnd, tn.status)
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
	r := DownloadPageWithUA(tn.url)
	if r == "" {
		logging("Получили пустую строку", tn.url)
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r))
	if err != nil {
		logging(err)
		return
	}
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
	if tn.orgName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_organizer FROM %sorganizer WHERE full_name = ?", prefix))
		rows, err := stmt.Query(tn.orgName)
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
			contactPerson := strings.TrimSpace(doc.Find("table.minimTable td:contains('Исполнитель:') + td").First().Text())
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, contact_person = ?", prefix))
			res, err := stmt.Exec(tn.orgName, contactPerson)
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
	etpName := "ЭТП «ФосАгро»"
	etpUrl := "https://etpreg.phosagro.ru/tenders"
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
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?", prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.datePub, tn.url, tn.purObjInfo, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.dateEnd, cancelStatus, upDate, version, tn.url, printForm, 0, tn.status)
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderPhosagro++
	} else {
		addtenderPhosagro++
	}
	idCustomer := 0
	if tn.orgName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name = ?", prefix))
		rows, err := stmt.Query(tn.orgName)
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
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, reg_num = ?, is223=1", prefix))
			res, err := stmt.Exec(tn.orgName, out)
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
	delivPlace := strings.TrimSpace(doc.Find("table.minimTable td:contains('Описание:') + td").First().Text())
	delivTerm1 := strings.TrimSpace(doc.Find("table.minimTable td:contains('Примечания поставщикам:') + td").First().Text())
	delivTerm2 := strings.TrimSpace(doc.Find("table.minimTable td:contains('Условия оплаты:') + td").First().Text())
	delivTerm3 := strings.TrimSpace(doc.Find("table.minimTable td:contains('Поставка до:') + td").First().Text())
	delivTerm := fmt.Sprintf("%s \nУсловия оплаты: %s \nПоставка до: %s", delivTerm1, delivTerm2, delivTerm3)
	stmtcr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer_requirement SET id_lot = ?, id_customer = ?, delivery_place = ?, delivery_term = ?", prefix))
	_, errr := stmtcr.Exec(idLot, idCustomer, delivPlace, delivTerm)
	stmtcr.Close()
	if errr != nil {
		logging("Ошибка вставки purchase_object", errr)
		return
	}
	doc.Find("table.tbl-tender_items tbody tr").Each(func(i int, s *goquery.Selection) {
		t.purObj(idLot, idCustomer, s, db)
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

func (t *parserPhosagro) purObj(idLot int, idCustomer int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	name := strings.TrimSpace(doc.Find("td:nth-child(2)").First().Text())
	quantity := strings.TrimSpace(doc.Find("span.b-pos-quantity").First().Text())
	okei := strings.TrimSpace(doc.Find("span.b-pos-uom").First().Text())
	stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?, quantity_value = ?, okei = ?, customer_quantity_value = ?", prefix))
	_, errr := stmtr.Exec(idLot, idCustomer, name, quantity, okei, quantity)
	stmtr.Close()
	if errr != nil {
		logging("Ошибка вставки purchase_object", errr)
		return
	}
}
