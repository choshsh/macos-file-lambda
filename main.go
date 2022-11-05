package main

import (
	"bufio"
	"context"
	"fmt"
	"golang.org/x/text/unicode/norm"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/mileusna/useragent"
)

var ginLambda *ginadapter.GinLambdaV2
var userAgent string

func init() {
	r := gin.Default()
	r.POST("/mac/convert", normalize)
	ginLambda = ginadapter.NewV2(r)
}

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	userAgent = req.Headers["user-agent"]
	return ginLambda.ProxyWithContext(ctx, req)
}

func normalize(c *gin.Context) {
	if c.Request.Header["X-Forwarded-For"][0] != "58.124.31.172" {
		if strings.Index(c.Request.Header["Origin"][0], "https://macfile.choshsh.com") != 0 &&
			strings.Index(c.Request.Header["Referer"][0], "https://macfile.choshsh.com") != 0 {
			c.JSON(http.StatusForbidden, gin.H{"msg": "forbidden"})
			return
		}
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err})
		return
	}

	f, _ := file.Open()
	defer f.Close()

	fileExtension := filepath.Ext(file.Filename)
	fileName := strings.TrimSuffix(file.Filename, fileExtension) + "_converted" + fileExtension
	fileNameEncoded := strings.Replace(url.QueryEscape(norm.NFC.String(fileName)), `+`, `%20`, -1)
	reader := bufio.NewReader(f)

	// Set header
	var contentDisposotion string

	switch useragent.Parse(userAgent).Name {
	case "Safari":
		contentDisposotion = `attachment; filename*=utf-8''` + fileNameEncoded
	default:
		contentDisposotion = `attachment; filename="` + fileNameEncoded + `"`
	}

	extraHeaders := map[string]string{"Content-Disposition": contentDisposotion}
	c.DataFromReader(http.StatusOK, file.Size, file.Header.Get("Content-Type"), reader, extraHeaders)

	fmt.Println("#################### Init")
	fmt.Println("original filename:", file.Filename)
	fmt.Println("filename:", norm.NFC.String(file.Filename))
	fmt.Printf("Size (byte): %d\n", file.Size)
	fmt.Printf("Content-Type: %+v\n", file.Header.Get("Content-Type"))
	fmt.Println("#################### End")
}

func main() {
	lambda.Start(Handler)
}
