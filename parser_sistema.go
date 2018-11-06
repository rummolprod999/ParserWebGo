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

var AddtenderSistema int
var UpdatetenderSistema int

type ParserSistema struct {
	TypeFz int
}

type TenderSistema struct {
	purName string
	purNum  string
	url     string
	pubDate time.Time
	endDate time.Time
}

func (t *ParserSistema) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	t.parsingPageAll()
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderSistema))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderSistema))
}

func (t *ParserSistema) parsingPageAll() {
	startUrl := "http://www.sistema.ru/zakupki/zakupki/"
	lst := t.getPageList(startUrl)
	for _, p := range lst {
		t.parsingPage(p)
	}

}
func (t *ParserSistema) getPageList(url string) []string {
	var l = make([]string, 0)
	l = append(l, url)
	r := DownloadPage(url)
	if r == "" {
		panic("empty start page")
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r))
	if err != nil {
		Logging(err)
		panic(err)
	}
	doc.Find("div.browseLinksWrap > a[href ^= 'zakupki/zakupki']").Each(func(i int, s *goquery.Selection) {
		href, exist := s.Attr("href")
		if !exist {
			Logging("The element can not have href attribute", s.Text())
			return
		}
		href = fmt.Sprintf("http://www.sistema.ru/%s", href)
		l = append(l, href)
	})
	l = l[0 : len(l)-2]
	return l

}
func (t *ParserSistema) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserSistema) parsingTenderList(p string, url string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("div.pagingResultsWrapper > div.zakupki").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)
	})
}

func (t *ParserSistema) parsingTenderFromList(p *goquery.Selection, url string) {
	hrefT := p.Find("a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://www.sistema.ru/%s", href)
	purName := strings.TrimSpace(p.Find("a").First().Text())
	if purName == "" {
		Logging("Can not find purName in ", url)
		return
	}
	purNum := findFromRegExp(href, `/(\d+)/$`)
	if purNum == "" {
		Logging("Can not find purnum in ", purName)
		return
	}
	endDate := time.Now()
	pubDateT := strings.TrimSpace(p.Find("h4").First().Text())
	pubDateT = cleanString(pubDateT)
	pubDate := getDateDixy(pubDateT)
	if (pubDate == time.Time{}) {
		Logging("Can not find pubDate in ", href, purNum)
		return
	}
	tnd := TenderSistema{purName: purName, purNum: purNum, pubDate: pubDate, endDate: endDate, url: href}
	t.Tender(tnd)
}
func (t *ParserSistema) Tender(tn TenderSistema) {
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
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND doc_publish_date = ?", Prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.pubDate)
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
	orgName := `АФК «Система»`
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
			contactPhone := "+7 (495) 737-01-01"
			contactEmail := "ipetrov@sistema.ru"
			inn := "7703104630"
			postAddress := "125009 Москва, ул. Моховая, 13 (м. Охотный ряд, выход к ул. Моховая, 100 метров)"
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
	idPlacingWay := 0
	IdEtp := 0
	etpName := orgName
	etpUrl := "http://www.sistema.ru/"
	IdEtp = getEtpId(etpName, etpUrl, db)
	endDateTmp := strings.TrimSpace(doc.Find("p:contains('Срок окончания приема предложений')").First().Text())
	if endDateTmp == "" {
		endDateTmp = strings.TrimSpace(doc.Find("p:contains('Дата окончания приема предложений')").First().Text())
	}
	if endDateTmp == "" {
		endDateTmp = strings.TrimSpace(doc.Find("p:contains('Срок окончания приема документов')").First().Text())
	}
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
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?", Prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.pubDate, tn.url, tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, tn.url, printForm, 0, "")
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		UpdatetenderSistema++
	} else {
		AddtenderSistema++
	}
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
			innCus := "7703104630"
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, reg_num = ?, is223=1, inn = ?", Prefix))
			res, err := stmt.Exec(orgName, out, innCus)
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
	stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, currency = ?", Prefix))
	resl, err := stmtl.Exec(idTender, LotNumber, "")
	stmtl.Close()
	if err != nil {
		Logging("Ошибка вставки lot", err)
		return
	}
	id, _ := resl.LastInsertId()
	idLot = int(id)
	stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?", Prefix))
	_, errr := stmtr.Exec(idLot, idCustomer, tn.purName)
	stmtr.Close()
	if errr != nil {
		Logging("Ошибка вставки purchase_object", errr)
		return
	}
	doc.Find("p.bodytext > a[href ^= 'fileadmin/user_upload']").Each(func(i int, s *goquery.Selection) {
		t.documents(idTender, s, db)
	})
}
func (t *ParserSistema) documents(idTender int, doc *goquery.Selection, db *sql.DB) {
	defer SaveStack()
	nameF := strings.TrimSpace(doc.First().Text())
	href, exist := doc.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", doc.Text())
		return
	}
	href = fmt.Sprintf("http://www.sistema.ru/%s", href)
	if nameF != "" {
		nameF = fmt.Sprintf("Закупочная документация %s", nameF)
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", Prefix))
		_, err := stmt.Exec(idTender, nameF, href)
		stmt.Close()
		if err != nil {
			Logging("Ошибка вставки attachment", err)
			return
		}
	}
}
