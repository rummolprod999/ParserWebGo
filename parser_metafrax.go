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

var AddtenderMetafrax int
var UpdatetenderMetafrax int

type ParserMetafrax struct {
	TypeFz int
	Urls   []string
}

type TenderMetafrax struct {
	purName string
	purNum  string
	url     string
	pubDate time.Time
	endDate time.Time
	docs    []map[string]string
}

func (t *ParserMetafrax) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	t.parsingPageAll()
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderMetafrax))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderMetafrax))
}

func (t *ParserMetafrax) parsingPageAll() {
	for _, p := range t.Urls {
		t.parsingPage(p)
	}
}
func (t *ParserMetafrax) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserMetafrax) parsingTenderList(p string, url string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("div.tender_itm").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Срок подачи заявки") {
			t.parsingTenderFromList(s, url)
		}

	})
}

func (t *ParserMetafrax) parsingTenderFromList(p *goquery.Selection, url string) {
	purName := strings.TrimSpace(p.Find("p > strong").First().Text())
	if purName == "" {
		Logging("cannot find purName in ", url)
		return
	}
	hrefT := p.Find("a.tenderDoc")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://metafrax.ru%s", href)
	dateTmp := strings.TrimSpace(p.Find("p:nth-child(2) > span:nth-child(2)").First().Text())
	endDateTT := findFromRegExp(dateTmp, `-\s*(\d{2}\.\d{2}\.\d{4})`)
	endDate := getTimeMoscowLayout(endDateTT, "02.01.2006")
	pubDateTT := findFromRegExp(dateTmp, `(\d{2}\.\d{2}\.\d{4})\s*-`)
	pubDate := getTimeMoscowLayout(pubDateTT, "02.01.2006")
	purNum := GetMd5(purName)
	if (endDate == time.Time{} || pubDate == time.Time{}) {
		Logging("cannot find endDate or pubDate in ", href, purNum)
		return
	}
	docList := make([]map[string]string, 0)
	p.Find("a.tenderDoc").Each(func(i int, s *goquery.Selection) {
		hrefDoc, existD := s.Attr("href")
		if existD {
			hrefDoc = fmt.Sprintf("http://metafrax.ru%s", hrefDoc)
			nameDoc := strings.TrimSpace(s.First().Text())
			if nameDoc != "" {
				dc := make(map[string]string)
				dc[hrefDoc] = nameDoc
				docList = append(docList, dc)
			}
		}
	})
	tnd := TenderMetafrax{purName: purName, purNum: purNum, url: href, pubDate: pubDate, endDate: endDate, docs: docList}
	t.Tender(tnd)

}
func (t *ParserMetafrax) Tender(tn TenderMetafrax) {
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
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ? AND doc_publish_date = ?", Prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.endDate, tn.pubDate)
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
	orgName := "ПАО «МЕТАФРАКС»"
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
			contactPhone := "(34248) 4-08-98"
			contactEmail := "metafrax@permonline.ru"
			inn := "5913001268"
			postAddress := "Россия, 618250, г. Губаха"
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
	etpUrl := "http://metafrax.ru"
	IdEtp = getEtpId(etpName, etpUrl, db)
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
		UpdatetenderMetafrax++
	} else {
		AddtenderMetafrax++
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
			innCus := "5913001268"
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
	for _, dd := range tn.docs {
		for k, v := range dd {
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", Prefix))
			_, err := stmt.Exec(idTender, v, k)
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки attachment", err)
				return
			}
		}
	}
	e := TenderKwords(db, idTender)
	if e != nil {
		Logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, tn.purNum, t.TypeFz)
	if e1 != nil {
		Logging("Ошибка обработки AddVerNumber", e1)
	}
}
