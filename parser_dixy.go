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

var addtenderDixy int
var updatetenderDixy int

type parserDixy struct {
	TypeFz int
	Urls   []string
}

func (t *parserDixy) parsing() {
	defer SaveStack()
	logging("Start parsing")
	t.parsingPageAll()
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderDixy))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderDixy))
}

func (t *parserDixy) parsingPageAll() {
	for _, p := range t.Urls {
		t.parsingPage(p)
	}
}

func (t *parserDixy) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPageWithUA(p)
	if r != "" {
		t.parsingTenderList(r)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserDixy) parsingTenderList(p string) {
	defer SaveStack()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		logging(err)
		return
	}
	doc.Find("div.tenderinfo div.invspan4").Each(func(i int, s *goquery.Selection) {
		t.parsingTenderFromList(s)

	})
}

func (t *parserDixy) parsingTenderFromList(p *goquery.Selection) {
	defer SaveStack()
	hrefT := p.Find("span p a")
	href, exist := hrefT.Attr("href")
	if !exist {
		logging("The element cannot have href attribute", hrefT.Text())
		return
	}
	href = fmt.Sprintf("http://www.dixygroup.ru%s", href)
	dateT := strings.TrimSpace(p.Find("p.date").First().Text())
	purName := strings.TrimSpace(p.Find("span p a").First().Text())
	pubDate := getDateDixy(dateT)
	if (pubDate == time.Time{}) {
		logging("cannot parse dates in ", href)
		return
	}
	purNum := findFromRegExp(href, `item(\d+)`)
	if purNum == "" {
		purNum = findFromRegExp(href, `it(\d+)\.aspx`)
	}
	if purNum == "" {
		md := md5.Sum([]byte(href))
		purNum = hex.EncodeToString(md[:])
	}
	if purNum == "" {
		logging("cannot find purnum in ", href)
		return
	}
	t.tender(purNum, href, pubDate, purName)
}

func (t *parserDixy) tender(purNum string, page string, pubDate time.Time, purName string) {
	defer SaveStack()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	upDate := time.Now()
	idXml := purNum
	version := 1
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND doc_publish_date = ? AND type_fz = ?", prefix))
	res, err := stmt.Query(purNum, pubDate, t.TypeFz)
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
	r := DownloadPage(page)
	if r == "" {
		logging("Получили пустую строку", page)
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r))
	if err != nil {
		logging(err)
		return
	}
	var cancelStatus = 0
	var updated = false
	if purNum != "" {
		stmt, err := db.Prepare(fmt.Sprintf("SELECT id_tender, date_version FROM %stender WHERE purchase_number = ? AND cancel=0 AND type_fz = ?", prefix))
		rows, err := stmt.Query(purNum, t.TypeFz)
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
	printForm := page
	idOrganizer := 0
	orgFullName := "«ПАО «ДИКСИ Групп»"
	if true {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_organizer FROM %sorganizer WHERE full_name = ?", prefix))
		rows, err := stmt.Query(orgFullName)
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
			email := "Yu.Dolgushkina@dixy.ru"
			phone := "+7 495 223-33-37"
			organizerINN := "7704249540"
			organizerKPP := "772901001"
			organizerPostAddress := "119361, г. Москва, ул. Б. Очаковская, 47А, стр.1"
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, inn = ?, kpp = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?", prefix))
			res, err := stmt.Exec(orgFullName, organizerINN, organizerKPP, organizerPostAddress, organizerPostAddress, email, phone)
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
	etpName := orgFullName
	etpUrl := "http://www.dixygroup.ru"
	if true {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_etp FROM %setp WHERE name = ? AND url = ? LIMIT 1", prefix))
		rows, err := stmt.Query(etpName, etpUrl)
		stmt.Close()
		if err != nil {
			logging("Ошибка выполения запроса", err)
			return
		}
		if rows.Next() {
			err = rows.Scan(&IdEtp)
			if err != nil {
				logging("Ошибка чтения результата запроса", err)
				return
			}
			rows.Close()
		} else {
			rows.Close()
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %setp SET name = ?, url = ?, conf=0", prefix))
			res, err := stmt.Exec(etpName, etpUrl)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки etp", err)
				return
			}
			id, err := res.LastInsertId()
			IdEtp = int(id)
		}
	}
	verInfo := strings.TrimSpace(doc.Find("#contentleft").First().Text())
	verInfo = cleanString(verInfo)
	endDate := time.Time{}
	endDateT := findFromRegExpDixy(verInfo, `пройдет с \d{2}\.\d{2}\.\d{4} по (\d{2}\.\d{2}\.\d{4}) до (\d{2}:\d{2})`)
	if endDateT != "" {
		endDate = getTimeMoscowLayout(endDateT, "02.01.2006 15:04")
	}
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, notice_version = ?, xml = ?, print_form = ?, id_region = ?", prefix))
	rest, err := stmtt.Exec(idXml, purNum, pubDate, page, purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, endDate, cancelStatus, upDate, version, verInfo, page, printForm, 0)
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderDixy++
	} else {
		addtenderDixy++
	}
	idCustomer := 0
	if orgFullName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name = ?", prefix))
		rows, err := stmt.Query(orgFullName)
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
			res, err := stmt.Exec(orgFullName, "7704249540", out)
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
	stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?", prefix))
	resl, err := stmtl.Exec(idTender, LotNumber)
	stmtl.Close()
	if err != nil {
		logging("Ошибка вставки lot", err)
		return
	}
	id, _ := resl.LastInsertId()
	idLot = int(id)
	stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?", prefix))
	_, errr := stmtr.Exec(idLot, idCustomer, purName)
	stmtr.Close()
	if errr != nil {
		logging("Ошибка вставки purchase_object", errr)
		return
	}
	e := TenderKwords(db, idTender)
	if e != nil {
		logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, purNum, t.TypeFz)
	if e1 != nil {
		logging("Ошибка обработки AddVerNumber", e1)
	}
}
