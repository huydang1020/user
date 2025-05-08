package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"

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

func GenerateVerifyCode() string {
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

func SendEmail(apiKey, url, to, subject, code string) error {
	content := fmt.Sprintf(`
		<html>
		<body>
			<h2>Chào mừng bạn đến với Huy Shop!</h2>
			<p>Đây là mã xác nhận của bạn:</p>
			<h3 style="color:rgb(4, 47, 61);">%s</h3>
			<p>Vui lòng nhập mã này để hoàn tất đăng ký tài khoản.</p>
			<p>Nếu bạn không thực hiện yêu cầu này, vui lòng bỏ qua email này.</p>
		</body>
		</html>
	`, code)
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

	fmt.Println("Status:", resp.Status)
	return nil
}
