package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"os/exec"
	"strings"
	"time"
)

var AddtenderNovatek int
var UpdatetenderNovatek int

type ParserNovatek struct {
	TypeFz int
	Urls   []string
}

type TenderNovatek struct {
	url     string
	purName string
	purNum  string
	cusName string
	pubDate time.Time
	endDate time.Time
}

func (t *ParserNovatek) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	t.parsingPageAll()
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderNovatek))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderNovatek))
}

func (t *ParserNovatek) parsingPageAll() {
	for _, p := range t.Urls {
		t.parsingPage(p)
	}
}

func (t *ParserNovatek) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserNovatek) parsingTenderList(p string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("div.tenders-list tbody tr").Each(func(i int, s *goquery.Selection) {
		txt := s.Text()
		if !strings.Contains(txt, "Название конкурса") {
			t.parsingTenderFromList(s)
		}
	})
}
func (t *ParserNovatek) parsingTenderFromList(p *goquery.Selection) {
	defer SaveStack()
	hrefT := p.Find("td:nth-child(1) a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://www.novatek.ru%s", href)
	purName := strings.TrimSpace(p.Find("td:nth-child(1) a").First().Text())
	purNum := findFromRegExp(href, `=(\d+)$`)
	if purName == "" {
		Logging("The element cannot have purNum", href)
		return
	}
	cusName := strings.TrimSpace(p.Find("td:nth-child(2)").First().Text())
	pubDateT := strings.TrimSpace(p.Find("td:nth-child(3)").First().Text())
	if pubDateT == "" {
		Logging("cannot find pubDateT in ", href)
		return
	}
	pubDate := getDateCpc(pubDateT)
	if (pubDate == time.Time{}) {
		Logging("cannot parse pubDate in ", href)
		return
	}
	endDateT := strings.TrimSpace(p.Find("td:nth-child(4)").First().Text())
	if endDateT == "" {
		Logging("cannot find endDateT in ", href)
		return
	}
	endDate := getDateCpc(endDateT)
	if (endDate == time.Time{}) {
		Logging("cannot parse endDate in ", href)
		return
	}
	tnd := TenderNovatek{url: href, purName: purName, purNum: purNum, cusName: cusName, pubDate: pubDate, endDate: endDate}
	t.Tender(tnd)

}

func (t *ParserNovatek) Tender(tn TenderNovatek) {
	defer SaveStack()
	r := DownloadPage(tn.url)
	if r == "" {
		Logging("Получили пустую строку", tn.url)
		return
	}
	if strings.Contains(r, "etp.gpb.ru/#com") {
		Logging("this tender was published on gpb", tn.url)
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r))
	if err != nil {
		Logging(err)
		return
	}
	db, err := sql.Open("mysql", Dsn)
	if err != nil {
		Logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
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
	upDate := time.Now()
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
	if tn.cusName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_organizer FROM %sorganizer WHERE full_name = ?", Prefix))
		rows, err := stmt.Query(tn.cusName)
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
			contactPerson := strings.TrimSpace(doc.Find("div:contains('Контактная информация:') + div").First().Text())
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, inn = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?, contact_person = ?", Prefix))
			res, err := stmt.Exec(tn.cusName, organizerINN, organizerPostAddress, organizerPostAddress, email, phone, contactPerson)
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
	etpName := "ПАО «НОВАТЭК»"
	etpUrl := "http://www.novatek.ru"
	IdEtp = getEtpId(etpName, etpUrl, db)
	idXml := tn.purNum
	version := 1
	idTender := 0
	noticeVersion := strings.TrimSpace(doc.Find("div.center").First().Text())
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?, bidding_date = ?", Prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.pubDate, tn.url, tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, tn.url, printForm, 0, noticeVersion, time.Time{})
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		UpdatetenderNovatek++
	} else {
		AddtenderNovatek++
	}
	dochrefT := doc.Find("div.download span a")
	hrefD, exist := dochrefT.Attr("href")
	if !exist {
		Logging("The element cannot have href attribute", dochrefT.Text())
	}
	hrefD = fmt.Sprintf("http://www.novatek.ru%s", hrefD)
	docName := strings.TrimSpace(doc.Find("div.download span a").First().Text())
	if docName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", Prefix))
		_, err := stmt.Exec(idTender, docName, hrefD)
		stmt.Close()
		if err != nil {
			Logging("Ошибка вставки attachment", err)
			return
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
	idCustomer := 0
	if tn.cusName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name = ?", Prefix))
		rows, err := stmt.Query(tn.cusName)
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
			res, err := stmt.Exec(tn.cusName, out, "")
			stmt.Close()
			if err != nil {
				Logging("Ошибка вставки заказчика", err)
				return
			}
			id, err := res.LastInsertId()
			idCustomer = int(id)
		}
	}
	stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?", Prefix))
	_, errr := stmtr.Exec(idLot, idCustomer, tn.purName)
	stmtr.Close()
	if errr != nil {
		Logging("Ошибка вставки purchase_object", errr)
		return
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
