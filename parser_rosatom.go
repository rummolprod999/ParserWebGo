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

var addtenderRosAtom int
var updatetenderRosAtom int

type parserRosAtom struct {
	TypeFz int
	Urls   []string
}

type tenderRosAtom struct {
	purName string
	purNum  string
	url     string
	status  string
	pubDate time.Time
	endDate time.Time
}

func (t *parserRosAtom) parsing() {
	defer SaveStack()
	logging("Start parsing")
	t.parsingPageAll()
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderRosAtom))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderRosAtom))
}

func (t *parserRosAtom) parsingPageAll() {
	for _, p := range t.Urls {
		t.parsingPage(p)
	}
}

func (t *parserRosAtom) parsingPage(p string) {
	defer SaveStack()
	for i := 1; i <= 20; i++ {
		url := fmt.Sprintf("%s%d", p, i)
		t.downPage(url)
	}
}

func (t *parserRosAtom) downPage(p string) {
	defer SaveStack()
	r := DownloadPageWithUA(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserRosAtom) parsingTenderList(p string, url string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		logging(err)
		return
	}
	doc.Find("div.table-lots-list table tbody tr:nth-child(2n+1)").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)
	})
}

func (t *parserRosAtom) parsingTenderFromList(p *goquery.Selection, url string) {
	hrefT := p.Find("td.description > p a")
	href, exist := hrefT.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", url)
		return
	}
	href = fmt.Sprintf("https://zakupki.rosatom.ru%s", href)
	purName := strings.TrimSpace(p.Find("td.description > p a").First().Text())
	if purName == "" {
		logging("cannot find purName in ", url)
		return
	}
	purNum := strings.TrimSpace(p.Find("td.description p").First().Text())
	if purNum == "" {
		logging("cannot find purNum in ", url)
		return
	}
	PubDateT := strings.TrimSpace(p.Find("td:nth-of-type(6) > p").First().Text())
	pubDate := getTimeMoscowLayout(PubDateT, "02.01.2006")
	if (pubDate == time.Time{}) {
		logging("cannot find pubDate in ", href, purNum)
		return
	}
	EndDateT := strings.TrimSpace(p.Find("td:nth-of-type(7) > p").First().Text())
	EndDateT = findFromRegExp(EndDateT, `(\d{2}\.\d{2}\.\d{4})`)
	endDate := getTimeMoscowLayout(EndDateT, "02.01.2006")
	if (endDate == time.Time{}) {
		endDate = pubDate.AddDate(0, 0, 2)
	}
	status := strings.TrimSpace(p.Find("td:nth-of-type(8) > p").First().Text())
	tnd := tenderRosAtom{purName: purName, purNum: purNum, url: href, status: status, pubDate: pubDate, endDate: endDate}
	t.tender(tnd)
}

func (t *parserRosAtom) tender(tn tenderRosAtom) {
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
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ? AND notice_version = ?", prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.endDate, tn.status)
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
	orgName := strings.TrimSpace(doc.Find("td:contains('Наименование организации') + td > a").First().Text())
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
			contactPhone := strings.TrimSpace(doc.Find("td:contains('Телефон') + td").First().Text())
			contactEmail := strings.TrimSpace(doc.Find("td:contains('Адрес электронной почты') + td > a").First().Text())
			inn := ""
			postAddress := strings.TrimSpace(doc.Find("td:contains('Место нахождения') + td").First().Text())
			contactPerson := strings.TrimSpace(doc.Find("td:contains('Контактное лицо') + td").First().Text())
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, contact_person = ?, contact_phone = ?, contact_email = ?, inn = ?, post_address = ?", prefix))
			res, err := stmt.Exec(orgName, contactPerson, contactPhone, contactEmail, inn, postAddress)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки организатора", err)
				return
			}
			id, err := res.LastInsertId()
			idOrganizer = int(id)
		}
	}
	etpName := "Государственная корпорация по атомной энергии «Росатом»"
	etpUrl := "http://zakupki.rosatom.ru/"
	IdEtp := 0
	IdEtp = getEtpId(etpName, etpUrl, db)
	idPlacingWay := 0
	placinWayName := strings.TrimSpace(doc.Find("td:contains('Способ закупки') + td").First().Text())
	if placinWayName != "" {
		idPlacingWay = getPlacingWayId(placinWayName, db)
	}
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?", prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.pubDate, tn.url, tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, tn.url, printForm, 0, tn.status)
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderRosAtom++
	} else {
		addtenderRosAtom++
	}
	doc.Find("td a[title='Скачать']").Each(func(i int, s *goquery.Selection) {
		t.documents(idTender, s, db)
	})
	doc.Find("td a[href *= '&mode=lot']").Each(func(i int, s *goquery.Selection) {
		t.lots(i+1, idTender, s, db)
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
func (t *parserRosAtom) lots(numLot int, idTender int, d *goquery.Selection, db *sql.DB) {
	href, exist := d.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", d.Text())
		return
	}
	href = fmt.Sprintf("https://zakupki.rosatom.ru%s", href)
	r := DownloadPageWithUA(href)
	if r == "" {
		logging("Получили пустую строку", href)
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r))
	if err != nil {
		logging(err)
		return
	}
	nmck := strings.TrimSpace(doc.Find("td:contains('Начальная цена (руб.)') + td").First().Text())
	nmck = strings.Replace(nmck, "&nbsp;", "", -1)
	nmck = strings.Replace(nmck, ",", ".", -1)
	nmck = delallwhitespaceAtom(nmck)
	lotName := strings.TrimSpace(doc.Find("h1").First().Text())
	currency := strings.TrimSpace(doc.Find("td:contains('Валюта лота') + td").First().Text())
	stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, currency = ?, max_price = ?, lot_name = ?", prefix))
	resl, err := stmtl.Exec(idTender, numLot, currency, nmck, lotName)
	stmtl.Close()
	if err != nil {
		logging("Ошибка вставки lot", err)
		return
	}
	id, _ := resl.LastInsertId()
	idLot := int(id)
	idCustomer := 0
	cusName := strings.TrimSpace(doc.Find("td:contains('Официальное наименование') + td > a").First().Text())
	if cusName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name = ?", prefix))
		rows, err := stmt.Query(cusName)
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
			innCus := ""
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, reg_num = ?, is223=1, inn = ?", prefix))
			res, err := stmt.Exec(cusName, out, innCus)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки заказчика", err)
				return
			}
			id, err := res.LastInsertId()
			idCustomer = int(id)
		}
	}
	doc.Find("div.table-lots-list table tbody#custitemsbody_1 tr").Each(func(i int, s *goquery.Selection) {
		purObj := strings.TrimSpace(s.Find("td:nth-child(2) p").First().Text())
		if purObj != "" {
			stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?", prefix))
			_, errr := stmtr.Exec(idLot, idCustomer, purObj)
			stmtr.Close()
			if errr != nil {
				logging("Ошибка вставки purchase_object", errr)
				return
			}
		}
	})
}

func (t *parserRosAtom) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.First().Text())
	href, exist := doc.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("https://zakupki.rosatom.ru%s", href)
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
