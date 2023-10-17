package models

import (
	"fmt"
	"html/template"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type PriceType int

const (
	FIXED_PRICE PriceType = iota
	HOURLY
)

func (pt PriceType) String() string {
	return []string{"Fixed-price", "Hourly"}[pt]
}

type Client struct {
	Name            string `json:"name" default:"Undefined"`
	PaymentVerified bool   `json:"is_payment_verified"`
	Stars           int    `json:"stars"`
	Spent           int    `json:"spent"`
	Location        string `json:"location"`
}

func (cl *Client) FromRawData(info *goquery.Selection) {
	if info.Find("[data-test=\"payment-verification-status\"]").Find(".text-muted").Text() == "Payment verified" {
		cl.PaymentVerified = true
	} else {
		cl.PaymentVerified = false
	}

	cl.Stars = len(info.Find(".up-rating.up-rating-sm.up-popper-trigger").Find(".up-rating-foreground").Nodes)
	spentAmountTxt := info.Find("[data-test=\"formatted-amount\"]").Text()
	spentAmount, err := parseSIStringToNumber(spentAmountTxt[4 : len(spentAmountTxt)-1])
	if err != nil {
		log.Printf("err: %v", err)
		cl.Spent = 0
	} else {
		cl.Spent = int(spentAmount)
	}
	cl.Location = info.Find("[data-test=\"client-country\"]").Find("strong").Text()
}

func parseSIStringToNumber(s string) (int64, error) {
	re := regexp.MustCompile(`(\d+M|\d+K|\d+)`)
	matches := re.FindAllString(s, -1)

	if len(matches) > 0 {
		match := matches[0]
		// Convert "K" to 1e3 and "M" to 1e6
		match = strings.ReplaceAll(match, "K", "e3")
		match = strings.ReplaceAll(match, "M", "e6")

		// Attempt to parse the modified string to a number
		num, err := strconv.ParseFloat(match, 64)
		if err != nil {
			return 0, err
		}

		// Convert the parsed float to an integer
		return int64(num), nil
	}

	return 0, fmt.Errorf("cant extract number from %v", s)
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
	Link          string    `json:"link"`
	Client        Client    `json:"client"`
}

func (job *Job) ImportData(info *goquery.Selection) {
	job.Title = info.Find(".up-n-link").Text()
	priceT := info.Find(".text-muted.display-inline-block.text-muted").Find("strong").Text()
	if priceT == "Fixed-price" {
		job.PriceType = FIXED_PRICE
		price := info.Find("[data-test=\"budget\"]").Text()
		if price != "" {
			job.Budget = strings.ReplaceAll(price, "\n", "")
		} else {
			job.Budget = "Not found"
		}
	} else {
		job.PriceType = HOURLY
		if len(priceT) > 8 {
			job.Budget = priceT[8:]
		} else {
			job.Budget = "Any"
		}
	}
	job.Description = info.Find("[data-test=\"job-description-text\"]").Text()
	job.TimePosted = toMilisecond(info.Find("[data-test=\"posted-on\"]").Text())
	job.ProposalCount = getProposalCount(info.Find("[data-test=\"proposals\"]").Text())
	var tags []string
	info.Find("a.up-skill-badge.text-muted").Each(func(i int, s *goquery.Selection) {
		tags = append(tags, s.Text())
	})

	job.Tags = tags
	info.Find("h2.my-0.p-sm-right.job-tile-title a.up-n-link").Each(func(i int, s *goquery.Selection) {
		url, ok := s.Attr("href")
		if ok {
			job.Link = url
		}
	})
	job.Client.FromRawData(info)
}

func (jb *Job) FormattedTimePosted() string {
	return time.UnixMilli(jb.TimePosted).Format("2006-01-02 15:04")
}

func getProposalCount(proposalText string) int {
	splited := strings.Split(proposalText, " ")
	for _, spl := range splited {
		if count, err := strconv.Atoi(spl); err == nil {
			return count
		}
	}
	return 0
}

func toMilisecond(timeStr string) int64 {
	splited := strings.Split(timeStr, " ")
	if len(splited) != 3 {
		return time.Now().UnixMilli()
	}
	amount, err := strconv.Atoi(splited[0])
	if err != nil {
		log.Printf("err : %v", err)
	}

	unit := splited[1]
	if unit[len(unit)-1] == 's' {
		unit = unit[:len(unit)-1]
	}

	switch unit {
	case "month":
		return time.Now().Add(-time.Duration(amount) * 30 * 24 * time.Hour).UnixMilli()
	case "day":
		return time.Now().Add(-time.Duration(amount) * 24 * time.Hour).UnixMilli()
	case "hour":
		return time.Now().Add(-time.Duration(amount) * time.Hour).UnixMilli()
	default:
		return time.Now().Add(-time.Duration(amount) * time.Minute).UnixMilli()
	}
}

func (job *Job) AsMessage() string {
	tmplStr := `
<strong>Title:</strong> {{ .Title | html }}
<strong>Price Type:</strong> {{ .PriceType.String | html }}
<strong>Budget:</strong> {{ .Budget | html }}
<strong>Description:</strong> {{ .Description | html }}
<strong>Time Posted:</strong> {{ .FormattedTimePosted | html }}
<strong>Proposal Count:</strong> {{ .ProposalCount | html }}
<strong><a href="{{ .Link | html }}">Link</a></strong> 
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
