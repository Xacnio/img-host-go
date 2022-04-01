package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"github.com/Xacnio/img-host-go/app/models"
	"github.com/Xacnio/img-host-go/platform/database"
	"github.com/lithammer/shortuuid/v3"
)

func ValidImageContentType(contentType string) bool {
	switch contentType {
	case "image/bmp", "image/jpeg", "image/gif", "image/png", "image/webp", "image/heif", "image/heic":
		return true
	}
	return false
}

func UrlListGenerator(images []models.ImageView) []string {
	var result = make([]string, 6)
	result[0] = ""
	result[1] = ""
	result[2] = ""
	result[3] = ""
	result[4] = ""
	result[5] = ""
	for _, image := range images {
		result[0] += image.DirectURL
		result[1] += fmt.Sprintf("[url=%s][img]%s[/img][/url]", image.MainURL, image.DirectURL)
		result[2] += fmt.Sprintf("[img]%s[/img]", image.DirectURL)
		result[3] += fmt.Sprintf("<img src=\"%s\" alt=\"%s\" />", image.DirectURL, image.Filename)
		result[4] += fmt.Sprintf("<a href=\"%s\"><img src=\"%s\" alt=\"%s\" /></a>", image.MainURL, image.DirectURL, image.Filename)
		result[5] += image.MainURL
		if image != images[len(images)-1] {
			result[0] += "\r\n"
			result[1] += "\r\n"
			result[2] += "\r\n"
			result[3] += "\r\n"
			result[4] += "\r\n"
			result[5] += "\r\n"
		}
	}
	return result
}

var bytes = []byte{25, 11, 44, 77, 66, 51, 54, 22, 31, 55, 88, 99, 74, 12, 00, 22}
const MySecret string = "MDS:MDS^%MW^:%MZS:FSAM:A12123125" // 32 characters
func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}
func Encrypt(text string) (string, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return "", err
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, bytes)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return Encode(cipherText), nil
}
func Decrypt(text string) (string, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return "", err
	}
	cipherText := Decode(text)
	cfb := cipher.NewCFBDecrypter(block, bytes)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}

func ShortUuidGenerate() string {
	rdb := database.NewRConnection()
	defer rdb.RClose()
	for t := 0; t < 20; t++ {
		u := shortuuid.New()[0:7]
		if a, _ := rdb.RGet(fmt.Sprintf("%s", u)); a != "" {
			continue
		}
		return u
	}
	return ""
}