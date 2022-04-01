package handlers

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Xacnio/img-host-go/app/bot"
	"github.com/Xacnio/img-host-go/app/models"
	"github.com/Xacnio/img-host-go/app/utils"
	"github.com/Xacnio/img-host-go/pkg/configs"
	"github.com/Xacnio/img-host-go/platform/database"
	"github.com/lithammer/shortuuid/v3"
	tele "gopkg.in/telebot.v3"
	"mime/multipart"
	"path"
	"strconv"
	"strings"
	"time"

	//"github.com/Xacnio/img-host-go/app/bot"
	//"github.com/Xacnio/img-host-go/app/models"
	//"github.com/Xacnio/img-host-go/app/utils"
	//"github.com/Xacnio/img-host-go/pkg/configs"
	//"github.com/Xacnio/img-host-go/platform/database"
	"github.com/gofiber/fiber/v2"
	//"github.com/lithammer/shortuuid/v3"
	//tele "gopkg.in/telebot.v3"
	//"time"
)

func GetIPAddress(c *fiber.Ctx) string {
	IPAddress := c.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = c.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = c.IP()
	}
	return IPAddress
}

func ErrorUpload(c *fiber.Ctx, err string) error {
	tUnix := time.Now().Unix() + 120
	randomKey := shortuuid.New()
	hash, _ := utils.Encrypt(fmt.Sprintf("%s#%s#%s#%d", GetIPAddress(c), configs.HASH1, randomKey, tUnix))
	return c.Render("index", fiber.Map{
		"Error": err,
		"Hash": hash,
		"Key": randomKey,
	})
}

func UploadImage(c *fiber.Ctx) error {
	rdb := database.NewRConnection()
	defer rdb.RClose()
	if form, err := c.MultipartForm(); err == nil {
		validFiles := make([]multipart.FileHeader, 0)
		KK, okKK := form.Value["kk"]
		HH, okHH := form.Value["hh"]
		if !okKK || !okHH {
			return ErrorUpload(c, "[1] Data could not be validated! Try again!")
		}
		decrpyted, _ := utils.Decrypt(strings.Join(HH, ""))
		splitted := strings.Split(decrpyted, "#")
		if len(splitted) == 4 {
			timeExpire, _ := strconv.ParseInt(splitted[3], 10, 64)
			hasher := md5.New()
			hasher.Write([]byte(strings.Join(KK, "")))
			if ky, _ := rdb.RGet(fmt.Sprintf("key-disabled-%s", hex.EncodeToString(hasher.Sum(nil)))); ky == "1" {
				return ErrorUpload(c, "[2] Data could not be validated! Try again!")
			}
			if splitted[0] == GetIPAddress(c) && splitted[2] == strings.Join(KK, "") && time.Now().Unix() < timeExpire {
				files, ok := form.File["images"]
				if ok {
					if len(files) > 8 {
						return ErrorUpload(c, "Max 8 images!")
					}
					for _, file := range files {
						if file.Size > 5 * 1024 * 1024 {
							return ErrorUpload(c, "Max 5 MB per image!")
						}
						f, ee := file.Open()
						if ee != nil {
							continue
						}
						contentType, err := utils.GetFileContentType(f)
						f.Close()
						if err != nil || contentType != file.Header.Get("Content-type") {
							continue
						}
						if utils.ValidImageContentType(file.Header.Get("Content-type")) {
							validFiles = append(validFiles, *file)
						}
					}
					if len(validFiles) == 0 {
						return ErrorUpload(c, "Image(s) could not upload!")
					} else {
						var documents tele.Album
						for _, file := range validFiles {
							u := utils.ShortUuidGenerate()
							if u == "" {
								continue
							}
							fileType := "jpg"
							contentType := file.Header.Get("Content-type")
							if contentType == "image/png" {
								fileType = "png"
							} else if contentType == "image/webp" {
								fileType = "webp"
							} else if contentType == "image/gif" {
								fileType = "gif"
							}
							buffer, err := file.Open()
							if err != nil {
								continue
							}
							reader := bufio.NewReader(buffer)
							fileTypeD := fileType
							if contentType == "image/gif" { // telegram is converting gifs to mp4, this is blocking it
								fileTypeD += ".gif2"
							}
							documents = append(documents, &tele.Document{File: tele.FromReader(reader), Caption: fmt.Sprintf("%s.%s ([Image URL](%s/%s.%s))", u, fileType, configs.Get("FIBER_URL"), u, fileType), FileName: fmt.Sprintf("%s.%s", u, fileTypeD)})
						}
						if len(documents) == 0 {
							return ErrorUpload(c, "Image(s) could not process!")
						} else {
							rdb.RSetTTL(fmt.Sprintf("user-%s", GetIPAddress(c)), fmt.Sprintf("%d", time.Now().Unix()+20), 20)
							hasher := md5.New()
							hasher.Write([]byte(strings.Join(KK, "")))
							rdb.RSetTTL(fmt.Sprintf("key-disabled-%s", hex.EncodeToString(hasher.Sum(nil))), "1", 60 * 60)
							result, err := bot.Bot.SendAlbum(&tele.Chat{ID: configs.GetInt64("TG_BOT_CHANNEL_ID")}, documents, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
							if err != nil {
								return ErrorUpload(c, "Image(s) could not process!")
							}
							var images []string
							if len(result) == 0 {
								return ErrorUpload(c, "Image(s) could not process!")
							} else if len(result) == 1 {
								image := result[0]
								name := strings.Split(image.Caption, " (Image URL)")[0]
								onlyName := strings.Split(name, ".")[0]
								fileType := strings.ReplaceAll(path.Ext(name), ".", "")
								images = append(images, fmt.Sprintf("%s", name))
								filef, errf := bot.Bot.FileByID(image.Document.FileID)
								if errf != nil {
									return ErrorUpload(c, "Image(s) could not process!")
								}
								data := models.Image{
									FileType: fileType,
									UploadDate:   time.Now().Unix(),
									FileID:       image.Document.FileID,
									FilePath:  filef.FilePath,
								}
								jsonData, _ := json.Marshal(data)
								errQ := rdb.RSet(onlyName, string(jsonData))
								if errQ != nil {
									return ErrorUpload(c, "Image(s) could not process!")
								}
								return c.Redirect(fmt.Sprintf("%s/%s", c.Get("FIBER_URL"), onlyName))
							} else {
								for _, image := range result {
									name := strings.Split(image.Caption, " (Image URL)")[0]
									onlyName := strings.Split(name, ".")[0]
									fileType := strings.ReplaceAll(path.Ext(name), ".", "")
									filef, errf := bot.Bot.FileByID(image.Document.FileID)
									if errf != nil {
										continue
									}
									data := models.Image{
										FileType: fileType,
										UploadDate:   time.Now().Unix(),
										FileID:       image.Document.FileID,
										FilePath:  filef.FilePath,
									}
									jsonData, _ := json.Marshal(data)
									errQ := rdb.RSet(onlyName, string(jsonData))
									if errQ != nil {
										continue
									}
									images = append(images, name)
								}
								c.Locals("Images", strings.Join(images, ","))
								return ImagesViewHandler(c)
							}
						}
					}
				}
			}
		}
		return ErrorUpload(c, "Data could not be validated! Try again!")
	}
	return ErrorUpload(c, "Error loading image!")
}