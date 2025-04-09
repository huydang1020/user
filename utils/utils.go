package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

func MakeUserId() string {
	return "user" + xid.New().String()
}

func MakePointExchange() string {
	return "pex" + xid.New().String()
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

func SendEmail(apiKey, url, to, subject, content string) error {
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

func UploadImageToCloudinary(file multipart.File, filePath, cldName, cldKey, cldSecret string) (string, error) {
	cld, err := cloudinary.NewFromParams(cldName, cldKey, cldSecret)
	if err != nil {
		return "", err
	}
	uploadParams := uploader.UploadParams{
		PublicID: filePath,
	}
	ctx := context.Background()
	result, err := cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", err
	}
	imageUrl := result.SecureURL
	return imageUrl, nil
}
