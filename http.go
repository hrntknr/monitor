package main

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var wsListeners = []chan Event{}

func initAPI(config *Config, eventCollector chan Event) error {
	go func() {
		for {
			event := <-eventCollector
			for _, listener := range wsListeners {
				listener <- event
			}
		}
	}()
	go listenAPI(config, eventCollector)
	return nil
}

func listenAPI(config *Config, eventCollector chan Event) {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.LoadHTMLGlob("./templates/*.html")

	r.StaticFile("/build.js", "./static/build.js")

	r.GET("/topology", func(c *gin.Context) {
		c.JSON(http.StatusOK, config.generateTopology())
	})

	r.GET("/ws", func(c *gin.Context) {
		conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		listener := make(chan Event)
		wsListeners = append(wsListeners, listener)

		for {
			event := <-listener
			body, err := json.Marshal(event)
			if err != nil {
				continue
				//TODO: error log
			}
			conn.WriteMessage(websocket.TextMessage, body)
		}
	})

	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	err := r.Run()
	if err != nil {
		panic(err)
	}
}
