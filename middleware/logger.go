package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/cristiangraz/kumi"
	"github.com/mssola/user_agent"
	"github.com/whitedevops/colors"
)

type (
	key int
)

const (
	startTimeKey key = iota
)

// Logger registers the logger.
func Logger(c *kumi.Context) {
	start := time.Now()
	c.Next()

	path := c.Request.URL.String()
	status := c.Status()

	log.Printf("%s   %s   %s  %s  %s", logDuration(start), logStatus(status), logMethod(c.Request.Method), logPath(path), logUserAgent(c.Request.UserAgent()))
}

func logDuration(start time.Time) string {
	return fmt.Sprintf("%s%s%13s%s", colors.ResetAll, colors.Dim, time.Since(start), colors.ResetAll)
}

func logStatus(code int) string {
	color := colors.White

	switch {
	case code >= 200 && code <= 299:
		color += colors.BackgroundGreen
	case code >= 300 && code <= 399:
		color += colors.BackgroundCyan
	case code >= 400 && code <= 499:
		color += colors.BackgroundYellow
	default:
		color += colors.BackgroundRed
	}

	return fmt.Sprintf("%s%s %3d %s", colors.ResetAll, color, code, colors.ResetAll)
}

func logMethod(method string) string {
	var color string

	switch method {
	case "GET":
		color += colors.Green
	case "POST":
		color += colors.Cyan
	case "PUT", "PATCH":
		color += colors.Blue
	case "DELETE":
		color += colors.Red
	}

	return fmt.Sprintf("%s%s%s%s", colors.ResetAll, color, method, colors.ResetAll)
}

func logPath(path string) string {
	return fmt.Sprintf("%s%s%s%s", colors.ResetAll, colors.Dim, path, colors.ResetAll)
}

func logUserAgent(agent string) string {
	ua := user_agent.New(agent)

	browser, version := ua.Browser()
	if ua.Bot() {
		return fmt.Sprintf("%s %s%s%s %s%s BOT %s", browser, colors.ResetAll, colors.Dim, version, colors.ResetAll, colors.White+colors.BackgroundRed, colors.ResetAll)
	}

	return fmt.Sprintf("%s %s%s%s%s", browser, colors.ResetAll, colors.Dim, version, colors.ResetAll)
}
