package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// requestTimes stores the timestamp of the last request for each IP address.
var requestTimes sync.Map

func rateLimiterMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ip := c.RealIP()
		currentTime := strconv.FormatInt(time.Now().UnixMicro(), 10)
		// 127.0.0.1:TIMESTAMPS:/ is the "key" of the rate limiter
		mapKey := ip + ":" + currentTime + ":" + c.Request().URL.String()

		if _, exists := requestTimes.Load(mapKey); exists {
			// If there is a record for this IP at this second, block the request
			return c.String(http.StatusTooManyRequests, "You have exceeded the rate limit of 1 request per second.")
		}

		// Record the request time for this IP
		requestTimes.Store(mapKey, true)
		//go func() {
		//	<-time.After(time.Second)
		//	requestTimes.Delete(mapKey)
		//}() to emulate the memory leak

		return next(c)
	}
}

func main() {
	e := echo.New()

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	}, rateLimiterMiddleware)

	e.GET("/mem-info", func(c echo.Context) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Format the memory usage data
		memUsage := fmt.Sprintf("Alloc = %v MiB, TotalAlloc = %v MiB, Sys = %v MiB, NumGC = %v",
			bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC)

		return c.String(http.StatusOK, memUsage)
	})

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
