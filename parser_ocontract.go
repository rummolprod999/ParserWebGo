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

var AddtenderOcontract int
var UpdatetenderOcontract int

type ParserOcontract struct {
	TypeFz int
}

type TenderOcontract struct {
	url        string
	purObjInfo string
	purNum     string
	dateEnd    time.Time
}

func (t *ParserOcontract) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	for i := 1; i <= 20; i++ {
		urllist := fmt.Sprintf("http://onlinecontract.ru/tenders?q=&o=&TorgType=0&TorgStatus=1&perpage=100&page=%d&accurate=0", i)
		t.parsingPage(urllist)
	}
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderOcontract))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderOcontract))
}

func (t *ParserOcontract) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserOcontract) parsingTenderList(p string, url string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("div.table-responsive > table.table-hover > tbody > tr").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)

	})
}

func (t *ParserOcontract) parsingTenderFromList(p *goquery.Selection, url string) {
	purNum := strings.TrimSpace(p.Find("td:nth-child(1) > a").First().Text())
	if purNum == "" {
		Logging("Can not find purnum in ", url)
		return
	}
	purObjInfo := strings.TrimSpace(p.Find("td:nth-child(3) a").First().Text())
	hrefT := p.Find("td:nth-child(3) a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://onlinecontract.ru%s", href)
	endDateT := strings.TrimSpace(p.Find("td:nth-child(4) span").First().Text())
	endDateT = strings.Replace(endDateT, "(мск)", "", -1)
	re := regexp.MustCompile(`\s+`)
	endDateT = re.ReplaceAllString(endDateT, " ")
	endDate := getTimeMoscowLayout(endDateT, "15:04 02.01.2006")
	if (endDate == time.Time{}) {
		Logging("Can not find enddate in ", url, purNum)
		return
	}
	tnd := TenderOcontract{url: href, purObjInfo: purObjInfo, dateEnd: endDate, purNum: purNum}
	t.Tender(tnd)
}

func (t *ParserOcontract) Tender(tn TenderOcontract) {
	defer SaveStack()
	db, err := sql.Open("mysql", Dsn)
	if err != nil {
		Logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ?", Prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.dateEnd)
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
	status := strings.TrimSpace(doc.Find("th:contains('Статус') + th").First().Text())
	upDate := time.Now()
	idXml := tn.purNum
	version := 1

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
	orgFullName := strings.TrimSpace(doc.Find("td:contains('Заказчик') + td > a").First().Text())
	if orgFullName == "" {
		orgFullName = strings.TrimSpace(doc.Find("td:contains('Заказчик') + td").First().Text())
	}
	if orgFullName != "" {
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
			email := ""
			phone := ""
			organizerINN := ""
			organizerPostAddress := ""
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, inn = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?", Prefix))
			res, err := stmt.Exec(orgFullName, organizerINN, organizerPostAddress, organizerPostAddress, email, phone)
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
	placinWayName := strings.TrimSpace(doc.Find("th:contains('Вид процедуры') + th").First().Text())
	if placinWayName != "" {
		idPlacingWay = getPlacingWayId(placinWayName, db)
	}
	IdEtp := 0
	etpName := "ЭТП ONLINECONTRACT"
	etpUrl := "http://onlinecontract.ru/tenders"
	IdEtp = getEtpId(etpName, etpUrl, db)
	datePub := upDate
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?, bidding_date = ?", Prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, datePub, tn.url, tn.purObjInfo, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.dateEnd, cancelStatus, upDate, version, tn.url, printForm, 0, status, tn.dateEnd)
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		UpdatetenderOcontract++
	} else {
		AddtenderOcontract++
	}
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
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, reg_num = ?, is223=1, inn = ?", Prefix))
			res, err := stmt.Exec(orgFullName, out, "")
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
	currency := strings.TrimSpace(doc.Find("td:contains('Валюта') + td").First().Text())
	nmckT := strings.TrimSpace(doc.Find("td:contains('Начальная цена контракта') + td").First().Text())
	nmck := strings.TrimSpace(findFromRegExp(nmckT, `^([\d\s,]+)`))
	nmck = strings.Replace(nmck, ",", ".", -1)
	re := regexp.MustCompile(`\s+`)
	nmck = re.ReplaceAllString(nmck, "")
	stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, currency = ?, max_price = ?", Prefix))
	resl, err := stmtl.Exec(idTender, LotNumber, currency, nmck)
	stmtl.Close()
	if err != nil {
		Logging("Ошибка вставки lot", err)
		return
	}
	id, _ := resl.LastInsertId()
	idLot = int(id)
	doc.Find("#positions div.table-responsive table.table-hover tbody tr").Each(func(i int, s *goquery.Selection) {
		t.purObj(idLot, idCustomer, s, db)
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

func (t *ParserOcontract) purObj(idLot int, idCustomer int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	name1 := strings.TrimSpace(doc.Find("td:nth-child(2)").First().Text())
	name2 := strings.TrimSpace(doc.Find("td:nth-child(6)").First().Text())
	name := name1 + "\n" + name2

	quantityT := strings.TrimSpace(doc.Find("td:nth-child(3)").First().Text())
	quantity := strings.TrimSpace(findFromRegExp(quantityT, `^([\d\s,]+)`))
	quantity = strings.Replace(quantity, ",", ".", -1)
	re := regexp.MustCompile(`\s+`)
	quantity = re.ReplaceAllString(quantity, "")

	okei := strings.TrimSpace(findFromRegExp(quantityT, `^[\d\s,]+(.+)`))

	price := strings.TrimSpace(doc.Find("td:nth-child(4)").First().Text())
	price = strings.Replace(price, ",", ".", -1)
	price = re.ReplaceAllString(price, "")

	sum := strings.TrimSpace(doc.Find("td:nth-child(5)").First().Text())
	sum = strings.Replace(sum, ",", ".", -1)
	sum = re.ReplaceAllString(sum, "")
	stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?, quantity_value = ?, okei = ?, customer_quantity_value = ?, price = ?, sum = ?", Prefix))
	_, errr := stmtr.Exec(idLot, idCustomer, name, quantity, okei, quantity, price, sum)
	stmtr.Close()
	if errr != nil {
		Logging("Ошибка вставки purchase_object", errr)
		return
	}
	delivTerm1 := strings.TrimSpace(doc.Find("td:nth-child(7)").First().Text())
	delivTerm2 := strings.TrimSpace(doc.Find("td:nth-child(8)").First().Text())
	delivTerm := fmt.Sprintf("Условия оплаты: %s \nПоставка до: %s", delivTerm1, delivTerm2)
	stmtcr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer_requirement SET id_lot = ?, id_customer = ?, delivery_term = ?", Prefix))
	_, err := stmtcr.Exec(idLot, idCustomer, delivTerm)
	stmtcr.Close()
	if err != nil {
		Logging("Ошибка вставки purchase_object", errr)
		return
	}
}
