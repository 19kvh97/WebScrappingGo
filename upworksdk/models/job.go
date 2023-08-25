package models

import (
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

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

func (job *Job) AsMessage() string {
	tmplStr := `
<strong>Title:</strong> {{ .Title | html }}
<strong>Price Type:</strong> {{ .PriceType | html }}
<strong>Budget:</strong> {{ .Budget | html }}
<strong>Description:</strong> {{ .Description | html }}
<strong>Time Posted:</strong> {{ .TimePosted | html }}
<strong>Proposal Count:</strong> {{ .ProposalCount | html }}
`
	tmpl, err := template.New("jobTemplate").Parse(tmplStr)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return ""
	}

	var result strings.Builder
	err = tmpl.Execute(&result, job)
	if err != nil {
		log.Printf("err : %s", err.Error())
		return ""
	}

	return result.String()

}
