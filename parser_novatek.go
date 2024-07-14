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

var addtenderNovatek int
var updatetenderNovatek int

type parserNovatek struct {
	TypeFz int
	Urls   []string
}

type tenderNovatek struct {
	url     string
	purName string
	purNum  string
	cusName string
	pubDate time.Time
	endDate time.Time
}

func (t *parserNovatek) parsing() {
	defer SaveStack()
	logging("Start parsing")
	t.parsingPageAll()
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderNovatek))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderNovatek))
}

func (t *parserNovatek) parsingPageAll() {
	for _, p := range t.Urls {
		t.parsingPage(p)
	}
}

func (t *parserNovatek) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPageInsecure(p)
	if r != "" {
		t.parsingTenderList(r)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserNovatek) parsingTenderList(p string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(p))
	if err != nil {
		logging(err)
		return
	}
	count := 0
	doc.Find("div.tlike div.tr").Each(func(i int, s *goquery.Selection) {
		if count > 0 {
			t.parsingTenderFromList(s)
		}
		count++
	})
}
func (t *parserNovatek) parsingTenderFromList(p *goquery.Selection) {
	defer SaveStack()
	hrefT := p.Find("div:nth-child(1) a")
	href, exist := hrefT.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://www.novatek.ru%s", href)
	purName := strings.TrimSpace(p.Find("div:nth-child(1) a").First().Text())
	purName = strings.TrimSpace(strings.Replace(purName, "Тендер", "", -1))
	purNum := findFromRegExp(href, `=(\d+)$`)
	if purName == "" {
		logging("The element cannot have purNum", href)
		return
	}
	cusName := strings.TrimSpace(p.Find("div:nth-child(2)").First().Text())
	cusName = strings.TrimSpace(strings.Replace(cusName, "Заказчик", "", -1))
	pubDateT := strings.TrimSpace(p.Find("div:nth-child(3)").First().Text())
	pubDateT = strings.TrimSpace(strings.Replace(pubDateT, "Начало сбора оферт", "", -1))
	if pubDateT == "" {
		logging("cannot find pubDateT in ", href)
		return
	}
	pubDate := getDateCpc(pubDateT)
	if (pubDate == time.Time{}) {
		logging("cannot parse pubDate in ", href)
		return
	}
	endDateT := strings.TrimSpace(p.Find("div:nth-child(4)").First().Text())
	endDateT = strings.TrimSpace(strings.Replace(endDateT, "Окончание сбора оферт", "", -1))
	if endDateT == "" {
		logging("cannot find endDateT in ", href)
		return
	}
	endDate := getDateCpc(endDateT)
	if (endDate == time.Time{}) {
		endDate = pubDate.AddDate(0, 0, 2)
	}
	tnd := tenderNovatek{url: href, purName: purName, purNum: purNum, cusName: cusName, pubDate: pubDate, endDate: endDate}
	t.tender(tnd)

}

func (t *parserNovatek) tender(tn tenderNovatek) {
	defer SaveStack()
	r := DownloadPageInsecure(tn.url)
	if r == "" {
		logging("Получили пустую строку", tn.url)
		return
	}
	/*if strings.Contains(r, "etp.gpb.ru/#com") {
		Logging("this tender was published on gpb", tn.url)
		return
	}*/
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r))
	if err != nil {
		logging(err)
		return
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ? AND doc_publish_date = ?", prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.endDate, tn.pubDate)
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
	upDate := time.Now()
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
	if tn.cusName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_organizer FROM %sorganizer WHERE full_name = ?", prefix))
		rows, err := stmt.Query(tn.cusName)
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
			email := ""
			phone := ""
			organizerINN := ""
			organizerPostAddress := ""
			contactPerson := strings.TrimSpace(doc.Find("div:contains('Контактная информация:') + div").First().Text())
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, inn = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?, contact_person = ?", prefix))
			res, err := stmt.Exec(tn.cusName, organizerINN, organizerPostAddress, organizerPostAddress, email, phone, contactPerson)
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
	etpName := "ПАО «НОВАТЭК»"
	etpUrl := "http://www.novatek.ru"
	IdEtp = getEtpId(etpName, etpUrl, db)
	idXml := tn.purNum
	version := 1
	idTender := 0
	noticeVersion := strings.TrimSpace(doc.Find("div.center").First().Text())
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?, bidding_date = ?", prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.pubDate, tn.url, tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, tn.url, printForm, 0, noticeVersion, time.Time{})
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderNovatek++
	} else {
		addtenderNovatek++
	}
	dochrefT := doc.Find("div.download a")
	hrefD, exist := dochrefT.Attr("href")
	if exist {
		hrefD = fmt.Sprintf("http://www.novatek.ru%s", hrefD)
		docName := strings.TrimSpace(dochrefT.Text())
		if docName != "" {
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", prefix))
			_, err := stmt.Exec(idTender, docName, hrefD)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки attachment", err)
				return
			}
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
	idCustomer := 0
	if tn.cusName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name = ?", prefix))
		rows, err := stmt.Query(tn.cusName)
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
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, reg_num = ?, is223=1, inn = ?", prefix))
			res, err := stmt.Exec(tn.cusName, out, "")
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки заказчика", err)
				return
			}
			id, err := res.LastInsertId()
			idCustomer = int(id)
		}
	}
	stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?", prefix))
	_, errr := stmtr.Exec(idLot, idCustomer, tn.purName)
	stmtr.Close()
	if errr != nil {
		logging("Ошибка вставки purchase_object", errr)
		return
	}
	e := TenderKwords(db, idTender)
	if e != nil {
		logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, tn.purNum, t.TypeFz)
	if e1 != nil {
		logging("Ошибка обработки AddVerNumber", e1)
	}
}
