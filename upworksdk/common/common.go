package common

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
)

func ExtractValidateCookies(cookies []models.Cookie) ([]models.Cookie, error) {
	var validCookies []models.Cookie
	for _, cookie := range cookies {
		if strings.Contains(cookie.Domain, "upwork") && cookie.Secure {
			validCookies = append(validCookies, cookie)
		}
	}
	if len(validCookies) == 0 {
		return nil, errors.New("have no cookie valid")
	}
	log.Printf("cookie leng %d", len(validCookies))
	return validCookies, nil
}

func GenerateUniqueId(config models.Config) (string, error) {
	if config.Account.Email == "" {
		return "", fmt.Errorf("account is invalid, email must not be empty")
	}
	hasher := sha1.New()
	hasher.Write([]byte(fmt.Sprintf("%s_%s_%s", config.Account.Email, config.Mode.GetName(), time.Now().String())))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil)), nil
}
