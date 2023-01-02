package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var store = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

func Put(key string, value string) error {
	store.Lock()
	store.m[key] = value
	store.Unlock()
	return nil
}

var ErrorNoSuchKey = errors.New("no such key") // sentinel error

func Get(key string) (string, error) {
	store.RLock()
	value, ok := store.m[key]
	store.RUnlock()
	if !ok {
		return "", ErrorNoSuchKey
	}
	return value, nil
}

func Delete(key string) error {
	store.Lock()
	delete(store.m, key)
	store.Unlock()
	return nil
}

func helloMuxHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello gorilla/mux!\n"))
}
func main() {
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
	log.Fatal(http.ListenAndServe(":8080", r))
}

// keyValuePutHandler expects to be called with a PUT request for // the "/v1/key/{key}" resource.
func keyValuePutHandler(c *gin.Context) {
	key := c.Param("key")

	value, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	err = Put(key, string(value))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, map[string]interface{}{
		"status": "created",
	})
}

func keyValueGetHandler(c *gin.Context) {
	key := c.Param("key")
	value, err := Get(key) // Get value for key
	if errors.Is(err, ErrorNoSuchKey) {
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
	err := Delete(key)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
}
