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
	RequestId  string   `json:"RequestId"`
	PartNumber int      `json:"PartNumber"`
	PartCount  int      `json:"PartCount"`
	Hash       string   `json:"Hash"`
	MaxLength  int      `json:"MaxLength"`
	Alphabet   Alphabet `json:"Alphabet"`
}

type Alphabet struct {
	Symbols []string `json:"symbols"`
}

type CrackHashWorkerResponse struct {
	XMLName    xml.Name `xml:"CrackHashWorkerResponse"`
	RequestId  string   `xml:"RequestId"`
	PartNumber int      `xml:"PartNumber"`
	Answers    struct {
		Words []string `xml:"words"`
	} `xml:"Answers"`
}
