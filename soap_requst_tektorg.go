package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var timeOffset = time.Now().Add(time.Hour * -24 * 30 * 6).Format("2006-01-02T15:04:05")
var templateSoap = `<soapenv:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:soap="http://api.tektorg.ru/procedures/soap">
   <soapenv:Header/>
   <soapenv:Body>
      <soap:procedures soapenv:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
         <soap:exportRequestType>
            <startDate xsi:type="xsd:dateTime">%s</startDate>
            <sectionCode xsi:type="xsd:string">%s</sectionCode>
            <page xsi:type="xsd:int">%d</page>
         </soap:exportRequestType>
      </soap:procedures>
   </soapenv:Body>
</soapenv:Envelope>`

func callSOAPClient(section string, page int) *fileProtocols {
	httpReq, err := generateSOAPRequest(section, page)
	if err != nil {
		logging(err)
		panic(err)
	}
	response, err := soapCall(httpReq)
	if err != nil {
		logging(err)
		panic(err)
	}
	return response
}

func generateSOAPRequest(section string, page int) (*http.Request, error) {
	buffer := &bytes.Buffer{}
	encoder := xml.NewEncoder(buffer)
	templateSoap := fmt.Sprintf(templateSoap, timeOffset, section, page)
	err := encoder.Encode(templateSoap)
	if err != nil {
		logging(err.Error())
		return nil, err
	}
	r, err := http.NewRequest(http.MethodPost, "https://api.tektorg.ru/procedures/soap", bytes.NewBuffer([]byte(templateSoap)))
	if err != nil {
		logging(err.Error())
		return nil, err
	}
	return r, nil
}

func soapCall(req *http.Request) (*fileProtocols, error) {
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r fileProtocols
	myString := string(body)
	xmlT := []byte(myString)
	_ = xml.Unmarshal(xmlT, &r)
	return &r, nil
}
