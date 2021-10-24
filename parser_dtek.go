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

var addtenderDtek int
var updatetenderDtek int

type parserDtek struct {
	TypeFz int
}

type tenderDtek struct {
	purName string
	purNum  string
	url     string
	pubDate time.Time
	endDate time.Time
	cusName string
	pwName  string
}

func (t *parserDtek) parsing() {
	defer SaveStack()
	logging("Start parsing")
	t.parsingPage("https://tenders.dtek.com/rus/purchase/?")
	for i := 19; i <= 325; i += 18 {
		url := fmt.Sprintf("%s%d", "https://tenders.dtek.com/rus/purchase/?&offers_next=", i)
		t.parsingPage(url)
	}
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderDtek))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderDtek))
}

func (t *parserDtek) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage1251(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserDtek) parsingTenderList(p string, url string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		logging(err)
		return
	}
	doc.Find("table[cellspacing='0'] tbody tr").Each(func(i int, s *goquery.Selection) {
		txt := s.Text()
		if !strings.Contains(txt, "Короткое описание") {
			t.parsingTenderFromList(s, url)
		}

	})
}

func (t *parserDtek) parsingTenderFromList(p *goquery.Selection, url string) {
	hrefT := p.Find("a")
	href, exist := hrefT.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	numTmp := strings.TrimSpace(p.Find("td:nth-child(1)").First().Text())
	href = fmt.Sprintf("https://tenders.dtek.com/rus/purchase/%s", strings.TrimSpace(href))
	purName := strings.TrimSpace(p.Find("td:nth-child(3)").First().Text())
	if purName == "" {
		purName = cleanString(numTmp)
	}
	if purName == "" {
		logging("cannot find purName in ", url)
		return
	}
	pwName := ""
	cusName := strings.TrimSpace(p.Find("td:nth-child(2) strong").First().Text())
	purNum := findFromRegExp(numTmp, `^(\d+)`)
	if purNum == "" {
		purNum = cleanString(numTmp)
	} else {
		pwName = strings.TrimSpace(strings.Replace(numTmp, purNum, "", -1))
	}
	if purNum == "" {
		logging("cannot find purnum in ", purName)
		return
	}

	pubDateT := strings.TrimSpace(p.Find("td:nth-child(4)").First().Text())
	pubDate := getTimeMoscowLayout(pubDateT, "02.01.2006 15:04")
	endDateT := strings.TrimSpace(p.Find("td:nth-child(5)").First().Text())
	endDateT = strings.Replace(endDateT, "(продлено)", "", -1)
	endDateT = strings.Trim(endDateT, "\n ")
	endDate := getTimeMoscowLayout(endDateT, "02.01.2006 15:04")
	if (pubDate == time.Time{} || endDate == time.Time{}) {
		logging("cannot find enddate or startdate in ", href, purNum)
		return
	}
	tnd := tenderDtek{purName: purName, purNum: purNum, url: href, pubDate: pubDate, endDate: endDate, cusName: cusName, pwName: pwName}
	t.tender(tnd)

}

func (t *parserDtek) tender(tn tenderDtek) {
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
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND doc_publish_date = ? AND type_fz = ? AND end_date = ?", prefix))
	res, err := stmt.Query(tn.purNum, tn.pubDate, t.TypeFz, tn.endDate)
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
	r := DownloadPage1251(tn.url)
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
	orgName := tn.cusName
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
			contactPerson := strings.TrimSpace(doc.Find("td:contains('Ответственное лицо') + td").First().Text())
			contactPhone := ""
			contactEmail := ""
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, contact_person = ?, contact_phone = ?, contact_email = ?", prefix))
			res, err := stmt.Exec(orgName, contactPerson, contactPhone, contactEmail)
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
	if tn.pwName != "" {
		idPlacingWay = getPlacingWayId(tn.pwName, db)
	}
	IdEtp := 0
	etpName := "ЭТП ДТЭК"
	etpUrl := "https://tenders.dtek.com"
	IdEtp = getEtpId(etpName, etpUrl, db)
	idTender := 0
	version_notice := strings.TrimSpace(doc.Find("td:contains('Описание') + td.content").First().Text())
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?", prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.pubDate, tn.url, tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, tn.url, printForm, 0, version_notice)
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderDtek++
	} else {
		addtenderDtek++
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
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, reg_num = ?, is223=1", prefix))
			res, err := stmt.Exec(orgName, out)
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
	e := TenderKwords(db, idTender)
	if e != nil {
		logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, tn.purNum, t.TypeFz)
	if e1 != nil {
		logging("Ошибка обработки AddVerNumber", e1)
	}
}
