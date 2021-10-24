package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"os/exec"
	"strings"
	"time"
)

var addtenderSistema int
var updatetenderSistema int

type parserSistema struct {
	TypeFz int
}

type tenderSistema struct {
	purName string
	purNum  string
	url     string
	pubDate time.Time
	endDate time.Time
}

func (t *parserSistema) parsing() {
	defer SaveStack()
	logging("Start parsing")
	t.parsingPageAll()
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderSistema))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderSistema))
}

func (t *parserSistema) parsingPageAll() {
	startUrl := "https://sistema.ru/procurements"
	t.parsingPage(startUrl)
}
func (t *parserSistema) getPageList(url string) []string {
	var l = make([]string, 0)
	l = append(l, url)
	r := DownloadPage(url)
	if r == "" {
		panic("empty start page")
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r))
	if err != nil {
		logging(err)
		panic(err)
	}
	doc.Find("div[class ^='Cells__Item']").Each(func(i int, s *goquery.Selection) {
		href, exist := s.Attr("href")
		if !exist {
			logging("The element cannot have href attribute", s.Text())
			return
		}
		href = fmt.Sprintf("http://www.sistema.ru/%s", href)
		l = append(l, href)
	})
	l = l[0 : len(l)-2]
	return l

}
func (t *parserSistema) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserSistema) parsingTenderList(p string, url string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		logging(err)
		return
	}
	doc.Find("div[class ^='Cells__Item']").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)
	})
}

func (t *parserSistema) parsingTenderFromList(p *goquery.Selection, url string) {
	hrefT := p.Find("a[href ^= '/procurements/']")
	href, exist := hrefT.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://www.sistema.ru%s", href)
	purName := strings.TrimSpace(p.Find("div[class ^= 'Cells__NameWrapper'] div[class ^= 'Cells__Title_'] + div").First().Text())
	if purName == "" {
		logging("cannot find purName in ", url)
		return
	}
	md := md5.Sum([]byte(purName))
	purNum := hex.EncodeToString(md[:])
	if purNum == "" {
		logging("cannot find purName in ", purName)
		return
	}
	dates := strings.TrimSpace(p.Find("div[class ^= 'Cells__Date']").First().Text())
	dates = strings.Replace(dates, "Начало и окончание приёма заявок", "", -1)
	datePubT, dateEndT := findFromRegExpTwoValue(dates, `(\d{2}.+\d{4}).+(\d{2}.+\d{4})`)
	pubDate := getDateCpc(datePubT)
	endDate := getDateCpc(dateEndT)
	if (pubDate == time.Time{}) {
		logging("cannot find pubDate in ", href, purNum)
		return
	}
	tnd := tenderSistema{purName: purName, purNum: purNum, pubDate: pubDate, endDate: endDate, url: href}
	t.tender(tnd)
}
func (t *parserSistema) tender(tn tenderSistema) {
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
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND doc_publish_date = ?", prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.pubDate)
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
	orgName := `АФК «Система»`
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
			contactPhone := "+7 (495) 737-01-01"
			contactEmail := "ipetrov@sistema.ru"
			inn := "7703104630"
			postAddress := "125009 Москва, ул. Моховая, 13 (м. Охотный ряд, выход к ул. Моховая, 100 метров)"
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
	idPlacingWay := 0
	IdEtp := 0
	etpName := orgName
	etpUrl := "http://www.sistema.ru/"
	IdEtp = getEtpId(etpName, etpUrl, db)
	tempEndDate := tn.endDate
	endDateTmp := strings.TrimSpace(doc.Find("p b:contains('Срок окончания приема предложений')").First().Text())
	if endDateTmp == "" {
		endDateTmp = strings.TrimSpace(doc.Find("p b:contains('Дата окончания приема предложений')").First().Text())
	}
	if endDateTmp == "" {
		endDateTmp = strings.TrimSpace(doc.Find("p b:contains('Срок окончания приема документов')").First().Text())
	}
	if endDateTmp != "" {
		endDateTmp = cleanString(endDateTmp)
		tm, dt := findTwoFromRegExp(endDateTmp, `до\s?(\d{2}\.\d{2}).*(\d{2}\.\d{2}\.\d{4})`)
		endDateT := fmt.Sprintf("%s %s", dt, strings.Replace(tm, ".", ":", -1))
		tn.endDate = getTimeMoscowLayout(endDateT, "02.01.2006 15:04")
		if (tn.endDate == time.Time{}) {
			endDateTmp = findFromRegExp(endDateTmp, `(["«]\d{2}["»].+\d{4})`)
			endDateTmp = strings.Replace(endDateTmp, "\"", "", -1)
			endDateTmp = strings.Replace(endDateTmp, "«", "", -1)
			endDateTmp = strings.Replace(endDateTmp, "»", "", -1)
			tn.endDate = getDateCpc(endDateTmp)
		}
	}
	if (tn.endDate == time.Time{}) {
		tn.endDate = tempEndDate
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
		updatetenderSistema++
	} else {
		addtenderSistema++
	}
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
			innCus := "7703104630"
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, reg_num = ?, is223=1, inn = ?", prefix))
			res, err := stmt.Exec(orgName, out, innCus)
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
	stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, currency = ?", prefix))
	resl, err := stmtl.Exec(idTender, LotNumber, "")
	stmtl.Close()
	if err != nil {
		logging("Ошибка вставки lot", err)
		return
	}
	id, _ := resl.LastInsertId()
	idLot = int(id)
	stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?", prefix))
	_, errr := stmtr.Exec(idLot, idCustomer, tn.purName)
	stmtr.Close()
	if errr != nil {
		logging("Ошибка вставки purchase_object", errr)
		return
	}
	doc.Find("p > a[href ^= '/upload/']").Each(func(i int, s *goquery.Selection) {
		t.documents(idTender, s, db)
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
func (t *parserSistema) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.First().Text())
	href, exist := doc.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("http://www.sistema.ru%s", href)
	if nameF != "" {
		nameF = fmt.Sprintf("Закупочная документация %s", nameF)
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", prefix))
		_, err := stmt.Exec(idTender, nameF, href)
		stmt.Close()
		if err != nil {
			logging("Ошибка вставки attachment", err)
			return
		}
	}
}
