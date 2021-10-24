package main

type protocol struct {
	RegistryNumber           string       `xml:"registryNumber"`
	Url                      string       `xml:"url_to_showcase"`
	IdProtocol               string       `xml:"id,attr"`
	DatePublished            string       `xml:"datePublished"`
	DateUpdated              string       `xml:"dateUpdated"`
	PurchaseObjectInfo       string       `xml:"title"`
	ProcedureTypeId          string       `xml:"procedureType>id"`
	ProcedureTypeName        string       `xml:"procedureType>title"`
	DateEndRegistration      string       `xml:"dateEndRegistration"`
	DateEndSecondPartsReview string       `xml:"dateEndSecondPartsReview"`
	Currency                 string       `xml:"currency"`
	Attachments              []attachment `xml:"documents>document"`
	Lots                     []lot        `xml:"lots>lot"`
	organizer
}

type fileProtocols struct {
	TotalPage         int        `xml:"Body>proceduresResponse>totalPage"`
	CurrentPage       int        `xml:"Body>proceduresResponse>currentPage"`
	HasMoreProcedures int        `xml:"Body>proceduresResponse>has_more_procedures"`
	Protocols         []protocol `xml:"Body>proceduresResponse>procedures>procedure"`
	Test              string     `xml:",innerxml"`
}

type organizer struct {
	OrganizerfullNameU string `xml:"organizer>fullName"`
	OrganizerIndexU    string `xml:"organizer>legal>index"`
	OrganizerRegionU   string `xml:"organizer>legal>region"`
	OrganizerCityU     string `xml:"organizer>legal>city"`
	OrganizerStreetU   string `xml:"organizer>legal>street"`
	OrganizerHouseU    string `xml:"organizer>legal>house"`
	OrganizerIndexP    string `xml:"organizer>postal>index"`
	OrganizerRegionP   string `xml:"organizer>postal>region"`
	OrganizerCityP     string `xml:"organizer>postal>city"`
	OrganizerStreetP   string `xml:"organizer>postal>street"`
	OrganizerHouseP    string `xml:"organizer>postal>house"`
	ContactEmail       string `xml:"contactEmail"`
	ContactPhone       string `xml:"contactPhone"`
	ContactPerson      string `xml:"contactPerson"`
}

type attachment struct {
	AttachName string `xml:"filename"`
	AttachUrl  string `xml:"file"`
}

type lot struct {
	LotNumber      int             `xml:"number"`
	LotSubject     string          `xml:"subject"`
	StartPrice     float64         `xml:"startPrice"`
	Okpd2Code      string          `xml:"nomenclature2>item>code"`
	OkpdName       string          `xml:"nomenclature2>item>name"`
	Customers      []customer      `xml:"customers>customer"`
	DeliveryPlaces []deliveryPlace `xml:"deliveryPlaces>deliveryPlace"`
	AttachmentsLot []attachmentLot `xml:"documents>document"`
	LotUnits       []lotUnits      `xml:"lotUnits>unit"`
}

type customer struct {
	FullName string `xml:"fullName"`
}

type deliveryPlace struct {
	Quantity string `xml:"quantity"`
	Term     string `xml:"term"`
	Address  string `xml:"address"`
}

type attachmentLot struct {
	AttachName string `xml:"filename"`
	AttachUrl  string `xml:"file"`
}

type lotUnits struct {
	Okp2Code string  `xml:"okpd2_code"`
	Name     string  `xml:"name"`
	Quantity float64 `xml:"quantity"`
}
