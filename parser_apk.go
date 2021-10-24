package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var addtenderApk int
var updatetenderApk int

type parserApk struct {
	TypeFz int
	Urls   []string
}

type tenderApk struct {
	purName string
	purNum  string
	url     string
	pwName  string
	pubDate time.Time
	endDate time.Time
}

func (t *parserApk) parsing() {
	defer SaveStack()
	logging("Start parsing")
	t.parsingPageAll()
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderApk))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderApk))
}

func (t *parserApk) parsingPageAll() {
	for i := 1; i <= 3; i++ {
		url := fmt.Sprintf("%s%d", t.Urls[0], i)
		t.parsingPage(url)
	}
}

func (t *parserApk) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserApk) parsingTenderList(p string, url string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		logging(err)
		return
	}
	doc.Find("table.table.table-striped tbody tr").Each(func(i int, s *goquery.Selection) {
		txt := s.Text()
		if !strings.Contains(txt, "Номер лота") {
			t.parsingTenderFromList(s, url)
		}
	})
}

func (t *parserApk) parsingTenderFromList(p *goquery.Selection, url string) {
	hrefT := p.Find("a.t_lot_title")
	href, exist := hrefT.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://tender-apk.ru/%s", href)
	purName := strings.TrimSpace(p.Find("a.t_lot_title").First().Text())
	if purName == "" {
		logging("cannot find purName in ", url)
		return
	}
	purNum := findFromRegExp(href, `LOT_ID=(\d{1,})`)
	if purNum == "" {
		logging("cannot find purnum in ", url)
		return
	}
	pwName := strings.TrimSpace(p.Find("div.t_lot_descr span.label-info").First().Text())
	PubDateT := strings.TrimSpace(p.Find("span.t_lot_start_date").First().Text())
	pubDate := getTimeMoscowLayout(PubDateT, "02.01.2006 15:04:05")
	if (pubDate == time.Time{}) {
		logging("cannot find pubDate in ", href, purNum)
		return
	}
	EndDateT := strings.TrimSpace(p.Find("span.t_lot_end_date").First().Text())
	endDate := getTimeMoscowLayout(EndDateT, "02.01.2006 15:04:05")
	if (endDate == time.Time{}) {
		logging("cannot find endDate in ", href, purNum)
		return
	}
	tnd := tenderApk{purNum: purNum, purName: purName, pwName: pwName, url: href, pubDate: pubDate, endDate: endDate}
	t.tender(tnd)
}
func (t *parserApk) tender(tn tenderApk) {
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
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ?", prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.endDate)
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
	r := DownloadPage(tn.url)
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
	orgName := strings.TrimSpace(doc.Find("td:contains('Организатор закупки:') + td").First().Text())
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
			contactPhone := ""
			contactEmail := ""
			inn := ""
			postAddress := ""
			contactPerson := ""
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
	idCustomer := 0
	cusName := strings.TrimSpace(doc.Find("td:contains('Заказчик(и):') + td").First().Text())
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
	etpName := "ООО «Заволжские просторы»"
	etpUrl := "http://tender-apk.ru/"
	IdEtp := 0
	IdEtp = getEtpId(etpName, etpUrl, db)
	idPlacingWay := 0
	if tn.pwName != "" {
		idPlacingWay = getPlacingWayId(tn.pwName, db)
	}
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?", prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.pubDate, tn.url, tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, tn.url, printForm, 0, "")
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderApk++
	} else {
		addtenderApk++
	}
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
	doc.Find("table.table.t_lot_table.tableScrollElNotEdit tbody tr").Each(func(i int, s *goquery.Selection) {
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
func (t *parserApk) purObj(idLot int, idCustomer int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	name1 := strings.TrimSpace(doc.Find("td:nth-child(2)").First().Text())
	name2 := strings.TrimSpace(doc.Find("td:nth-child(3)").First().Text())
	name := strings.TrimSpace(name1 + "\n" + name2)
	quantityT := strings.TrimSpace(doc.Find("td:nth-child(4)").First().Text())
	quantity := strings.TrimSpace(findFromRegExp(quantityT, `^([\d\s,]+)`))
	quantity = strings.Replace(quantity, ",", ".", -1)
	re := regexp.MustCompile(`\s+`)
	quantity = re.ReplaceAllString(quantity, "")
	okei := strings.TrimSpace(findFromRegExp(quantityT, `^[\d\s,]+(.+)`))
	stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?, quantity_value = ?, okei = ?, customer_quantity_value = ?", prefix))
	_, errr := stmtr.Exec(idLot, idCustomer, name, quantity, okei, quantity)
	stmtr.Close()
	if errr != nil {
		logging("Ошибка вставки purchase_object", errr)
		return
	}
	delivTerm1 := strings.TrimSpace(doc.Find("td:nth-child(5)").First().Text())
	delivTerm2 := strings.TrimSpace(doc.Find("td:nth-child(6)").First().Text())
	delivTerm := fmt.Sprintf("Период поставки «месяц»: %s \nПримечание: %s", delivTerm1, delivTerm2)
	stmtcr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer_requirement SET id_lot = ?, id_customer = ?, delivery_term = ?", prefix))
	_, err := stmtcr.Exec(idLot, idCustomer, delivTerm)
	stmtcr.Close()
	if err != nil {
		logging("Ошибка вставки purchase_object", err)
		return
	}

}
