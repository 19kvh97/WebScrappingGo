package models

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultInit(t *testing.T) {
	config := Config{}
	log.Printf("config %v", config)
}

func TestEqualation(t *testing.T) {
	config1 := Config{
		Id:   "1",
		Mode: SYNC_BEST_MATCH,
		Account: UpworkAccount{
			Email:    "test@email.com",
			Password: "test",
		},
		State:    NEW_STATE,
		Interval: 3000,
	}

	config2 := Config{
		Id:   "1",
		Mode: SYNC_BEST_MATCH,
		Account: UpworkAccount{
			Email:    "test@email.com",
			Password: "test",
		},
		State:    NEW_STATE,
		Interval: 3000,
	}

	require.Equal(t, config1, config2)
}
