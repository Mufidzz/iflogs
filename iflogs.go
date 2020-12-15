package iflogs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Engine struct {
	Barrier Barrier
}

type ApiEndpointLog struct {
	Ip     string
	Path   string
	Method string
	Token  string
}
type AuthEndpointLog struct {
	Ip     string
	Client string
}

func GenerateAPILog(c *gin.Context) ApiEndpointLog {
	tok, err := c.Cookie("ifx-at")
	if err != nil {
		c.JSON(http.StatusBadRequest, fmt.Sprintf("Unable to write log, error : %v", err.Error()))
		c.Abort()
	}

	log := ApiEndpointLog{
		Ip:     c.ClientIP(),
		Path:   c.FullPath(),
		Method: c.Request.Method,
		Token:  tok,
	}

	return log
}

func GenerateAuthLog(c *gin.Context) AuthEndpointLog {
	tok := c.GetHeader("IFX-CLIENT")
	log := AuthEndpointLog{
		Ip:     c.ClientIP(),
		Client: tok,
	}

	return log
}

func (engine *Engine) Push(log interface{}) error {
	b, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("json marshall error, with real error : %v", err)
	}

	req, err := http.NewRequest("POST", engine.Barrier.URL, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("request error, with real error : %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("response error, with real error : %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		return fmt.Errorf("not success")
	}

	return nil
}

func (engine *Engine) GinForwardMiddleware(tp string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var log interface{}

		switch tp {
		case "AUTH":
			log = GenerateAuthLog(c)
			break
		case "API":
			log = GenerateAPILog(c)
			break
		}

		err := engine.Push(log)
		if err != nil {
			c.JSON(http.StatusBadRequest, fmt.Sprintf("Unable to write log, error : %v", err.Error()))
			c.Abort()
		}

		c.Next()
	}
}

func (engine *Engine) GinForwardLog(handler func(c *gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		tok, err := c.Cookie("IFX-ACCESS-TOKEN")
		if err != nil {
			c.JSON(http.StatusBadRequest, fmt.Sprintf("Unable to write log, error : %v", err.Error()))
		}

		log := ApiEndpointLog{
			Ip:     c.ClientIP(),
			Path:   c.FullPath(),
			Method: c.Request.Method,
			Token:  tok,
		}

		err = engine.Push(log)
		if err != nil {
			c.JSON(http.StatusBadRequest, fmt.Sprintf("Unable to write log, error : %v", err.Error()))
		}

		handler(c)
	}
}
