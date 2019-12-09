package main

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/tls"
	"database/sql"
	"encoding/hex"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

func SaveStack() {
	if p := recover(); p != nil {
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		file, err := os.OpenFile(string(FileLog), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		defer file.Close()
		if err != nil {
			fmt.Println("Ошибка записи stack log", err)
			return
		}
		fmt.Fprintln(file, fmt.Sprintf("Time %v", time.Now()))
		fmt.Fprintln(file, fmt.Sprintf("Fatal Error %v", p))
		fmt.Fprintf(file, "%v  ", string(buf[:n]))
	}

}

func DownloadPage1251(url string) string {
	var st string
	count := 0
	for {
		if count > 50 {
			Logging(fmt.Sprintf("Не скачали файл за %d попыток", count))
			return st
		}
		st = GetPage(url)
		if st == "" {
			count++
			Logging("Получили пустую страницу", url)
			continue
		}
		dec := charmap.Windows1251.NewDecoder()
		newBody := make([]byte, len(st)*2)
		n, _, err := dec.Transform(newBody, []byte(st), false)
		if err != nil {
			panic(err)
		}
		newBody = newBody[:n]
		return string(newBody)
	}
}

func DownloadPage1251NoCheckSSL(url string) string {
	var st string
	count := 0
	for {
		if count > 50 {
			Logging(fmt.Sprintf("Не скачали файл за %d попыток", count))
			return st
		}
		st = GetPageNoSSL(url)
		if st == "" {
			count++
			Logging("Получили пустую страницу", url)
			continue
		}
		dec := charmap.Windows1251.NewDecoder()
		newBody := make([]byte, len(st)*2)
		n, _, err := dec.Transform(newBody, []byte(st), false)
		if err != nil {
			panic(err)
		}
		newBody = newBody[:n]
		return string(newBody)
	}
}

func DownloadPage(url string) string {
	count := 0
	var st string
	for {
		//fmt.Println("Start download file")
		if count > 50 {
			Logging(fmt.Sprintf("Не скачали файл за %d попыток %s", count, url))
			return st
		}
		st = GetPage(url)
		if st == "" {
			count++
			Logging("Получили пустую страницу", url)
			time.Sleep(time.Second * 5)
			continue
		}
		return st

	}
	return st
}

func DownloadPageGzip(url string) string {
	count := 0
	var st string
	for {
		//fmt.Println("Start download file")
		if count > 50 {
			Logging(fmt.Sprintf("Не скачали файл за %d попыток %s", count, url))
			return st
		}
		st = GetPageGzip(url)
		if st == "" {
			count++
			Logging("Получили пустую страницу", url)
			time.Sleep(time.Second * 5)
			continue
		}
		return st

	}
	return st
}

func DownloadPageWithUA(url string) string {
	count := 0
	var st string
	for {
		//fmt.Println("Start download file")
		if count > 50 {
			Logging(fmt.Sprintf("Не скачали файл за %d попыток %s", count, url))
			return st
		}
		st = GetPageUA(url)
		if st == "" {
			count++
			Logging("Получили пустую страницу", url)
			time.Sleep(time.Second * 5)
			continue
		}
		return st

	}
	return st
}

func DownloadPageWithUAIceTrade(url string) string {
	count := 0
	var st string
	for {
		//fmt.Println("Start download file")
		if count > 50 {
			Logging(fmt.Sprintf("Не скачали файл за %d попыток %s", count, url))
			return st
		}
		st = GetPageUA(url)
		if st == "" {
			count++
			Logging("Получили пустую страницу", url)
			time.Sleep(time.Second * 5)
			continue
		}
		if strings.Contains(st, "Ваш IP-адрес попал в список заблокированных") {
			count++
			Logging("IP was banned, i am sleep", url)
			time.Sleep(time.Second * 310)
			continue
		}
		return st

	}
	return st
}

func GetPage(url string) (ret string) {
	var st string
	defer func() {
		if r := recover(); r != nil {
			Logging(fmt.Sprintf("was panic, recovered value: %v", r))
			ret = ""
		}
	}()
	resp, err := http.Get(url)
	if err != nil {
		Logging("Ошибка response", url, err)
		return st
	}
	defer resp.Body.Close()
	if err != nil {
		Logging("Ошибка скачивания", url, err)
		return st
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logging("Ошибка чтения", url, err)
		return st
	}

	return string(body)
}

func GetPageNoSSL(url string) string {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	var st string
	resp, err := http.Get(url)
	if err != nil {
		Logging("Ошибка response", url, err)
		return st
	}
	defer resp.Body.Close()
	if err != nil {
		Logging("Ошибка скачивания", url, err)
		return st
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logging("Ошибка чтения", url, err)
		return st
	}

	return string(body)
}

func GetPageGzip(url string) string {
	var st string
	client := new(http.Client)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Logging("Ошибка request", url, err)
		return st
	}
	request.Header.Add("Accept-Encoding", "gzip")

	resp, err := client.Do(request)
	if err != nil {
		Logging("Ошибка resp", url, err)
		return st
	}
	defer resp.Body.Close()
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			Logging("Ошибка reader", url, err)
			return st
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	return buf.String()
}

func GetPageUA(url string) (ret string) {
	var st string
	defer func() {
		if r := recover(); r != nil {
			Logging(fmt.Sprintf("was panic, recovered value: %v", r))
			ret = ""
		}
	}()
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Logging("Ошибка request", url, err)
		return st
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko)")
	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		Logging("Ошибка скачивания", url, err)
		return st
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logging("Ошибка чтения", url, err)
		return st
	}

	return string(body)
}

func getTimeMoscow(st string) time.Time {
	var p = time.Time{}
	location, _ := time.LoadLocation("Europe/Moscow")
	p, err := time.ParseInLocation("02.01.2006 15:04", st, location)
	if err != nil {
		Logging(err)
		return time.Time{}
	}

	return p
}

func TenderKwords(db *sql.DB, idTender int) error {
	resString := ""
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT DISTINCT po.name, po.okpd_name FROM %spurchase_object AS po LEFT JOIN %slot AS l ON l.id_lot = po.id_lot WHERE l.id_tender = ?", Prefix, Prefix))
	rows, err := stmt.Query(idTender)
	stmt.Close()
	if err != nil {
		Logging("Ошибка выполения запроса", err)
		return err
	}
	for rows.Next() {
		var name sql.NullString
		var okpdName sql.NullString
		err = rows.Scan(&name, &okpdName)
		if err != nil {
			Logging("Ошибка чтения результата запроса", err)
			return err
		}
		if name.Valid {
			resString = fmt.Sprintf("%s %s ", resString, name.String)
		}
		if okpdName.Valid {
			resString = fmt.Sprintf("%s %s ", resString, okpdName.String)
		}
	}
	rows.Close()
	stmt1, _ := db.Prepare(fmt.Sprintf("SELECT DISTINCT file_name FROM %sattachment WHERE id_tender = ?", Prefix))
	rows1, err := stmt1.Query(idTender)
	stmt1.Close()
	if err != nil {
		Logging("Ошибка выполения запроса", err)
		return err
	}
	for rows1.Next() {
		var attName sql.NullString
		err = rows1.Scan(&attName)
		if err != nil {
			Logging("Ошибка чтения результата запроса", err)
			return err
		}
		if attName.Valid {
			resString = fmt.Sprintf("%s %s ", resString, attName.String)
		}
	}
	rows1.Close()
	idOrg := 0
	stmt2, _ := db.Prepare(fmt.Sprintf("SELECT purchase_object_info, id_organizer FROM %stender WHERE id_tender = ?", Prefix))
	rows2, err := stmt2.Query(idTender)
	stmt2.Close()
	if err != nil {
		Logging("Ошибка выполения запроса", err)
		return err
	}
	for rows2.Next() {
		var idOrgNull sql.NullInt64
		var purOb sql.NullString
		err = rows2.Scan(&purOb, &idOrgNull)
		if err != nil {
			Logging("Ошибка чтения результата запроса", err)
			return err
		}
		if idOrgNull.Valid {
			idOrg = int(idOrgNull.Int64)
		}
		if purOb.Valid {
			resString = fmt.Sprintf("%s %s ", resString, purOb.String)
		}

	}
	rows2.Close()
	if idOrg != 0 {
		stmt3, _ := db.Prepare(fmt.Sprintf("SELECT full_name, inn FROM %sorganizer WHERE id_organizer = ?", Prefix))
		rows3, err := stmt3.Query(idOrg)
		stmt3.Close()
		if err != nil {
			Logging("Ошибка выполения запроса", err)
			return err
		}
		for rows3.Next() {
			var innOrg sql.NullString
			var nameOrg sql.NullString
			err = rows3.Scan(&nameOrg, &innOrg)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
				return err
			}
			if innOrg.Valid {

				resString = fmt.Sprintf("%s %s ", resString, innOrg.String)
			}
			if nameOrg.Valid {
				resString = fmt.Sprintf("%s %s ", resString, nameOrg.String)
			}

		}
		rows3.Close()
	}
	stmt4, _ := db.Prepare(fmt.Sprintf("SELECT DISTINCT cus.inn, cus.full_name FROM %scustomer AS cus LEFT JOIN %spurchase_object AS po ON cus.id_customer = po.id_customer LEFT JOIN %slot AS l ON l.id_lot = po.id_lot WHERE l.id_tender = ?", Prefix, Prefix, Prefix))
	rows4, err := stmt4.Query(idTender)
	stmt4.Close()
	if err != nil {
		Logging("Ошибка выполения запроса", err)
		return err
	}
	for rows4.Next() {
		var innC sql.NullString
		var fullNameC sql.NullString
		err = rows4.Scan(&innC, &fullNameC)
		if err != nil {
			Logging("Ошибка чтения результата запроса", err)
			return err
		}
		if innC.Valid {

			resString = fmt.Sprintf("%s %s ", resString, innC.String)
		}
		if fullNameC.Valid {
			resString = fmt.Sprintf("%s %s ", resString, fullNameC.String)
		}
	}
	rows4.Close()
	re := regexp.MustCompile(`\s+`)
	resString = re.ReplaceAllString(resString, " ")
	stmtr, _ := db.Prepare(fmt.Sprintf("UPDATE %stender SET tender_kwords = ? WHERE id_tender = ?", Prefix))
	_, errr := stmtr.Exec(resString, idTender)
	stmtr.Close()
	if errr != nil {
		Logging("Ошибка вставки TenderKwords", errr)
		return err
	}
	return nil
}

func AddVerNumber(db *sql.DB, RegistryNumber string, typeFz int) error {
	verNum := 1
	mapTenders := make(map[int]int)
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE purchase_number = ? AND type_fz = ? ORDER BY UNIX_TIMESTAMP(date_version) ASC", Prefix))
	rows, err := stmt.Query(RegistryNumber, typeFz)
	stmt.Close()
	if err != nil {
		Logging("Ошибка выполения запроса", err)
		return err
	}
	for rows.Next() {
		var rNum int
		err = rows.Scan(&rNum)
		if err != nil {
			Logging("Ошибка чтения результата запроса", err)
			return err
		}
		mapTenders[verNum] = rNum
		verNum++
	}
	rows.Close()
	for vn, idt := range mapTenders {
		stmtr, _ := db.Prepare(fmt.Sprintf("UPDATE %stender SET num_version = ? WHERE id_tender = ?", Prefix))
		_, errr := stmtr.Exec(vn, idt)
		stmtr.Close()
		if errr != nil {
			Logging("Ошибка вставки NumVersion", errr)
			return err
		}
	}

	return nil
}

func getDateDixy(s string) time.Time {
	var p = time.Time{}
	if s != "" {
		dt := ""
		if strings.Contains(s, "янв") {
			dt = strings.Replace(s, "янв", "01", -1)
		} else if strings.Contains(s, "фев") {
			dt = strings.Replace(s, "фев", "02", -1)
		} else if strings.Contains(s, "мар") {
			dt = strings.Replace(s, "мар", "03", -1)
		} else if strings.Contains(s, "апр") {
			dt = strings.Replace(s, "апр", "04", -1)
		} else if strings.Contains(s, "май") {
			dt = strings.Replace(s, "май", "05", -1)
		} else if strings.Contains(s, "июн") {
			dt = strings.Replace(s, "июн", "06", -1)
		} else if strings.Contains(s, "июл") {
			dt = strings.Replace(s, "июл", "07", -1)
		} else if strings.Contains(s, "авг") {
			dt = strings.Replace(s, "авг", "08", -1)
		} else if strings.Contains(s, "сен") {
			dt = strings.Replace(s, "сен", "09", -1)
		} else if strings.Contains(s, "окт") {
			dt = strings.Replace(s, "окт", "10", -1)
		} else if strings.Contains(s, "ноя") {
			dt = strings.Replace(s, "ноя", "11", -1)
		} else if strings.Contains(s, "дек") {
			dt = strings.Replace(s, "дек", "12", -1)
		}
		p = getTimeMoscowLayout(dt, "02 01 2006")
	}
	return p
}

func getDateCpc(s string) time.Time {
	var p = time.Time{}
	if s != "" {
		dt := ""
		if strings.Contains(s, "янв") {
			dt = strings.Replace(s, "января", "01", -1)
		} else if strings.Contains(s, "фев") {
			dt = strings.Replace(s, "февраля", "02", -1)
		} else if strings.Contains(s, "мар") {
			dt = strings.Replace(s, "марта", "03", -1)
		} else if strings.Contains(s, "апр") {
			dt = strings.Replace(s, "апреля", "04", -1)
		} else if strings.Contains(s, "мая") {
			dt = strings.Replace(s, "мая", "05", -1)
		} else if strings.Contains(s, "июн") {
			dt = strings.Replace(s, "июня", "06", -1)
		} else if strings.Contains(s, "июл") {
			dt = strings.Replace(s, "июля", "07", -1)
		} else if strings.Contains(s, "авг") {
			dt = strings.Replace(s, "августа", "08", -1)
		} else if strings.Contains(s, "сен") {
			dt = strings.Replace(s, "сентября", "09", -1)
		} else if strings.Contains(s, "окт") {
			dt = strings.Replace(s, "октября", "10", -1)
		} else if strings.Contains(s, "ноя") {
			dt = strings.Replace(s, "ноября", "11", -1)
		} else if strings.Contains(s, "дек") {
			dt = strings.Replace(s, "декабря", "12", -1)
		}
		p = getTimeMoscowLayout(dt, "02 01 2006")
	}
	return p
}

func getDateMonetka(s string) time.Time {
	var p = time.Time{}
	if s != "" {
		dt := ""
		if strings.Contains(s, "янв") {
			dt = strings.Replace(s, "января", "01", -1)
		} else if strings.Contains(s, "фев") {
			dt = strings.Replace(s, "февраля", "02", -1)
		} else if strings.Contains(s, "мар") {
			dt = strings.Replace(s, "марта", "03", -1)
		} else if strings.Contains(s, "апр") {
			dt = strings.Replace(s, "апреля", "04", -1)
		} else if strings.Contains(s, "мая") {
			dt = strings.Replace(s, "мая", "05", -1)
		} else if strings.Contains(s, "июн") {
			dt = strings.Replace(s, "июня", "06", -1)
		} else if strings.Contains(s, "июл") {
			dt = strings.Replace(s, "июля", "07", -1)
		} else if strings.Contains(s, "авг") {
			dt = strings.Replace(s, "августа", "08", -1)
		} else if strings.Contains(s, "сен") {
			dt = strings.Replace(s, "сентября", "09", -1)
		} else if strings.Contains(s, "окт") {
			dt = strings.Replace(s, "октября", "10", -1)
		} else if strings.Contains(s, "ноя") {
			dt = strings.Replace(s, "ноября", "11", -1)
		} else if strings.Contains(s, "дек") {
			dt = strings.Replace(s, "декабря", "12", -1)
		}
		p = getTimeMoscowLayout(dt, "2 01 2006")
	}
	return p
}

func getDateIes(s string) time.Time {
	var p = time.Time{}
	s = strings.ToLower(s)
	if s != "" {
		dt := ""
		if strings.Contains(s, "янв") {
			dt = strings.Replace(s, "январь", "01", -1)
		} else if strings.Contains(s, "фев") {
			dt = strings.Replace(s, "февраль", "02", -1)
		} else if strings.Contains(s, "мар") {
			dt = strings.Replace(s, "март", "03", -1)
		} else if strings.Contains(s, "апр") {
			dt = strings.Replace(s, "апрель", "04", -1)
		} else if strings.Contains(s, "мая") {
			dt = strings.Replace(s, "май", "05", -1)
		} else if strings.Contains(s, "июн") {
			dt = strings.Replace(s, "июнь", "06", -1)
		} else if strings.Contains(s, "июл") {
			dt = strings.Replace(s, "июль", "07", -1)
		} else if strings.Contains(s, "авг") {
			dt = strings.Replace(s, "август", "08", -1)
		} else if strings.Contains(s, "сен") {
			dt = strings.Replace(s, "сентябрь", "09", -1)
		} else if strings.Contains(s, "окт") {
			dt = strings.Replace(s, "октябрь", "10", -1)
		} else if strings.Contains(s, "ноя") {
			dt = strings.Replace(s, "ноябрь", "11", -1)
		} else if strings.Contains(s, "дек") {
			dt = strings.Replace(s, "декабрь", "12", -1)
		}
		dt = cleanString(dt)
		p = getTimeMoscowLayout(dt, "02 01 2006")
	}
	return p
}
func getTimeMoscowLayout(st string, l string) time.Time {
	var p = time.Time{}
	location, _ := time.LoadLocation("Europe/Moscow")
	p, err := time.ParseInLocation(l, st, location)
	if err != nil {
		Logging(err)
		return time.Time{}
	}

	return p
}

func getTimeMoscowLayoutIceTrade(st string, l string) time.Time {
	var p = time.Time{}
	location, _ := time.LoadLocation("Europe/Moscow")
	p, err := time.ParseInLocation(l, st, location)
	if err != nil {
		//Logging(err)
		return time.Time{}
	}

	return p
}
func findFromRegExp(s string, t string) string {
	r := ""
	re := regexp.MustCompile(t)
	match := re.FindStringSubmatch(s)
	if match != nil && len(match) > 1 {
		r = match[1]
	}
	return r
}

func findTwoFromRegExp(s string, t string) (string, string) {
	r := ""
	p := ""
	re := regexp.MustCompile(t)
	match := re.FindStringSubmatch(s)
	if match != nil && len(match) > 1 {
		r = match[1]
		p = match[2]
	}
	return r, p
}

func cleanString(s string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, " ")
}
func delallwhitespace(s string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, "")
}
func delallwhitespaceAtom(s string) string {
	re := regexp.MustCompile(`[^\d.]`)
	return re.ReplaceAllString(s, "")
}
func findFromRegExpDixy(s string, t string) string {
	r := ""
	re := regexp.MustCompile(t)
	match := re.FindStringSubmatch(s)
	if len(match) > 2 {
		r = fmt.Sprintf("%s %s", match[1], match[2])
	}
	return r
}

func GetConformity(conf string) int {
	s := strings.ToLower(conf)
	switch {
	case strings.Index(s, "открыт") != -1:
		return 5
	case strings.Index(s, "аукцион") != -1:
		return 1
	case strings.Index(s, "котиров") != -1:
		return 2
	case strings.Index(s, "предложен") != -1:
		return 3
	case strings.Index(s, "единств") != -1:
		return 4
	default:
		return 6
	}

}

func getPlacingWayId(pwName string, db *sql.DB) int {
	idPlacingWay := 0
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_placing_way FROM %splacing_way WHERE name = ? LIMIT 1", Prefix))
	rows, err := stmt.Query(pwName)
	stmt.Close()
	if err != nil {
		Logging("Ошибка выполения запроса", err)
		return 0
	}
	if rows.Next() {
		err = rows.Scan(&idPlacingWay)
		if err != nil {
			Logging("Ошибка чтения результата запроса", err)
			return 0
		}
		rows.Close()
	} else {
		rows.Close()
		conf := GetConformity(pwName)
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %splacing_way SET name = ?, conformity = ?", Prefix))
		res, err := stmt.Exec(pwName, conf)
		stmt.Close()
		if err != nil {
			Logging("Ошибка вставки placing way", err)
			return 0
		}
		id, err := res.LastInsertId()
		idPlacingWay = int(id)
		return idPlacingWay
	}
	return idPlacingWay
}

func getEtpId(etpName string, etpUrl string, db *sql.DB) int {
	IdEtp := 0
	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_etp FROM %setp WHERE name = ? AND url = ? LIMIT 1", Prefix))
	rows, err := stmt.Query(etpName, etpUrl)
	stmt.Close()
	if err != nil {
		Logging("Ошибка выполения запроса", err)
		return 0
	}
	if rows.Next() {
		err = rows.Scan(&IdEtp)
		if err != nil {
			Logging("Ошибка чтения результата запроса", err)
			return 0
		}
		rows.Close()
	} else {
		rows.Close()
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %setp SET name = ?, url = ?, conf=0", Prefix))
		res, err := stmt.Exec(etpName, etpUrl)
		stmt.Close()
		if err != nil {
			Logging("Ошибка вставки etp", err)
			return 0
		}
		id, err := res.LastInsertId()
		IdEtp = int(id)
		return IdEtp
	}
	return IdEtp
}

func getRegionId(addrName string, db *sql.DB) int {
	idReg := 0
	if addrName != "" {
		regT := getRegion(addrName)
		stmt, _ := db.Prepare("SELECT id FROM region WHERE name LIKE ?")
		rows, err := stmt.Query("%" + regT + "%")
		stmt.Close()
		if err != nil {
			Logging("Ошибка выполения запроса", err)
			return 0
		}
		if rows.Next() {
			err = rows.Scan(&idReg)
			if err != nil {
				Logging("Ошибка чтения результата запроса", err)
				return 0
			}
			rows.Close()
		} else {
			rows.Close()
		}
	}
	return idReg
}
func getRegion(sp string) string {
	sp = strings.ToLower(sp)
	s := ""
	switch {
	case strings.Contains(sp, "белгор"):
		s = "белгор"
	case strings.Contains(sp, "брянск"):
		s = "брянск"
	case strings.Contains(sp, "владимир"):
		s = "владимир"
	case strings.Contains(sp, "воронеж"):
		s = "воронеж"
	case strings.Contains(sp, "иванов"):
		s = "иванов"
	case strings.Contains(sp, "калужск"):
		s = "калужск"
	case strings.Contains(sp, "костром"):
		s = "костром"
	case strings.Contains(sp, "курск"):
		s = "курск"
	case strings.Contains(sp, "липецк"):
		s = "липецк"
	case strings.Contains(sp, "москва"):
		s = "москва"
	case strings.Contains(sp, "московск"):
		s = "московск"
	case strings.Contains(sp, "орлов"):
		s = "орлов"
	case strings.Contains(sp, "рязан"):
		s = "рязан"
	case strings.Contains(sp, "смолен"):
		s = "смолен"
	case strings.Contains(sp, "тамбов"):
		s = "тамбов"
	case strings.Contains(sp, "твер"):
		s = "твер"
	case strings.Contains(sp, "тульс"):
		s = "тульс"
	case strings.Contains(sp, "яросл"):
		s = "яросл"
	case strings.Contains(sp, "архан"):
		s = "архан"
	case strings.Contains(sp, "вологод"):
		s = "вологод"
	case strings.Contains(sp, "калинин"):
		s = "калинин"
	case strings.Contains(sp, "карел"):
		s = "карел"
	case strings.Contains(sp, "коми"):
		s = "коми"
	case strings.Contains(sp, "ленинг"):
		s = "ленинг"
	case strings.Contains(sp, "мурм"):
		s = "мурм"
	case strings.Contains(sp, "ненец"):
		s = "ненец"
	case strings.Contains(sp, "новгор"):
		s = "новгор"
	case strings.Contains(sp, "псков"):
		s = "псков"
	case strings.Contains(sp, "санкт"):
		s = "санкт"
	case strings.Contains(sp, "адыг"):
		s = "адыг"
	case strings.Contains(sp, "астрахан"):
		s = "астрахан"
	case strings.Contains(sp, "волгог"):
		s = "волгог"
	case strings.Contains(sp, "калмык"):
		s = "калмык"
	case strings.Contains(sp, "краснод"):
		s = "краснод"
	case strings.Contains(sp, "ростов"):
		s = "ростов"
	case strings.Contains(sp, "дагест"):
		s = "дагест"
	case strings.Contains(sp, "ингуш"):
		s = "ингуш"
	case strings.Contains(sp, "кабардин"):
		s = "кабардин"
	case strings.Contains(sp, "карача"):
		s = "карача"
	case strings.Contains(sp, "осети"):
		s = "осети"
	case strings.Contains(sp, "ставроп"):
		s = "ставроп"
	case strings.Contains(sp, "чечен"):
		s = "чечен"
	case strings.Contains(sp, "башкор"):
		s = "башкор"
	case strings.Contains(sp, "киров"):
		s = "киров"
	case strings.Contains(sp, "марий"):
		s = "марий"
	case strings.Contains(sp, "мордов"):
		s = "мордов"
	case strings.Contains(sp, "нижегор"):
		s = "нижегор"
	case strings.Contains(sp, "оренбур"):
		s = "оренбур"
	case strings.Contains(sp, "пензен"):
		s = "пензен"
	case strings.Contains(sp, "пермс"):
		s = "пермс"
	case strings.Contains(sp, "самар"):
		s = "самар"
	case strings.Contains(sp, "сарат"):
		s = "сарат"
	case strings.Contains(sp, "татарс"):
		s = "татарс"
	case strings.Contains(sp, "удмурт"):
		s = "удмурт"
	case strings.Contains(sp, "ульян"):
		s = "ульян"
	case strings.Contains(sp, "чуваш"):
		s = "чуваш"
	case strings.Contains(sp, "курган"):
		s = "курган"
	case strings.Contains(sp, "свердлов"):
		s = "свердлов"
	case strings.Contains(sp, "тюмен"):
		s = "тюмен"
	case strings.Contains(sp, "ханты"):
		s = "ханты"
	case strings.Contains(sp, "челяб"):
		s = "челяб"
	case strings.Contains(sp, "ямало"):
		s = "ямало"
	case strings.Contains(sp, "алтайск"):
		s = "алтайск"
	case strings.Contains(sp, "алтай"):
		s = "алтай"
	case strings.Contains(sp, "бурят"):
		s = "бурят"
	case strings.Contains(sp, "забайк"):
		s = "забайк"
	case strings.Contains(sp, "иркут"):
		s = "иркут"
	case strings.Contains(sp, "кемеров"):
		s = "кемеров"
	case strings.Contains(sp, "краснояр"):
		s = "краснояр"
	case strings.Contains(sp, "новосиб"):
		s = "новосиб"
	case strings.Contains(sp, "томск"):
		s = "томск"
	case strings.Contains(sp, "омск"):
		s = "омск"
	case strings.Contains(sp, "тыва"):
		s = "тыва"
	case strings.Contains(sp, "хакас"):
		s = "хакас"
	case strings.Contains(sp, "амурск"):
		s = "амурск"
	case strings.Contains(sp, "еврей"):
		s = "еврей"
	case strings.Contains(sp, "камчат"):
		s = "камчат"
	case strings.Contains(sp, "магад"):
		s = "магад"
	case strings.Contains(sp, "примор"):
		s = "примор"
	case strings.Contains(sp, "сахалин"):
		s = "сахалин"
	case strings.Contains(sp, "якут"):
		s = "якут"
	case strings.Contains(sp, "саха"):
		s = "саха"
	case strings.Contains(sp, "хабар"):
		s = "хабар"
	case strings.Contains(sp, "чукот"):
		s = "чукот"
	case strings.Contains(sp, "крым"):
		s = "крым"
	case strings.Contains(sp, "севастоп"):
		s = "севастоп"
	case strings.Contains(sp, "байкон"):
		s = "байкон"
	default:
		s = ""
	}
	return s
}

func GetMd5(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}
