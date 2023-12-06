package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

const (
	contentTypeHeaderKey = "Content-Type"
	contentTypeValue     = "application/json"
)

func ReverseProxy(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := readBody(c)
		if err != nil {
			respondWithError(c, http.StatusInternalServerError, err.Error())
			return
		}

		proxyReq, err := createProxyRequest(c, target, body)
		if err != nil {
			respondWithError(c, http.StatusInternalServerError, "Error creating request to target server")
			return
		}

		resp, err := sendProxyRequest(proxyReq)
		if err != nil {
			respondWithError(c, http.StatusBadGateway, err.Error())
			return
		}
		defer resp.Body.Close()

		setContentTypeHeaderIfMissing(c)
		copyResponseHeaders(resp, c)
		writeResponseBody(c, resp)
	}
}

func readBody(c *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func createProxyRequest(c *gin.Context, target string, body []byte) (*http.Request, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	query := c.Request.URL.Query()
	targetURL.RawQuery = query.Encode()

	proxyReq, err := http.NewRequest(c.Request.Method, targetURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	proxyReq.Header = c.Request.Header
	proxyReq.Header.Set(contentTypeHeaderKey, contentTypeValue)
	return proxyReq, nil
}

func sendProxyRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

func respondWithError(c *gin.Context, statusCode int, errMsg string) {
	http.Error(c.Writer, errMsg, statusCode)
}

func setContentTypeHeaderIfMissing(c *gin.Context) {
	if _, ok := c.Writer.Header()[contentTypeHeaderKey]; !ok {
		c.Writer.Header().Set(contentTypeHeaderKey, contentTypeValue)
	}
}

func copyResponseHeaders(resp *http.Response, c *gin.Context) {
	for h, val := range resp.Header {
		c.Writer.Header()[h] = val
	}
}

func writeResponseBody(c *gin.Context, resp *http.Response) {
	bodyContent, err := io.ReadAll(resp.Body)
	if err != nil {
		respondWithError(c, http.StatusBadGateway, "Error reading from target server")
		return
	}
	c.Writer.Write(bodyContent)
}
