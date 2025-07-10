package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/unicode/norm"
)

func MakeUserId() string {
	return "user" + xid.New().String()
}

func MakePointExchangeId() string {
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

func MakePlanId() string {
	return "pla" + xid.New().String()
}

func MakeOrderPlanId() string {
	return "opl" + xid.New().String()
}

func MakeUserAddressId() string {
	return "add" + xid.New().String()
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func Include(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GenerateVerifyOtp() string {
	code := ""
	for i := 0; i < 6; i++ {
		if i == 0 {
			// Số đầu tiên từ 1-9 (không bao gồm 0)
			n, err := rand.Int(rand.Reader, big.NewInt(9)) // số từ 0-8
			if err != nil {
				log.Println("error generating random number:", err)
				return ""
			}
			code += fmt.Sprintf("%d", n.Int64()+1) // cộng 1 để có số từ 1-9
		} else {
			// Các số tiếp theo từ 0-9
			n, err := rand.Int(rand.Reader, big.NewInt(10)) // số từ 0-9
			if err != nil {
				log.Println("error generating random number:", err)
				return ""
			}
			code += fmt.Sprintf("%d", n.Int64())
		}
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

func SortParams(params url.Values) url.Values {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sortedParams := make(url.Values)
	for _, k := range keys {
		sortedParams[k] = params[k]
	}

	return sortedParams
}

func SendEmail(apiKey, url, to, subject, code, typeEmail string) error {
	bin, err := os.ReadFile(typeEmail)
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

func SendEmailPartner(apiKey, url, to, subject, content string) error {
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

func ToSlug(input string) string {
	// Normalize để tách dấu ra
	t := norm.NFD.String(input)
	slug := strings.Builder{}
	for _, r := range t {
		switch {
		case unicode.Is(unicode.Mn, r):
			continue // bỏ dấu
		case r == 'đ':
			slug.WriteRune('d')
		case r == 'Đ':
			slug.WriteRune('d')
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			slug.WriteRune(unicode.ToLower(r))
		default:
			slug.WriteRune(' ')
		}
	}
	// Thay nhiều dấu cách thành dấu gạch ngang
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(strings.TrimSpace(slug.String()), "-")
}
