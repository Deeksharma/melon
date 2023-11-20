package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"log"
	"melon/internal/service"
	"net/http"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	err := service.InitializeTransactionLog()
	if err != nil {
		logger.Info("error initializing the transaction logger",
			zap.String("err", err.Error()),
		)
		return
	}
	r := gin.Default() // mux router implements the Handler interface
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello gorilla/mux!",
		})
	})
	r.PUT("/v1/key/:key", keyValuePutHandler)
	r.GET("/v1/key/:key", keyValueGetHandler)
	r.DELETE("/v1/key/:key/", keyValueDeleteHandler)
	log.Fatal(http.ListenAndServeTLS(":8080", "./deeksha-cert.pem", "./deeksha-key.pem", r))
}

// keyValuePutHandler expects to be called with a PUT request for // the "/v1/key/{key}" resource.
func keyValuePutHandler(c *gin.Context) {
	if c.Request.TLS != nil {
		fmt.Println("Certificate used by server:")
		state := c.Request.TLS.ServerName
		fmt.Println("tls server name", state)
	}
	key := c.Param("key")

	value, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	err = service.Put(key, string(value))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	service.WritePut(key, string(value))
	c.JSON(http.StatusCreated, map[string]interface{}{
		"status": "created",
	})
}

func keyValueGetHandler(c *gin.Context) {
	key := c.Param("key")
	value, err := service.Get(key) // Get value for key
	if errors.Is(err, service.ErrorNoSuchKey) {
		c.AbortWithStatusJSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	c.Writer.Write([]byte(value))
}

func keyValueDeleteHandler(c *gin.Context) {
	key := c.Param("key")
	err := service.Delete(key)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	service.WriteDelete(key)
}
