package models

import "github.com/PuerkitoBio/goquery"

type PriceType int

const (
	FIXED_PRICE PriceType = iota
	HOURLY
)

type Client struct {
	Name            string `json:"name" default:"Undefined"`
	PaymentVerified bool   `json:"is_payment_verified"`
	Stars           int    `json:"stars"`
	Spent           int    `json:"spent"`
	Location        string `json:"location"`
}

type Job struct {
	Parcell
	Title         string    `json:"title"`
	PriceType     PriceType `json:"price_type"`
	Budget        string    `json:"budget"`
	Description   string    `json:"description"`
	TimePosted    int64     `json:"time_posted"`
	ProposalCount int       `json:"proposal_count"`
	Tags          []string  `json:"tags"`
	Client        Client    `json:"client"`
}

func (job *Job) ImportData(info *goquery.Selection) {
	job.Title = info.Find(".up-n-link").Text()
}

func (job *Job) ToString() string {
	tempalteString := "<b>{{.Title}}</b>\n    {{.Description}}\n    "
	return ""
}
