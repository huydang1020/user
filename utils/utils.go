package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

func MakeUserId() string {
	return "user" + xid.New().String()
}

func MakePointExchange() string {
	return "pex" + xid.New().String()
}

func MakeStoreId() string {
	return "sto" + xid.New().String()
}

func MakePartnerId() string {
	return "par" + xid.New().String()
}

func MakePartnerRegistrationId() string {
	return "pre" + xid.New().String()
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GenerateVerifyOtp() string {
	code := ""
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10)) // số từ 0-9
		if err != nil {
			log.Println("error generating random number:", err)
			return ""
		}
		code += fmt.Sprintf("%d", n)
	}
	return code
}

func ConvertUnixToDateTime(format string, t int64) (string, error) {
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Println("load location err:", err)
		return "", err
	}
	formattedDate := time.Unix(t, 0).In(location).Format(format)
	return formattedDate, nil
}

func SendEmail(apiKey, url, to, subject, code string) error {
	bin, err := os.ReadFile("assets/confirm_email.html")
	if err != nil {
		log.Println("read file err:", err)
		return err
	}
	bodyMail := string(bin)
	metric := map[string]string{
		"code": code,
	}
	for k, v := range metric {
		bodyMail = strings.Replace(bodyMail, "{{"+k+"}}", v, -1)
	}
	payload := map[string]interface{}{
		"sender": map[string]string{
			"name":  "Huy Shop",
			"email": "huyhoang2028dv@gmail.com",
		},
		"to": []map[string]string{
			{
				"email": to,
			},
		},
		"subject":     subject,
		"htmlContent": bodyMail,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.Status)
	return nil
}

func SendEmailPartnerRegistration(apiKey, url, to, status, subject, content string) error {
	payload := map[string]interface{}{
		"sender": map[string]string{
			"name":  "Huy Shop",
			"email": "huyhoang2028dv@gmail.com",
		},
		"to": []map[string]string{
			{
				"email": to,
			},
		},
		"subject":     subject,
		"htmlContent": content,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Email Status:", resp.Status)
	return nil
}
