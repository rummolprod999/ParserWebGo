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

var AddtenderDtek int
var UpdatetenderDtek int

type ParserDtek struct {
	TypeFz int
}

type TenderDtek struct {
	purName string
	purNum  string
	url     string
	pubDate time.Time
	endDate time.Time
	cusName string
	pwName  string
}

func (t *ParserDtek) parsing() {
	defer SaveStack()
	Logging("Start parsing")
	t.parsingPage("https://tenders.dtek.com/rus/purchase/?")
	for i := 19; i <= 325; i += 18 {
		url := fmt.Sprintf("%s%d", "https://tenders.dtek.com/rus/purchase/?&offers_next=", i)
		t.parsingPage(url)
	}
	Logging("End parsing")
	Logging(fmt.Sprintf("Добавили тендеров %d", AddtenderDtek))
	Logging(fmt.Sprintf("Обновили тендеров %d", UpdatetenderDtek))
}

func (t *ParserDtek) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage1251(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		Logging("Получили пустую строку", p)
	}
}

func (t *ParserDtek) parsingTenderList(p string, url string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		Logging(err)
		return
	}
	doc.Find("table[cellspacing='0'] tbody tr").Each(func(i int, s *goquery.Selection) {
		txt := s.Text()
		if !strings.Contains(txt, "Короткое описание") {
			t.parsingTenderFromList(s, url)
		}

	})
}

func (t *ParserDtek) parsingTenderFromList(p *goquery.Selection, url string) {
	hrefT := p.Find("a")
	href, exist := hrefT.Attr("href")
	if !exist {
		Logging("The element can not have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("https://tenders.dtek.com/rus/purchase/%s", strings.TrimSpace(href))
	purName := strings.TrimSpace(p.Find("td:nth-child(3)").First().Text())
	if purName == "" {
		Logging("Can not find purName in ", url)
		return
	}
	pwName := ""
	cusName := strings.TrimSpace(p.Find("td:nth-child(2) strong").First().Text())
	numTmp := strings.TrimSpace(p.Find("td:nth-child(1)").First().Text())
	purNum := findFromRegExp(numTmp, `^(\d+)`)
	if purNum == "" {
		purNum = cleanString(numTmp)
	} else {
		pwName = strings.TrimSpace(strings.Replace(numTmp, purNum, "", -1))
	}
	if purNum == "" {
		Logging("Can not find purnum in ", purName)
		return
	}

	pubDateT := strings.TrimSpace(p.Find("td:nth-child(4)").First().Text())
	pubDate := getTimeMoscowLayout(pubDateT, "02.01.2006 15:04")
	endDateT := strings.TrimSpace(p.Find("td:nth-child(5)").First().Text())
	endDate := getTimeMoscowLayout(endDateT, "02.01.2006 15:04")
	if (pubDate == time.Time{} || endDate == time.Time{}) {
		Logging("Can not find enddate or startdate in ", href, purNum)
		return
	}
	tnd := TenderDtek{purName: purName, purNum: purNum, url: href, pubDate: pubDate, endDate: endDate, cusName: cusName, pwName: pwName}
	t.Tender(tnd)

}

func (t *ParserDtek) Tender(tn TenderDtek) {
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
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND doc_publish_date = ? AND type_fz = ? AND end_date = ?", Prefix))
	res, err := stmt.Query(tn.purNum, tn.pubDate, t.TypeFz, tn.endDate)
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
	r := DownloadPage1251(tn.url)
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
	orgName := tn.cusName
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
			contactPerson := strings.TrimSpace(doc.Find("td:contains('Ответственное лицо') + td").First().Text())
			contactPhone := ""
			contactEmail := ""
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, contact_person = ?, contact_phone = ?, contact_email = ?", Prefix))
			res, err := stmt.Exec(orgName, contactPerson, contactPhone, contactEmail)
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
	if tn.pwName != "" {
		idPlacingWay = getPlacingWayId(tn.pwName, db)
	}
	IdEtp := 0
	etpName := "ЭТП ДТЭК"
	etpUrl := "https://tenders.dtek.com"
	IdEtp = getEtpId(etpName, etpUrl, db)
	idTender := 0
	version_notice := strings.TrimSpace(doc.Find("td:contains('Описание') + td.content").First().Text())
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?", Prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.pubDate, tn.url, tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, tn.url, printForm, 0, version_notice)
	stmtt.Close()
	if err != nil {
		Logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		UpdatetenderDtek++
	} else {
		AddtenderDtek++
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
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, reg_num = ?, is223=1", Prefix))
			res, err := stmt.Exec(orgName, out)
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
	e := TenderKwords(db, idTender)
	if e != nil {
		Logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, tn.purNum, t.TypeFz)
	if e1 != nil {
		Logging("Ошибка обработки AddVerNumber", e1)
	}
}
