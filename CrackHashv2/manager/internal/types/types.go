package types

import "encoding/xml"

type CrackRequest struct {
	Hash      string `json:"hash"`
	MaxLength int    `json:"maxLength"`
}

type RequestResponse struct {
	RequestID string `json:"requestId"`
}

type StatusResponse struct {
	Status   string   `json:"status"`
	Data     []string `json:"data,omitempty"`
	Progress int      `json:"progress"`
}

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
