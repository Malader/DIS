package types

import "encoding/xml"

type CrackHashManagerRequest struct {
	XMLName    xml.Name `xml:"CrackHashManagerRequest"`
	RequestId  string   `xml:"RequestId"`
	PartNumber int      `xml:"PartNumber"`
	PartCount  int      `xml:"PartCount"`
	Hash       string   `xml:"Hash"`
	MaxLength  int      `xml:"MaxLength"`
	Alphabet   Alphabet `xml:"Alphabet"`
}

type Alphabet struct {
	XMLName xml.Name `xml:"Alphabet"`
	Symbols []string `xml:"symbols>symbol"`
}

type CrackHashWorkerResponse struct {
	XMLName    xml.Name `xml:"CrackHashWorkerResponse"`
	RequestId  string   `xml:"RequestId"`
	PartNumber int      `xml:"PartNumber"`
	Answers    struct {
		Words []string `xml:"words"`
	} `xml:"Answers"`
	DeliveryTag uint64 `xml:"-"`
}
