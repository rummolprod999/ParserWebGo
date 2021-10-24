package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/buger/jsonparser"
	_ "github.com/go-sql-driver/mysql"
	strip "github.com/grokify/html-strip-tags-go"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

type parserMmk2 struct {
	TypeFz int
	Urls   []string
}

type tenderMmk2 struct {
	purName       string
	purNum        string
	url           string
	pubDate       time.Time
	endDate       time.Time
	status        string
	contactPerson string
	email         string
	cusName       string
}

func (t *parserMmk2) parsing() {
	defer SaveStack()
	t.parsingPageAll()
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderMmk))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderMmk))
}

func (t *parserMmk2) parsingPageAll() {
	for _, p := range t.Urls {
		t.parsingPage(p)
	}
}

func (t *parserMmk2) parsingPage(p string) {
	defer SaveStack()
	r := DownloadPage(p)
	if r != "" {
		t.parsingTenderList(r, p)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserMmk2) parsingTenderList(p string, url string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p))
	if err != nil {
		logging(err)
		return
	}
	doc.Find("a[href = '/for_suppliers/auction/index.php']").Each(func(i int, s *goquery.Selection) {
		t.parsingCategoryFromList(s, url)
	})
}

func (t *parserMmk2) parsingCategoryFromList(p *goquery.Selection, _ string) {
	onclick, exist := p.Attr("onclick")
	cusName := strings.TrimSpace(p.Text())
	if !exist {
		return
	}
	val1, val2 := findFromRegExpTwoValue(onclick, `'(.+)','(.+)'`)
	val1 = strings.Replace(val1, "\\", "", -1)
	val2 = strings.Replace(val2, "\\", "", -1)
	urlP := url.QueryEscape(fmt.Sprintf("\\'%s\\'", val1))
	urlT := fmt.Sprintf("http://mmk.ru/for_suppliers/auction/source.php?LOCATION_CODE=%s", urlP)
	r := DownloadPage(urlT)
	if r != "" {
		t.parsingTenderFromJsonList(r, cusName)
	} else {
		logging("Получили пустую строку", p)
	}
}

func (t *parserMmk2) parsingTenderFromJsonList(page, cusName string) {
	defer SaveStack()
	_, err := jsonparser.ArrayEach([]byte(page), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			logging(err, "callback workWithResponse")
			return
		}
		purNum, err := jsonparser.GetString(value, "[0]")
		if err != nil {
			logging(err)
			return
		}
		hrefT, err := jsonparser.GetString(value, "[1]")
		if err != nil {
			logging(err)
			return
		}
		href, purName := findTwoFromRegExp(hrefT, `href=(.+?)>(.+)<`)
		href = fmt.Sprintf("http://mmk.ru%s", href)
		pubDateT, err := jsonparser.GetString(value, "[2]")
		if err != nil {
			logging(err)
			return
		}
		pubDateT = strings.Replace(pubDateT, "(MSK)", "", -1)
		pubDate := getTimeMoscowLayout(pubDateT, "02.01.2006 15:04:05")

		endDateT, err := jsonparser.GetString(value, "[3]")
		if err != nil {
			logging(err)
			return
		}
		endDateT = strings.Replace(endDateT, "(MSK)", "", -1)
		endDate := getTimeMoscowLayout(endDateT, "02.01.2006 15:04:05")

		person, err := jsonparser.GetString(value, "[4]")
		if err != nil {
			logging(err)
			return
		}
		person = strip.StripTags(person)
		person = strings.Replace(person, "E-mail: ", " E-mail: ", 1)
		email := findFromRegExp(person, `E-mail: (.+)`)
		status, err := jsonparser.GetString(value, "[5]")
		if err != nil {
			logging(err)
			return
		}
		tnd := tenderMmk2{purName: purName, purNum: purNum, url: href, pubDate: pubDate, endDate: endDate, status: status, contactPerson: person, email: email, cusName: cusName}
		t.tender(tnd)
	}, "aaData")
	if err != nil {
		logging(err, "AAAAAAAAAAAAAAAAAAAAAA!!!!!!!!!!!!!!")
		return
	}
}

func (t *parserMmk2) tender(tn tenderMmk2) {
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
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? AND end_date = ? AND notice_version = ?", prefix))
	res, err := stmt.Query(tn.purNum, t.TypeFz, tn.endDate, tn.status)
	stmt.Close()
	if err != nil {
		logging("Ошибка выполения запроса", err)
		return
	}
	if res.Next() {
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
	orgName := `ПАО "Магнитогорский металлургический комбинат"`
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
			contactPhone := "+7 (3519) 24-40-09"
			contactEmail := "infommk@mmk.ru"
			inn := "7414003633"
			postAddress := "Россия, Челябинская область, 455000, г. Магнитогорск, ул.Кирова, 93"
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, contact_person = ?, contact_phone = ?, contact_email = ?, inn = ?, post_address = ?", prefix))
			res, err := stmt.Exec(orgName, tn.contactPerson, contactPhone, contactEmail, inn, postAddress)
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
	etpUrl := "http://mmk.ru/"
	IdEtp = getEtpId(etpName, etpUrl, db)
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, cancel = ?, date_version = ?, num_version = ?, xml = ?, print_form = ?, id_region = ?, notice_version = ?", prefix))
	rest, err := stmtt.Exec(idXml, tn.purNum, tn.pubDate, tn.url, tn.purName, t.TypeFz, idOrganizer, idPlacingWay, IdEtp, tn.endDate, cancelStatus, upDate, version, tn.url, printForm, 0, tn.status)
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderMmk++
	} else {
		addtenderMmk++
	}
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
			innCus := ""
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, reg_num = ?, is223=1, inn = ?", prefix))
			res, err := stmt.Exec(tn.cusName, out, innCus)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки заказчика", err)
				return
			}
			id, err := res.LastInsertId()
			idCustomer = int(id)
		}
	}
	urlT := fmt.Sprintf("http://mmk.ru/for_suppliers/auction/source_dt_l.php?id=%s&did=%s", tn.purNum, tn.purNum)
	r := DownloadPage(urlT)
	if r != "" {
		t.lots(tn, r, idCustomer, idTender, db)
	} else {
		logging("Получили пустую строку", tn.url)
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

func (t *parserMmk2) lots(tn tenderMmk2, lots string, idCustomer, idTender int, db *sql.DB) {
	LotNumber := 1
	_, err := jsonparser.ArrayEach([]byte(lots), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			logging(err, "callback workWithResponse")
			return
		}
		lotName, err := jsonparser.GetString(value, "[0]")
		if err != nil {
			lotName = ""
		}
		lotName = strip.StripTags(lotName)
		idLot := 0
		stmtl, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, currency = ?, lot_name = ?", prefix))
		resl, err := stmtl.Exec(idTender, LotNumber, "", lotName)
		stmtl.Close()
		if err != nil {
			logging("Ошибка вставки lot", err)
			return
		}
		id, _ := resl.LastInsertId()
		idLot = int(id)
		LotNumber++
		purName, err := jsonparser.GetString(value, "[1]")
		if err != nil {
			purName = tn.purName
		}
		okei, err := jsonparser.GetString(value, "[3]")
		if err != nil {
			okei = ""
		}
		quant, err := jsonparser.GetString(value, "[4]")
		if err != nil {
			quant = ""
		}
		stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, name = ?, okpd2_code = ?, sum = ?, quantity_value = ?, customer_quantity_value = ?, okei = ?", prefix))
		_, errr := stmtr.Exec(idLot, idCustomer, purName, "", "", quant, quant, okei)
		stmtr.Close()
		if errr != nil {
			logging("Ошибка вставки purchase_object", errr)
			return
		}
		delivTerm, err := jsonparser.GetString(value, "[2]")
		if err != nil {
			delivTerm = ""
		}
		if delivTerm != "" && delivTerm != "false" {
			stmtcr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer_requirement SET id_lot = ?, id_customer = ?, delivery_term = ?, delivery_place = ?, max_price = ?", prefix))
			_, err := stmtcr.Exec(idLot, idCustomer, delivTerm, "", "")
			stmtcr.Close()
			if err != nil {
				logging("Ошибка вставки purchase_object", errr)
				return
			}
		}

	}, "aaData")
	if err != nil {
		logging(err, "AAAAAAAAAAAAAAAAAAAAAA!!!!!!!!!!!!!!")
		return
	}
}
