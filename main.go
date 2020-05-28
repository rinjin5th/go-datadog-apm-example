package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	gintrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type ResultA struct {
	MessageA string `json:"message_a"`
	MessageB string `json:"message_b"`
}

type ResultB struct {
	MessageB string `json:"message_b"`
}

func main() {

	addr := net.JoinHostPort(
		"datadog",
		"8126",
	)
	serviceName := fmt.Sprintf("%s-%s", "ig-test-go", os.Getenv("APP_NAME"))
	middlewareName := fmt.Sprintf("%s-%s", "ig-test-gino", os.Getenv("APP_NAME"))
	tracer.Start(tracer.WithServiceName(serviceName), tracer.WithAgentAddr(addr))
	defer tracer.Stop()

	r := gin.Default()
	r.Use(gintrace.Middleware(middlewareName))
	r.GET("/", func(c *gin.Context) {
		ctx := c.Request.Context()
		encodeSpan, _ := tracer.StartSpanFromContext(ctx, "alpha api call")

		req, _ := http.NewRequest("GET", "http://api-alpha:8080/alpha", nil)
		traceID := strconv.FormatUint(encodeSpan.Context().TraceID(), 10)
		spanID := strconv.FormatUint(encodeSpan.Context().SpanID(), 10)
		req.Header.Set("x-datadog-trace-id", traceID)
		req.Header.Set("x-datadog-parent-id", spanID)
		client := new(http.Client)
		resp, _ := client.Do(req)
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var output ResultA
		err := json.Unmarshal(body, &output)
		if err != nil {
			panic("fxxk")
		}

		encodeSpan.Finish()
		c.String(http.StatusOK, "OK,  A is %s, B is %s", output.MessageA, output.MessageB)
	})
	r.GET("/alpha", func(c *gin.Context) {
		ctx := c.Request.Context()
		encodeSpan, _ := tracer.StartSpanFromContext(ctx, "beta api call")

		traceID := c.GetHeader("x-datadog-trace-id")
		spanID := strconv.FormatUint(encodeSpan.Context().SpanID(), 10)

		req, _ := http.NewRequest("GET", "http://api-beta:8080/beta", nil)
		req.Header.Set("x-datadog-trace-id", traceID)
		req.Header.Set("x-datadog-parent-id", spanID)
		client := new(http.Client)
		resp, _ := client.Do(req)
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var output ResultB
		err := json.Unmarshal(body, &output)
		if err != nil {
			panic("fxxk a")
		}

		encodeSpan.Finish()
		c.JSON(200, gin.H{
			"message_a": "ok",
			"message_b": output.MessageB,
		})
	})
	r.GET("/beta", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message_b": "ok",
		})
	})
	r.Run()
}
