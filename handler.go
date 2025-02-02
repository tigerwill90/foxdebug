package foxdebug

import (
	"fmt"
	"github.com/tigerwill90/fox"
	"github.com/tigerwill90/foxdebug/internal/humanize"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// DebugHandler returns a HandlerFunc that responds with detailed system and request information. Additionally, if a
// "sleep" query parameter is provided with a valid duration, the handler will sleep for the specified duration
// before responding. This function may leak sensitive information and is only useful for debugging purposes, providing
// a comprehensive overview of the incoming request and the system it is running on.
func DebugHandler() fox.HandlerFunc {
	return func(c fox.Context) {
		// Sleep if "sleep" query parameter is provided with a valid duration
		if sleep := c.QueryParam("sleep"); sleep != "" {
			if d, err := time.ParseDuration(sleep); err == nil {
				time.Sleep(d)
			}
		}

		// Send the response
		c.SetHeader(fox.HeaderServer, "fox:v0.22.0")
		c.SetHeader(fox.HeaderCacheControl, "max-age=0, must-revalidate, no-cache, no-store, private")
		_ = c.String(http.StatusOK, dumpSysInfo(c))
	}
}

func dumpSysInfo(c fox.Context) string {
	req := c.Request()

	// Get host information
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Dump the request
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		requestDump = []byte("Failed to dump request")
	}

	f := c.Fox()
	stats := f.Stats()
	txn := f.Txn(false)
	defer txn.Abort()

	// Use strings.Builder to build the response
	var builder strings.Builder
	builder.WriteString("Fox: awesome and blazing fast Go router!\n")
	builder.WriteString("Repo: https://github.com/tigerwill90/fox\n\n")
	builder.WriteString("Router Information:\n")
	builder.WriteString("Redirect Trailing Slash: ")
	builder.WriteString(strconv.FormatBool(stats.RedirectTrailingSlash))
	builder.WriteByte('\n')
	builder.WriteString("Ignore Trailing Slash: ")
	builder.WriteString(strconv.FormatBool(stats.IgnoreTrailingSlash))
	builder.WriteByte('\n')
	builder.WriteString("Auto OPTIONS: ")
	builder.WriteString(strconv.FormatBool(stats.AutoOptions))
	builder.WriteByte('\n')
	builder.WriteString("Handle 405: ")
	builder.WriteString(strconv.FormatBool(stats.MethodNotAllowed))
	builder.WriteByte('\n')
	builder.WriteString("Client IP strategy: ")
	builder.WriteString(strconv.FormatBool(stats.ClientIP))
	builder.WriteByte('\n')
	builder.WriteString("Registered route:\n")
	it := txn.Iter()
	for method, route := range it.All() {
		builder.WriteString("- ")
		builder.WriteString(method)
		builder.WriteString(" ")
		builder.WriteString(route.Pattern())
		builder.WriteString(" [RTS: ")
		builder.WriteString(strconv.FormatBool(route.RedirectTrailingSlashEnabled()))
		builder.WriteString(", ITS: ")
		builder.WriteString(strconv.FormatBool(route.IgnoreTrailingSlashEnabled()))
		builder.WriteString(", CIR: ")
		builder.WriteString(strconv.FormatBool(route.ClientIPResolver() != nil))
		builder.WriteString("]\n")
	}

	builder.WriteString("\n\nHandler Information:\n")
	if ip := c.RemoteIP(); ip != nil {
		builder.WriteString("Remote Address: ")
		builder.WriteString(ip.String())
		builder.WriteByte('\n')
	}
	if c.Route().ClientIPResolver() != nil {
		builder.WriteString("Client IP: ")
		ip, err := c.ClientIP()
		if err != nil {
			builder.WriteString(err.Error())
		} else {
			builder.WriteString(ip.String())
		}
		builder.WriteByte('\n')
	}

	builder.WriteString("Matched Route: ")
	builder.WriteString(c.Pattern())
	builder.WriteByte('\n')
	builder.WriteString("Route Parameters:\n")
	hasParams := false
	for param := range c.Params() {
		builder.WriteString("- ")
		builder.WriteString(param.Key)
		builder.WriteString(": ")
		builder.WriteString(param.Value)
		builder.WriteByte('\n')
		hasParams = true
	}
	if !hasParams {
		builder.WriteString("- None\n")
	}

	builder.WriteString("\n\nFull Request Dump:\n")
	builder.WriteString(string(requestDump))
	builder.WriteString("\nSystem Information:\n")
	builder.WriteString("Time: ")
	builder.WriteString(time.Now().Format(time.RFC3339))
	builder.WriteByte('\n')
	builder.WriteString("Hostname: ")
	builder.WriteString(hostname)
	builder.WriteByte('\n')
	builder.WriteString("OS: ")
	builder.WriteString(runtime.GOOS)
	builder.WriteByte('\n')
	builder.WriteString("Arch: ")
	builder.WriteString(runtime.GOARCH)
	builder.WriteByte('\n')
	builder.WriteString("Go Version: ")
	builder.WriteString(runtime.Version())
	builder.WriteByte('\n')
	builder.WriteString("Pid: ")
	builder.WriteString(strconv.Itoa(os.Getpid()))
	builder.WriteByte('\n')
	builder.WriteString("CPU Cores: ")
	builder.WriteString(fmt.Sprintf("%d", runtime.NumCPU()))
	builder.WriteByte('\n')
	builder.WriteString("Number of Goroutines: ")
	builder.WriteString(fmt.Sprintf("%d", runtime.NumGoroutine()))
	builder.WriteByte('\n')
	builder.WriteString("Allocated Memory: ")
	builder.WriteString(humanize.Bytes(memStats.Alloc))
	builder.WriteByte('\n')
	builder.WriteString("Total Allocated Memory: ")
	builder.WriteString(humanize.Bytes(memStats.TotalAlloc))
	builder.WriteByte('\n')
	builder.WriteString("RSS Memory: ")
	builder.WriteString(humanize.Bytes(memStats.Sys))
	builder.WriteByte('\n')

	return builder.String()
}
