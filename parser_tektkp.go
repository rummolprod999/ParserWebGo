package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os/exec"
	"strings"
	"time"
)

var addtenderTektkp int
var updatetenderTektkp int

type parserTekTkp struct {
	maxPage int
	TypeFz  int
}

func (prot *parserTekTkp) parsing() {
	defer SaveStack()
	logging("Start parsing")
	protocols := callSOAPClient("17", 1)
	if protocols.TotalPage == 0 {
		logging("total page not found")
		return
	}
	prot.maxPage = protocols.TotalPage
	prot.parserProtocols(protocols)
	for i := 2; i <= prot.maxPage; i++ {
		protocols = callSOAPClient("17", i)
		prot.parserProtocols(protocols)
	}
	logging("End parsing")
	logging(fmt.Sprintf("Добавили тендеров %d", addtenderTektkp))
	logging(fmt.Sprintf("Обновили тендеров %d", updatetenderTektkp))
}

func (prot *parserTekTkp) parserProtocols(protocols *fileProtocols) {
	defer SaveStack()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Second * 3600)
	if err != nil {
		logging("Ошибка подключения к БД", err)
	}
	for _, pr := range protocols.Protocols {
		_ = prot.parserProtocol(pr, db)
	}
}
func (prot *parserTekTkp) parserProtocol(p protocol, db *sql.DB) error {
	layout := "2006-01-02T15:04:05"
	RegistryNumber := p.RegistryNumber
	DatePublishedS := p.DatePublished[:19]
	DateUpdatedS := p.DateUpdated
	if DateUpdatedS == "" {
		DateUpdatedS = DatePublishedS
	}
	DateUpdatedS = DateUpdatedS[:19]
	DatePublished, _ := time.Parse(layout, DatePublishedS)
	DateUpdated, _ := time.Parse(layout, DateUpdatedS)
	IdXml := p.IdProtocol
	Version := 0

	stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_tender FROM %stender WHERE id_xml = ? AND purchase_number = ? AND date_version = ? AND type_fz = ?", prefix))
	res, err := stmt.Query(IdXml, RegistryNumber, DateUpdated, prot.TypeFz)
	stmt.Close()
	if err != nil {
		logging("Ошибка выполения запроса", err)
		return err
	}
	if res.Next() {
		res.Close()
		return nil
	}
	res.Close()
	var cancelStatus = 0
	var updated = false
	if RegistryNumber != "" {
		stmt, err := db.Prepare(fmt.Sprintf("SELECT id_tender, date_version FROM %stender WHERE purchase_number = ? AND cancel=0 AND type_fz = ?", prefix))
		rows, err := stmt.Query(RegistryNumber, prot.TypeFz)
		stmt.Close()
		if err != nil {
			logging("Ошибка выполения запроса", err)
			return err
		}
		for rows.Next() {
			updated = true
			var idTender int
			var dateVersion time.Time
			err = rows.Scan(&idTender, &dateVersion)
			if err != nil {
				logging("Ошибка чтения результата запроса", err)
				return err
			}
			//fmt.Println(DateUpdated.Sub(dateVersion))
			if dateVersion.Sub(DateUpdated) <= 0 {
				stmtupd, _ := db.Prepare(fmt.Sprintf("UPDATE %stender SET cancel=1 WHERE id_tender = ?", prefix))
				_, err = stmtupd.Exec(idTender)
				stmtupd.Close()

			} else {
				cancelStatus = 1
			}

		}
		rows.Close()
	}
	Href := p.Url
	PurchaseObjectInfo := p.PurchaseObjectInfo
	NoticeVersion := ""
	PrintForm := Href
	IdOrganizer := 0
	OrganizerfullName := p.OrganizerfullNameU
	if OrganizerfullName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_organizer FROM %sorganizer WHERE full_name LIKE ? LIMIT 1", prefix))
		rows, err := stmt.Query(OrganizerfullName)
		stmt.Close()
		if err != nil {
			logging("Ошибка выполения запроса", err)
			return err
		}
		if rows.Next() {
			err = rows.Scan(&IdOrganizer)
			if err != nil {
				logging("Ошибка чтения результата запроса", err)
				return err
			}
			rows.Close()
		} else {
			rows.Close()
			OrgPostAddress := strings.TrimSpace(fmt.Sprintf("%s %s %s %s %s", p.OrganizerIndexP, p.OrganizerRegionP, p.OrganizerCityP, p.OrganizerStreetP, p.OrganizerHouseP))
			OrgUrAddress := strings.TrimSpace(fmt.Sprintf("%s %s %s %s %s", p.OrganizerIndexU, p.OrganizerRegionU, p.OrganizerCityU, p.OrganizerStreetU, p.OrganizerHouseU))
			ContactEmail := p.ContactEmail
			ContactPhone := p.ContactPhone
			ContactPerson := p.ContactPerson
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sorganizer SET full_name = ?, post_address = ?, fact_address = ?, contact_email = ?, contact_phone = ?, contact_person = ?", prefix))
			res, err := stmt.Exec(OrganizerfullName, OrgPostAddress, OrgUrAddress, ContactEmail, ContactPhone, ContactPerson)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки организатора", err)
				return err
			}
			id, err := res.LastInsertId()
			IdOrganizer = int(id)
		}

	}
	IdPlacingWay := 0
	PwCode := p.ProcedureTypeId
	PwName := p.ProcedureTypeName
	if PwCode != "" && PwName != "" {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_placing_way FROM %splacing_way WHERE code = ? AND name = ? LIMIT 1", prefix))
		rows, err := stmt.Query(PwCode, PwName)
		stmt.Close()
		if err != nil {
			logging("Ошибка выполения запроса", err)
			return err
		}
		if rows.Next() {
			err = rows.Scan(&IdPlacingWay)
			if err != nil {
				logging("Ошибка чтения результата запроса", err)
				return err
			}
			rows.Close()
		} else {
			rows.Close()
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %splacing_way SET code= ?, name= ?", prefix))
			res, err := stmt.Exec(PwCode, PwName)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки placing way", err)
				return err
			}
			id, err := res.LastInsertId()
			IdPlacingWay = int(id)
		}
	}

	IdEtp := 0
	etpName := "ТЭК Торг  Роснефть - Запросы (Т)КП"
	etpUrl := "https://www.tektorg.ru/rosnefttkp/procedures"
	if true {
		stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_etp FROM %setp WHERE name = ? AND url = ? LIMIT 1", prefix))
		rows, err := stmt.Query(etpName, etpUrl)
		stmt.Close()
		if err != nil {
			logging("Ошибка выполения запроса", err)
			return err
		}
		if rows.Next() {
			err = rows.Scan(&IdEtp)
			if err != nil {
				logging("Ошибка чтения результата запроса", err)
				return err
			}
			rows.Close()
		} else {
			rows.Close()
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %setp SET name = ?, url = ?, conf=0", prefix))
			res, err := stmt.Exec(etpName, etpUrl)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки etp", err)
				return err
			}
			id, err := res.LastInsertId()
			IdEtp = int(id)
		}
	}

	var EndDate, BiddingDate, ScoringDate = time.Time{}, time.Time{}, time.Time{}
	EndDateS := p.DateEndRegistration
	if EndDateS != "" {
		EndDate, _ = time.Parse(layout, EndDateS[:19])
	}
	ScoringDateS := p.DateEndSecondPartsReview
	if ScoringDateS != "" {
		ScoringDate, _ = time.Parse(layout, ScoringDateS[:19])
	}
	idTender := 0
	stmtt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %stender SET id_region = 0, id_xml = ?, purchase_number = ?, doc_publish_date = ?, href = ?, purchase_object_info = ?, type_fz = ?, id_organizer = ?, id_placing_way = ?, id_etp = ?, end_date = ?, scoring_date = ?, bidding_date = ?, cancel = ?, date_version = ?, num_version = ?, notice_version = ?, xml = ?, print_form = ?", prefix))
	rest, err := stmtt.Exec(IdXml, RegistryNumber, DatePublished, Href, PurchaseObjectInfo, prot.TypeFz, IdOrganizer, IdPlacingWay, IdEtp, EndDate, ScoringDate, BiddingDate, cancelStatus, DateUpdated, Version, NoticeVersion, Href, PrintForm)
	stmtt.Close()
	if err != nil {
		logging("Ошибка вставки tender", err)
		return err
	}
	idt, err := rest.LastInsertId()
	idTender = int(idt)
	if updated {
		updatetenderTektkp++
	} else {
		addtenderTektkp++
	}
	for _, att := range p.Attachments {
		attachName := att.AttachName
		attachUrl := att.AttachUrl
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", prefix))
		_, err := stmt.Exec(idTender, attachName, attachUrl)
		stmt.Close()
		if err != nil {
			logging("Ошибка вставки attachment", err)
			return err
		}
	}

	for _, lot := range p.Lots {
		LotNumber := lot.LotNumber
		MaxPrice := lot.StartPrice
		Currency := p.Currency
		idLot := 0
		stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %slot SET id_tender = ?, lot_number = ?, max_price = ?, currency = ?", prefix))
		res, err := stmt.Exec(idTender, LotNumber, MaxPrice, Currency)
		stmt.Close()
		if err != nil {
			logging("Ошибка вставки lot", err)
			return err
		}
		id, _ := res.LastInsertId()
		idLot = int(id)
		for _, attL := range lot.AttachmentsLot {
			attachName := attL.AttachName
			attachUrl := attL.AttachUrl
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %sattachment SET id_tender = ?, file_name = ?, url = ?", prefix))
			_, err := stmt.Exec(idTender, attachName, attachUrl)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки attachmentLot", err)
				return err
			}
		}
		idCustomer := 0
		if len(lot.Customers) > 0 {
			if lot.Customers[0].FullName != "" {
				stmt, _ := db.Prepare(fmt.Sprintf("SELECT id_customer FROM %scustomer WHERE full_name LIKE ? LIMIT 1", prefix))
				rows, err := stmt.Query(lot.Customers[0].FullName)
				stmt.Close()
				if err != nil {
					logging("Ошибка выполения запроса", err)
					return err
				}
				if rows.Next() {
					err = rows.Scan(&idCustomer)
					if err != nil {
						logging("Ошибка чтения результата запроса", err)
						return err
					}
					rows.Close()
				} else {
					rows.Close()
					out, err := exec.Command("uuidgen").Output()
					if err != nil {
						logging("Ошибка генерации UUID", err)
						return err
					}
					stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer SET full_name = ?, is223=1, reg_num = ?", prefix))
					res, err := stmt.Exec(lot.Customers[0].FullName, out)
					stmt.Close()
					if err != nil {
						logging("Ошибка вставки организатора", err)
						return err
					}
					id, err := res.LastInsertId()
					idCustomer = int(id)
				}
			}
		}

		for _, cusR := range lot.DeliveryPlaces {
			deliveryPlace := cusR.Address
			deliveryTerm := cusR.Term
			stmt, _ := db.Prepare(fmt.Sprintf("INSERT INTO %scustomer_requirement SET id_lot = ?, id_customer = ?, delivery_place = ?, delivery_term = ?", prefix))
			_, err := stmt.Exec(idLot, idCustomer, deliveryPlace, deliveryTerm)
			stmt.Close()
			if err != nil {
				logging("Ошибка вставки customer_requirement", err)
				return err

			}
		}
		QuantityValue := ""
		if len(lot.DeliveryPlaces) == 1 {
			QuantityValue = lot.DeliveryPlaces[0].Quantity
		}

		okpd2Code := lot.Okpd2Code
		okpdName := lot.OkpdName
		okpd2GroupCode, okpd2GroupLevel1Code := GetOkpd(okpd2Code)
		stmtr, _ := db.Prepare(fmt.Sprintf("INSERT INTO %spurchase_object SET id_lot = ?, id_customer = ?, okpd2_code = ?, okpd2_group_code = ?, okpd2_group_level1_code = ?, okpd_name = ?, name = ?, quantity_value = ?, customer_quantity_value = ?", prefix))
		_, errr := stmtr.Exec(idLot, idCustomer, okpd2Code, okpd2GroupCode, okpd2GroupLevel1Code, okpdName, lot.LotSubject, QuantityValue, QuantityValue)
		stmtr.Close()
		if errr != nil {
			logging("Ошибка вставки purchase_object", errr)
			return err
		}

	}
	e := TenderKwords(db, idTender)
	if e != nil {
		logging("Ошибка обработки TenderKwords", e)
	}

	e1 := AddVerNumber(db, RegistryNumber, prot.TypeFz)
	if e1 != nil {
		logging("Ошибка обработки AddVerNumber", e1)
	}
	return nil
}
