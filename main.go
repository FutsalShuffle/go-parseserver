package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"log"
	"net/http"
)

var browser *rod.Browser
var tabs chan (*rod.Page)

func main() {
	tabLimit := 10
	port := 3000

	launcher := launcher.New().Headless(false).MustLaunch()
	browser = rod.New().ControlURL(launcher).MustConnect()

	tabs = make(chan *rod.Page, tabLimit)

	for i := 0; i < tabLimit; i++ {
		tabs <- browser.MustPage()
	}

	defer browser.MustClose()

	r := gin.Default()
	r.GET("/parse", fetchHTML)

	err := r.Run(fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func fetchHTML(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'url' query parameter"})
		return
	}

	html := make(chan string)

	go func() {
		var page *rod.Page
		page = <-tabs
		page.MustNavigate(url)
		page.WaitLoad()

		result, _ := page.HTML()

		html <- result
		tabs <- page
	}()

	c.String(http.StatusOK, <-html)
}
