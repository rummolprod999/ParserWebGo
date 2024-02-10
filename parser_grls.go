package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/shakinm/xlsReader/xls"
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type GrlsReader struct {
	Url         string
	Added       int
	AddedExcept int
}

func (t *GrlsReader) parsing() {
	logging("Процесс обновления базы запущен")
	p := t.downloadString()
	if p == "" {
		logging("get empty string", p)
		return
	}
	url := t.extractUrl(p)
	if url == "" {
		logging("get empty url", p)
		return
	}
	t.downloadArchive(url)
	logging("Процесс обновления базы завершен")
	logging(fmt.Sprintf("Добавлено %d элементов", t.Added))
	logging(fmt.Sprintf("Добавлено %d исключенных элементов", t.AddedExcept))

}

func (t *GrlsReader) downloadString() string {
	pageSource := DownloadPage(t.Url)
	return pageSource

}

func (t *GrlsReader) extractUrl(p string) string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(p))
	if err != nil {
		logging(err)
		return ""
	}
	aTag := doc.Find("#ctl00_plate_tdzip > button").First()
	if aTag == nil {
		logging("button tag not found")
		return ""
	}
	href, ok := aTag.Attr("onclick")
	if !ok {
		logging("href attr in a tag not found")
		return ""
	}
	href = strings.Replace(href, "go('", "", -1)
	href = strings.Replace(href, "'); return false;", "", -1)
	return fmt.Sprintf("https://grls.rosminzdrav.ru/%s", href)
}

func (t *GrlsReader) downloadArchive(url string) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	filePath := filepath.FromSlash(fmt.Sprintf("%s/%s/%s", dir, tempGrls, ArZir))
	err := DownloadFile(filePath, url)
	if err != nil {
		logging("file was not downloaded, exit", err)
		return
	}
	dirZip := filepath.FromSlash(fmt.Sprintf("%s/%s/", dir, tempGrls))
	err = Unzip(filePath, dirZip)
	if err != nil {
		logging("file was not unzipped, exit", err)
		return
	}
	files, err := FilePathWalkDir(dirZip)
	if err != nil {
		logging("filelist return error, exit", err)
		return
	}
	for _, f := range files {
		if strings.HasSuffix(f, "xls") {
			t.extractXlsData(f)
		}
		if strings.HasSuffix(f, "xlsx") {
			t.extractXlsxData(f)
		}
	}
}

func (t *GrlsReader) extractXlsData(nameFile string) {
	defer SaveStack()
	xlFile, err := xls.OpenFile(nameFile)
	if err != nil {
		logging("error open excel file, exit", err)
		return
	}
	sheet, _ := xlFile.GetSheet(0)
	t.insertToBase(sheet)
	sheetExcept, _ := xlFile.GetSheet(2)
	t.insertToBaseExcept(sheet, sheetExcept)

}

func (t *GrlsReader) extractXlsxData(nameFile string) {
	defer SaveStack()
	xlFile, err := excelize.OpenFile(nameFile)
	if err != nil {
		logging("error open excel file, exit", err)
		return
	}
	sheetInd := xlFile.GetSheetName(0)
	sheet, _ := xlFile.GetRows(sheetInd)
	sheetIndE := xlFile.GetSheetName(2)
	sheetE, _ := xlFile.GetRows(sheetIndE)
	t.insertToBaseNew(sheet, xlFile)
	t.insertToBaseExceptNew(sheet, sheetE, xlFile)

}

func (t *GrlsReader) insertToBaseExceptNew(sheet [][]string, sheetE [][]string, xlFile *excelize.File) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging(err)
		return
	}
	defer db.Close()
	_, err = db.Exec("DELETE FROM grls_except WHERE 1 = 1;")
	if err != nil {
		logging(err)
		return
	}
	c0 := sheet[0][0]
	datePubT := FindFromRegExp(c0, `(\d{2}\.\d{2}\.\d{4})`)
	if datePubT == "" {
		logging("datePub is empty")
	}
	datePub := getTimeMoscowLayout(datePubT, "02.01.2006")
	for r := 3; r < int(len(sheetE)); r++ {
		col := sheetE[r]
		mnn0 := col[0]
		mnn := ReplaceBadSymbols(mnn0)
		name0 := col[1]
		name := ReplaceBadSymbols(name0)
		form0 := col[2]
		form := ReplaceBadSymbols(form0)
		owner0 := col[3]
		owner := ReplaceBadSymbols(owner0)
		atx0 := col[4]
		atx := ReplaceBadSymbols(atx0)
		quantity0 := col[5]
		quantity := ReplaceBadSymbols(quantity0)
		maxPrice0 := col[6]
		maxPrice := strings.ReplaceAll(ReplaceBadSymbols(maxPrice0), ",", ".")
		firstPrice0 := col[7]
		firstPrice := strings.ReplaceAll(ReplaceBadSymbols(firstPrice0), ",", ".")
		ru0 := col[8]
		ru := ReplaceBadSymbols(ru0)
		dateReg0 := col[9]
		dateRegT := ReplaceBadSymbols(dateReg0)
		dateRegR := FindFromRegExp(dateRegT, `(\d{2}\.\d{2}\.\d{4})`)
		dateReg := getTimeMoscowLayout(dateRegR, "02.01.2006")
		code0 := col[10]
		code := ReplaceBadSymbols(code0)
		exceptCause0 := col[11]
		exceptCause := ReplaceBadSymbols(exceptCause0)
		exceptDate0 := col[12]
		exceptDateT := ReplaceBadSymbols(exceptDate0)
		if exceptDateT == "" {
			logging(fmt.Sprintf("exceptDate is empty, row %d, mnn - %s", r, mnn))
		}
		exceptDateR := FindFromRegExp(exceptDateT, `(\d{2}\-\d{2}\-\d{2})`)
		ea := strings.Split(exceptDateR, "-")
		exceptDateN := ""
		if (len(ea)) == 3 {
			exceptDateN = ea[1] + "." + ea[0] + ".20" + ea[2]
		}
		exceptDate := getTimeMoscowLayout(exceptDateN, "02.01.2006")
		if mnn == "" && name == "" && form == "" && owner == "" && atx == "" && quantity == "" && maxPrice == "" && firstPrice == "" && ru == "" && code == "" && exceptCause == "" && exceptDate.String() == "" {
			return
		}
		_, err := db.Exec("INSERT INTO grls_except (id, mnn, name, form, owner, atx, quantity, max_price, first_price, ru, date_reg, code, except_cause, except_date, date_pub) VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", mnn, name, form, owner, atx, quantity, maxPrice, firstPrice, ru, dateReg, code, exceptCause, exceptDate, datePub)
		t.AddedExcept++
		if err != nil {
			logging(err)
		}
	}
}

func (t *GrlsReader) insertToBaseNew(sheet [][]string, file *excelize.File) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging(err)
		return
	}
	defer db.Close()
	_, err = db.Exec("DELETE FROM grls WHERE 1 = 1;")
	if err != nil {
		logging(err)
		return
	}
	c0 := sheet[0][0]
	datePubT := FindFromRegExp(c0, `(\d{2}\.\d{2}\.\d{4})`)
	if datePubT == "" {
		logging("datePub is empty")
	}
	datePub := getTimeMoscowLayout(datePubT, "02.01.2006")
	for r := 3; r < int(len(sheet)); r++ {
		col := sheet[r]
		mnn0 := col[0]
		mnn := ReplaceBadSymbols(mnn0)
		name0 := col[1]
		name := ReplaceBadSymbols(name0)
		form0 := col[2]
		form := ReplaceBadSymbols(form0)
		owner0 := col[3]
		owner := ReplaceBadSymbols(owner0)
		atx0 := col[4]
		atx := ReplaceBadSymbols(atx0)
		quantity0 := col[5]
		quantity := ReplaceBadSymbols(quantity0)
		maxPrice0 := col[6]
		maxPrice := strings.ReplaceAll(ReplaceBadSymbols(maxPrice0), ",", ".")
		firstPrice0 := col[7]
		firstPrice := strings.ReplaceAll(ReplaceBadSymbols(firstPrice0), ",", ".")
		ru0 := col[8]
		ru := ReplaceBadSymbols(ru0)
		dateReg0 := col[9]
		dateRegT := ReplaceBadSymbols(dateReg0)
		dateRegR := FindFromRegExp(dateRegT, `(\d{2}\.\d{2}\.\d{4})`)
		dateReg := getTimeMoscowLayout(dateRegR, "02.01.2006")
		code0 := col[10]
		code := ReplaceBadSymbols(code0)
		if mnn == "" && name == "" && form == "" && owner == "" && atx == "" && quantity == "" && maxPrice == "" && firstPrice == "" && ru == "" && code == "" {
			return
		}
		_, err := db.Exec("INSERT INTO grls (id, mnn, name, form, owner, atx, quantity, max_price, first_price, ru, date_reg, code, date_pub) VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", mnn, name, form, owner, atx, quantity, maxPrice, firstPrice, ru, dateReg, code, datePub)
		t.Added++
		if err != nil {
			logging(err)
		}
	}
}

func (t *GrlsReader) insertToBase(sheet *xls.Sheet) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging(err)
		return
	}
	defer db.Close()
	_, err = db.Exec("DELETE FROM grls WHERE 1 = 1;")
	if err != nil {
		logging(err)
		return
	}
	d0, _ := sheet.GetRow(0)
	c0, _ := d0.GetCol(0)
	datePubT := FindFromRegExp(c0.GetString(), `(\d{2}\.\d{2}\.\d{4})`)
	if datePubT == "" {
		logging("datePub is empty")
	}
	datePub := getTimeMoscowLayout(datePubT, "02.01.2006")
	for r := 3; r <= int(sheet.GetNumberRows()); r++ {
		col, _ := sheet.GetRow(r)
		mnn0, _ := col.GetCol(0)
		mnn := ReplaceBadSymbols(mnn0.GetString())
		name0, _ := col.GetCol(1)
		name := ReplaceBadSymbols(name0.GetString())
		form0, _ := col.GetCol(2)
		form := ReplaceBadSymbols(form0.GetString())
		owner0, _ := col.GetCol(3)
		owner := ReplaceBadSymbols(owner0.GetString())
		atx0, _ := col.GetCol(4)
		atx := ReplaceBadSymbols(atx0.GetString())
		quantity0, _ := col.GetCol(5)
		quantity := ReplaceBadSymbols(quantity0.GetString())
		maxPrice0, _ := col.GetCol(6)
		maxPrice := strings.ReplaceAll(ReplaceBadSymbols(maxPrice0.GetString()), ",", ".")
		firstPrice0, _ := col.GetCol(7)
		firstPrice := strings.ReplaceAll(ReplaceBadSymbols(firstPrice0.GetString()), ",", ".")
		ru0, _ := col.GetCol(8)
		ru := ReplaceBadSymbols(ru0.GetString())
		dateReg0, _ := col.GetCol(9)
		dateRegT := ReplaceBadSymbols(dateReg0.GetString())
		dateRegR := FindFromRegExp(dateRegT, `(\d{2}\.\d{2}\.\d{4})`)
		dateReg := getTimeMoscowLayout(dateRegR, "02.01.2006")
		code0, _ := col.GetCol(10)
		code := ReplaceBadSymbols(code0.GetString())
		if mnn == "" && name == "" && form == "" && owner == "" && atx == "" && quantity == "" && maxPrice == "" && firstPrice == "" && ru == "" && code == "" {
			return
		}
		_, err := db.Exec("INSERT INTO grls (id, mnn, name, form, owner, atx, quantity, max_price, first_price, ru, date_reg, code, date_pub) VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", mnn, name, form, owner, atx, quantity, maxPrice, firstPrice, ru, dateReg, code, datePub)
		t.Added++
		if err != nil {
			logging(err)
		}
	}
}

func (t *GrlsReader) insertToBaseExcept(sheet0 *xls.Sheet, sheet *xls.Sheet) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging(err)
		return
	}
	defer db.Close()
	_, err = db.Exec("DELETE FROM grls_except WHERE 1 = 1;")
	if err != nil {
		logging(err)
		return
	}
	d0, _ := sheet0.GetRow(0)
	c0, _ := d0.GetCol(0)
	datePubT := FindFromRegExp(c0.GetString(), `(\d{2}\.\d{2}\.\d{4})`)
	if datePubT == "" {
		logging("datePub is empty")
	}
	datePub := getTimeMoscowLayout(datePubT, "02.01.2006")

	for r := 3; r <= int(sheet.GetNumberRows()); r++ {
		col, _ := sheet.GetRow(r)
		mnn0, _ := col.GetCol(0)
		mnn := ReplaceBadSymbols(mnn0.GetString())
		name0, _ := col.GetCol(1)
		name := ReplaceBadSymbols(name0.GetString())
		form0, _ := col.GetCol(2)
		form := ReplaceBadSymbols(form0.GetString())
		owner0, _ := col.GetCol(3)
		owner := ReplaceBadSymbols(owner0.GetString())
		atx0, _ := col.GetCol(4)
		atx := ReplaceBadSymbols(atx0.GetString())
		quantity0, _ := col.GetCol(5)
		quantity := ReplaceBadSymbols(quantity0.GetString())
		maxPrice0, _ := col.GetCol(6)
		maxPrice := strings.ReplaceAll(ReplaceBadSymbols(maxPrice0.GetString()), ",", ".")
		firstPrice0, _ := col.GetCol(7)
		firstPrice := strings.ReplaceAll(ReplaceBadSymbols(firstPrice0.GetString()), ",", ".")
		ru0, _ := col.GetCol(8)
		ru := ReplaceBadSymbols(ru0.GetString())
		dateReg0, _ := col.GetCol(9)
		dateRegT := ReplaceBadSymbols(dateReg0.GetString())
		dateRegR := FindFromRegExp(dateRegT, `(\d{2}\.\d{2}\.\d{4})`)
		dateReg := getTimeMoscowLayout(dateRegR, "02.01.2006")
		code0, _ := col.GetCol(10)
		code := ReplaceBadSymbols(code0.GetString())
		exceptCause0, _ := col.GetCol(11)
		exceptCause := ReplaceBadSymbols(exceptCause0.GetString())
		exceptDate0, _ := col.GetCol(13)
		exceptDateT := ReplaceBadSymbols(exceptDate0.GetString())
		if exceptDateT == "" {
			logging(fmt.Sprintf("exceptDate is empty, row %d, mnn - %s", r, mnn))
		}
		c, _ := strconv.ParseInt(exceptDateT, 10, 64)
		exceptDate := getTimeMoscowLayout("01.01.1900", "02.01.2006")
		exceptDate = exceptDate.AddDate(0, 0, int(c-2))
		if mnn == "" && name == "" && form == "" && owner == "" && atx == "" && quantity == "" && maxPrice == "" && firstPrice == "" && ru == "" && code == "" && exceptCause == "" && exceptDate.String() == "" {
			return
		}
		_, err := db.Exec("INSERT INTO grls_except (id, mnn, name, form, owner, atx, quantity, max_price, first_price, ru, date_reg, code, except_cause, except_date, date_pub) VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", mnn, name, form, owner, atx, quantity, maxPrice, firstPrice, ru, dateReg, code, exceptCause, exceptDate, datePub)
		t.AddedExcept++
		if err != nil {
			logging(err)
		}
	}
}
