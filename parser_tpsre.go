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

var addtenderTpsre int
var updatetenderTpsre int

type parserTpsre struct {
	TypeFz int
	Urls   []string
}

type tenderTpsre struct {
	purName string
	purNum  string
	url     string
}

func (t *parserTpsre) parsing() {
	defer SaveStack()
	logging("Start parsing")
	t.parsingPageAll()
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderTpsre))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderTpsre))
}

func (t *parserTpsre) parsingPageAll() {
	for _, p := range t.Urls {
		t.parsingPage(p)
	}
}
func (t *parserTpsre) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserTpsre) parsingTenderList(p string, url string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		logging(err)
		return
	}
	doc.Find("tr[data-url] td a").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s, url)
	})
}

func (t *parserTpsre) parsingTenderFromList(p *goquery.Selection, url string) {
	href, exist := p.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", url)
		return
	}
	href = fmt.Sprintf("https://www.tpsre.ru%s", href)
	purName := strings.TrimSpace(p.Text())
	if purName == "" {
		logging("Cannot find purName in ", url)
		return
	}
	purNum := GetMd5(purName)
	tnd := tenderTpsre{purName: purName, purNum: purNum, url: href}
	t.tender(tnd)
}
func (t *parserTpsre) tender(tn tenderTpsre) {
	defer SaveStack()
	r := DownloadPage(tn.url)
	if r == "" {
		logging("Получили пустую строку", tn.url)
		return
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging("Ошибка подключения к БД", err)
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r))
	if err != nil {
		logging(err)
		return
	}
	dateStart := strings.TrimSpace(doc.Find("p:contains('Сроки проведения') + p").First().Text())
	dateStart = findFromRegExp(dateStart, `^(\d{2}\.\d{2}\.\d{4})`)
	pubDate := getTimeMoscowLayout(dateStart, "02.01.2006")
	if (pubDate == time.Time{}) {
		logging("cannot find pubDate in ", tn.url)
		return
	}
	EndDateT := strings.TrimSpace(doc.Find("p:contains('Сроки проведения') + p").First().Text())
	EndDateT = findFromRegExp(EndDateT, `(\d{2}\.\d{2}\.\d{4})$`)
	endDate := getTimeMoscowLayout(EndDateT, "02.01.2006")
	if (endDate == time.Time{}) {
		endDate = pubDate.AddDate(0, 0, 2)
	}
	tenderSubj := strings.TrimSpace(doc.Find("p:contains('Объект тендера') + p > a").First().Text())
	if tenderSubj != "" {
		tn.purName = fmt.Sprintf("%s %s", tn.purName, tenderSubj)
	}
	tn.purNum = strings.TrimSpace(findFromRegExp(tn.url, `/(\d+)/$`))
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	upDate := time.Now()
	idXml := tn.purNum
	version := 1
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ?", prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, endDate)
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
	orgName := "АО «ТПС Недвижимость»"
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
	etpName := orgName
	etpUrl := "https://www.tpsre.ru/"
	IdEtp := 0
	IdEtp = getEtpId(etpName, etpUrl, db)
	idPlacingWay := 0
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?", prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, pubDate, tn.url, tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, endDate, cancelStatus, upDate, version, tn.url, printForm, 0)
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderTpsre++
	} else {
		addtenderTpsre++
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
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, inn = ?, reg_num = ?, is223=1", prefix))
			res, err := stmt.Exec(orgName, "", out)
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
