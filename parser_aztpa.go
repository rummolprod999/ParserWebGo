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

var AddtenderAztpa int
var UpdatetenderAztpa int

type ParserAztpa struct {
	TypeFz int
	Urls   []string
}

type TenderAztpa struct {
	purName string
	purNum  string
	url     string
	pubDate time.Time
	endDate time.Time
}

func (t *ParserAztpa) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	t.parsingPageAll()
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderAztpa))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderAztpa))
}

func (t *ParserAztpa) parsingPageAll() {
	for _, p := range t.Urls {
		t.parsingPage(p)
	}
}

func (t *ParserAztpa) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserAztpa) parsingTenderList(p string, url string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("#table_zakupki_all tbody tr").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)
	})

}

func (t *ParserAztpa) parsingTenderFromList(p *goquery.Selection, url string) {
	hrefT := p.Find("td > a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("https://zakupki.aztpa.ru%s", href)
	purName := strings.TrimSpace(p.Find("td > a").First().Text())
	if purName == "" {
		Logging("cannot find purName in ", url)
		return
	}
	purNum := strings.TrimSpace(p.Find("td:first-of-type").First().Text())
	if purNum == "" {
		Logging("cannot find purNum in ", url)
		return
	}
	PubDateT := strings.TrimSpace(p.Find("td:nth-of-type(4)").First().Text())
	pubDate := getTimeMoscowLayout(PubDateT, "02.01.2006")
	if (pubDate == time.Time{}) {
		Logging("cannot find pubDate in ", href, purNum)
		return
	}
	EndDateT := strings.TrimSpace(p.Find("td:nth-of-type(5)").First().Text())
	endDate := getTimeMoscowLayout(EndDateT, "02.01.2006 15:04")
	if (endDate == time.Time{}) {
		Logging("cannot find endDate in ", href, purNum)
		return
	}
	tnd := TenderAztpa{purName: purName, purNum: purNum, url: href, pubDate: pubDate, endDate: endDate}
	t.Tender(tnd)

}

func (t *ParserAztpa) Tender(tn TenderAztpa) {
	defer SaveStack()
	db, err := sql.Open("mysql", Dsn)
	if err != nil {
		Logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	upDate := time.Now()
	idXml := tn.purNum
	version := 1
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ?", Prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.endDate)
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
	orgName := "АО НПО \"Тяжпромарматура\""
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
			contactPhone := "+7 (48753) 9-07-70, 2-04-80"
			contactEmail := "office@aztpa.ru"
			inn := "7111502104"
			postAddress := "301368, Россия, Тульская область, г. Алексин, ул. Некрасова, д. 60."
			contactPerson := ""
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, contact_person = ?, contact_phone = ?, contact_email = ?, inn = ?, post_address = ?", Prefix))
			res, err := stmt.Exec(orgName, contactPerson, contactPhone, contactEmail, inn, postAddress)
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки организатора", err)
				return
			}
			id, err := res.LastInsertId()
			idOrganizer = int(id)
		}
	}
	idCustomer := 0
	cusName := orgName
	if cusName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name = ?", Prefix))
		rows, err := stmt.Query(cusName)
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
			innCus := "7111502104"
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, reg_num = ?, is223=1, inn = ?", Prefix))
			res, err := stmt.Exec(cusName, out, innCus)
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки заказчика", err)
				return
			}
			id, err := res.LastInsertId()
			idCustomer = int(id)
		}
	}
	etpName := orgName
	etpUrl := "https://zakupki.aztpa.ru/"
	IdEtp := 0
	IdEtp = getEtpId(etpName, etpUrl, db)
	idPlacingWay := 0
	status := strings.TrimSpace(doc.Find("td:contains('Статус закупки') + td > span").First().Text())
	var purName string

	if purDesc := strings.TrimSpace(doc.Find("td:contains('Описание') + td > span").First().Text()); purDesc != "" {
		purName = fmt.Sprintf("%s (%s)", tn.purName, purDesc)
	} else {
		purName = tn.purName
	}
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?", Prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.pubDate, tn.url, purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, tn.url, printForm, 0, status)
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		UpdatetenderAztpa++
	} else {
		AddtenderAztpa++
	}
	doc.Find("td:contains('Правила проведения закупки') + td > span a").Each(func(i int, s *goquery.Selection) {
		t.documents(idTender, idCustomer, s, db)
	})
	doc.Find("td a[href *= '/zakupkilot/']").Each(func(i int, s *goquery.Selection) {
		t.lots(i+1, idTender, idCustomer, s, db)
	})
	e := TenderKwords(db, idTender)
	if e != nil {
		Logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, tn.purNum, t.TypeFz)
	if e1 != nil {
		Logging("Ошибка обработки AddVerNumber", e1)
	}

}

func (t *ParserAztpa) documents(idTender int, idCus int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.First().Text())
	href, exist := doc.Attr("href")
	if !exist {
		Logging("The element cannot have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("https://zakupki.aztpa.ru%s", href)
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

func (t *ParserAztpa) lots(numLot int, idTender int, idCustomer int, doc *goquery.Selection, db *sql.DB) {
	idLot := 0
	stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, currency = ?", Prefix))
	resl, err := stmtl.Exec(idTender, numLot, "")
	stmtl.Close()
	if err != nil {
		Logging("Ошибка вставки lot", err)
		return
	}
	id, _ := resl.LastInsertId()
	idLot = int(id)
	href, exist := doc.Attr("href")
	if !exist {
		Logging("The element cannot have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("https://zakupki.aztpa.ru%s", href)
	r := DownloadPage(href)
	if r == "" {
		Logging("Получили пустую строку", href)
		return
	}
	doclot, err := goquery.NewDocumentFromReader(strings.NewReader(r))
	if err != nil {
		Logging(err)
		return
	}
	purObj := strings.TrimSpace(doclot.Find("td:contains('Наименование лота') + td > span").First().Text())
	if purObj != "" {
		stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?", Prefix))
		_, errr := stmtr.Exec(idLot, idCustomer, purObj)
		stmtr.Close()
		if errr != nil {
			Logging("Ошибка вставки purchase_object", errr)
			return
		}
	}
	delivTerm1 := strings.TrimSpace(doclot.Find("td:contains('Требуемые сроки поставки') + td > span").First().Text())
	delivTerm2 := strings.TrimSpace(doclot.Find("td:contains('Общие условия оплаты') + td > span").First().Text())
	delivTerm3 := strings.TrimSpace(doclot.Find("td:contains('Особые требования') + td > span").First().Text())
	delivTerm := fmt.Sprintf("Требуемые сроки поставки: %s \nОбщие условия оплаты: %s \nОсобые требования: %s", delivTerm1, delivTerm2, delivTerm3)
	stmtcr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer_requirement SET id_lot = ?, id_customer = ?, delivery_term = ?", Prefix))
	_, errr := stmtcr.Exec(idLot, idCustomer, delivTerm)
	stmtcr.Close()
	if err != nil {
		Logging("Ошибка вставки purchase_object", errr)
		return
	}

}
