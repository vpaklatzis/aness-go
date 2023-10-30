package api

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vpaklatzis/conduit-go/models"
)

var k models.KeywordRequest

func (s *Server) KeywordHandler(c *gin.Context) {
	if err := c.ShouldBindJSON(&k); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if k.Keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide a valid keyword"})
		return
	}
	client := &http.Client{}
	requestUri := s.conf.NistBaseUrl + s.conf.KeywordSearchPath + k.Keyword

	r, _ := http.NewRequest(http.MethodGet, requestUri, nil)
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("apiKey", s.conf.ApiKey)

	res, err := client.Do(r)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			s.log.Fatal("Failed to close the response body")
		}
	}(res.Body)
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json", respBody)
}
