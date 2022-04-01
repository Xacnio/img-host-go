package main

import (
	"fmt"
	"github.com/Xacnio/img-host-go/app/bot"
	"github.com/Xacnio/img-host-go/app/handlers"
	"github.com/Xacnio/img-host-go/app/utils"
	"github.com/Xacnio/img-host-go/pkg/configs"
	"github.com/Xacnio/img-host-go/platform/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"
	"github.com/lithammer/shortuuid/v3"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func SetupFiber() {
	engine := html.New("./web/views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
			BodyLimit: 41 * 1024 * 1024,
		ProxyHeader: fiber.HeaderXForwardedFor,
	})
	app.Use(logger.New())
	app.Use(handlers.HttpsForwarder)
	app.Static("/", "./web/static")
	app.Get("/:filename", func(c *fiber.Ctx) error {
		if c.Hostname() == strings.Join(strings.Split(configs.Get("FIBER_IMG_URL"), "://")[1:], "") {
			c.Set("cache-control", "no-store, no-cache, must-revalidate")
			return handlers.ProxyHandler(c)
		} else {
			c.Set("cache-control", "no-store, no-cache, must-revalidate")
			return handlers.ImageViewHandler(c)
		}
	})
	app.Use(func (c *fiber.Ctx) error {
		if c.Hostname() == strings.Join(strings.Split(configs.Get("FIBER_IMG_URL"), "://")[1:], "") {
			return c.Redirect(configs.Get("FIBER_URL"))
		} else {
			return c.Next()
		}
	})
	app.Post("/upload", handlers.UploadImage)
	app.Get("/upload-success", handlers.ImagesViewHandler)
	app.Get("/", func(c *fiber.Ctx) error {
		c.Set("cache-control", "no-store, no-cache, must-revalidate")
		tUnix := time.Now().Unix() + 120
		randomKey := shortuuid.New()
		hash, _ := utils.Encrypt(fmt.Sprintf("%s#%s#%s#%d", handlers.GetIPAddress(c), configs.HASH1, randomKey, tUnix))
		return c.Render("index", fiber.Map{"Key": randomKey, "Hash": hash})
	})

	err := app.Listen(fmt.Sprintf("%s:%s", configs.Get("FIBER_HOSTNAME"), configs.Get("FIBER_PORT")))
	if err != nil {
		panic(err)
	}
}

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		os.Exit(1)
	}()
	rd := database.NewRConnection()
	if err := rd.RPing(); err != nil {
		rd.RClose()
		panic(err)
	}
	rd.RClose()
	err := bot.Create()
	if err != nil {
		panic(err)
	}
	SetupFiber()
}
