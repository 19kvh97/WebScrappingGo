package models

import (
	"log"
	"testing"
)

func TestDefaultInit(t *testing.T) {
	config := Config{}
	log.Printf("config %v", config)
}
