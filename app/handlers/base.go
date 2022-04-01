package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Xacnio/img-host-go/app/bot"
	"github.com/Xacnio/img-host-go/app/models"
	"github.com/Xacnio/img-host-go/app/utils"
	"github.com/Xacnio/img-host-go/pkg/configs"
	"github.com/Xacnio/img-host-go/platform/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"log"
	"regexp"
	"strings"
)

func ProxyHandler(c *fiber.Ctx) error {
	filename := c.Params("filename")
	if strings.HasPrefix(filename, "upload") {
		return c.Next()
	}

	if strings.Count(filename, ".") > 1 {
		return c.SendStatus(fiber.StatusNotFound)
	}

	rd := database.NewRConnection()
	defer rd.RClose()
	data, err := rd.RGet(strings.Split(filename, ".")[0])
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if data == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}
	dataS := models.Image{}
	err2 := json.Unmarshal([]byte(data), &dataS)
	if err2 != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", configs.Get("TG_BOT_TOKEN"), dataS.FilePath)
	if err := proxy.Do(c, url); err != nil {
		return err
	}

	// Telegram file paths is expiring after a while, if not found, renew file path with fileID
	if string(c.Response().Header.ContentType()) == "application/json" {
		foundAgain := false
		result := new(struct {
			Ok bool `json:"ok"`
			Error uint16 `json:"error_code"`
			Description string `json:"description"`
		})
		if err := json.Unmarshal(c.Response().Body(), result); err == nil {
			if !result.Ok && result.Error == 404 {
				file, err2 := bot.Bot.FileByID(dataS.FileID)
				if err2 == nil {
					foundAgain = true
					dataS.FilePath = file.FilePath
					jsonData, _ := json.Marshal(dataS)
					e := rd.RSet(strings.Split(filename, ".")[0], string(jsonData))
					if e == nil {
						url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", configs.Get("TG_BOT_TOKEN"), file.FilePath)
						if err := proxy.Do(c, url); err != nil {
							return err
						}
					}
				}
			}
		}
		if !foundAgain {
			return c.SendStatus(fiber.StatusNotFound)
		}
	}
	// Remove Server headers from response
	c.Response().Header.Del(fiber.HeaderServer)
	switch dataS.FileType {
	case "png":
		c.Response().Header.Del(fiber.HeaderContentDisposition)
		c.Response().Header.Set("Content-type", "image/png")
		c.Set("Content-type", "image/png")
	case "jpg":
		c.Response().Header.Del(fiber.HeaderContentDisposition)
		c.Response().Header.Set("Content-type", "image/jpeg")
		c.Set("Content-type", "image/jpeg")
	case "gif":
		c.Response().Header.Del(fiber.HeaderContentDisposition)
		c.Response().Header.Set("Content-type", "image/gif")
		c.Set("Content-type", "image/gif")
	case "webp":
		c.Response().Header.Del(fiber.HeaderContentDisposition)
		c.Response().Header.Set("Content-type", "image/webp")
		c.Set("Content-type", "image/webp")
	}
	return nil
}

func ImagesViewHandler(c *fiber.Ctx) error {
	cFilenames := c.Locals("Images")
	if cFilenames == nil {
		return c.Redirect("/", fiber.StatusUnauthorized)
	}
	cFilenames1 := cFilenames.(string)
	reg, err := regexp.Compile("[^a-zA-Z0-9,.]+")
	if err != nil {
		log.Fatal(err)
	}
	cFilenames1 = reg.ReplaceAllString(cFilenames1, "")
	filenames := strings.Split(cFilenames1, ",")
	var images []models.ImageView
	for _, filename := range filenames {
		onlyName := strings.Split(filename, ".")[0]
		images = append(images, models.ImageView{DirectURL: fmt.Sprintf("%s/%s", configs.Get("FIBER_IMG_URL"), filename), MainURL: fmt.Sprintf("%s/%s", configs.Get("FIBER_URL"), onlyName), Filename: filename})
	}
	return c.Render("index", fiber.Map{
		"Images": images,
		"ImageView": true,
		"UrlList": utils.UrlListGenerator(images),
		"Title": fmt.Sprintf("Upload Successful - Image Upload", images[0].Filename),
	})
}

func ImageViewHandler(c *fiber.Ctx) error {
	filename := c.Params("filename")
	if filename == "upload" {
		return c.Render("index", fiber.Map{})
	}
	if strings.Count(filename, ".") > 1 {
		return c.SendStatus(fiber.StatusNotFound)
	} else if strings.Count(filename, ".") == 1 {
		filename = strings.Split(filename, ".")[0]
	}
	rd := database.NewRConnection()
	defer rd.RClose()
	data, err := rd.RGet(filename)
	if err != nil {
		return ErrorUpload(c, "An error occurred while fetching the image!")
	}
	if data == "" {
		return ErrorUpload(c, "Image not found!")
	}
	dataS := models.Image{}
	err2 := json.Unmarshal([]byte(data), &dataS)
	if err2 != nil {
		return ErrorUpload(c, "An error occurred while fetching the image!")
	}
	var images []models.ImageView
	images = append(images, models.ImageView{DirectURL: fmt.Sprintf("%s/%s.%s", configs.Get("FIBER_IMG_URL"), filename, dataS.FileType), MainURL: fmt.Sprintf("%s/%s", configs.Get("FIBER_URL"), filename), Filename: fmt.Sprintf("%s.%s", filename, dataS.FileType)})
	return c.Render("index", fiber.Map{
		"Images": images,
		"ImagePreview": images[0].DirectURL,
		"ImageView": true,
		"Title": fmt.Sprintf("%s - Image Upload", images[0].Filename),
		"UrlList": utils.UrlListGenerator(images),
	})
}

func HttpsForwarder(c *fiber.Ctx) error {
	https := c.Get(fiber.HeaderXForwardedProto)
	if https == "http" {
		return c.Redirect(strings.Replace(c.Request().URI().String(), "http://", "https://", 1))
	}
	return c.Next()
}