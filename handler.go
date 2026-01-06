package foxdebug

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/tigerwill90/fox"
	"github.com/tigerwill90/foxdebug/internal/humanize"
)

const (
	version = "fox:v0.26.0"
)

// Handler returns a HandlerFunc that responds with detailed system, router and request information. This function may leak
// sensitive information and is only useful for debugging purposes, providing a comprehensive overview of the incoming
// request and the system it is running on.
func Handler() fox.HandlerFunc {
	return func(c *fox.Context) {
		c.SetHeader(fox.HeaderServer, version)
		c.SetHeader(fox.HeaderCacheControl, "max-age=0, must-revalidate, no-cache, no-store, private")
		_ = c.String(http.StatusOK, dumpSysInfo(c, c.Router()))
	}
}

// HandlerWith returns a HandlerFunc that responds with detailed system, router and request information using the
// provided router instance. This function may leak sensitive information and is only useful for debugging purposes,
// providing a comprehensive overview of the incoming request and the system it is running on.
func HandlerWith(f *fox.Router) fox.HandlerFunc {
	return func(c *fox.Context) {
		c.SetHeader(fox.HeaderServer, version)
		c.SetHeader(fox.HeaderCacheControl, "max-age=0, must-revalidate, no-cache, no-store, private")
		_ = c.String(http.StatusOK, dumpSysInfo(c, f))
	}
}

func dumpSysInfo(c *fox.Context, f *fox.Router) string {
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

	stats := f.RouterInfo()
	txn := f.Txn(false)
	defer txn.Abort()

	// Use strings.Builder to build the response
	var builder strings.Builder
	builder.WriteString("Fox: awesome and blazing fast Go router!\n")
	builder.WriteString("Repo: https://github.com/tigerwill90/fox\n\n")
	builder.WriteString("Router Information:\n")
	builder.WriteString("Trailing Slash Option: ")
	builder.WriteString(trailingSlashOption(stats.TrailingSlashOption))
	builder.WriteByte('\n')
	builder.WriteString("Fixed Path Option: ")
	builder.WriteString(fixedPathOption(stats.FixedPathOption))
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
	for route := range it.All() {
		builder.WriteString("- ")
		sbLen := builder.Len()
		for method := range route.Methods() {
			if builder.Len() > sbLen {
				builder.WriteString(",")
			}
			builder.WriteString(method)
		}
		builder.WriteString(" ")
		builder.WriteString(route.Pattern())
		builder.WriteString(" [TSO: ")
		builder.WriteString(trailingSlashOption(route.TrailingSlashOption()))
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
	if c.Route() != nil && c.Route().ClientIPResolver() != nil {
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

func trailingSlashOption(mode fox.TrailingSlashOption) string {
	switch mode {
	case fox.RedirectSlash:
		return "RedirectSlash"
	case fox.RelaxedSlash:
		return "RelaxedSlash"
	default:
		return "StrictSlash"
	}
}

func fixedPathOption(mode fox.FixedPathOption) string {
	switch mode {
	case fox.RedirectPath:
		return "RedirectPath"
	case fox.RelaxedPath:
		return "RelaxedPath"
	default:
		return "StrictPath"
	}
}
