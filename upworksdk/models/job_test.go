package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestJobMessage(t *testing.T) {
	job := Job{
		Title:         "TestTitle",
		PriceType:     HOURLY,
		Budget:        "$30-$50",
		Description:   "Demo description",
		TimePosted:    time.Now().AddDate(0, 0, -2).UnixMilli(),
		ProposalCount: 5,
		Tags:          []string{"demoTages"},
		Client: Client{
			Name:            "empty",
			PaymentVerified: true,
			Stars:           4,
			Spent:           3000,
			Location:        "Vietnam",
		},
	}

	require.Contains(t, job.AsMessage(), time.Now().AddDate(0, 0, -2).Format("2006-01-02 15:04"))
}
