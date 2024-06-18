// Package plugindemo a demo plugin.
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler"
	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
	"github.com/juliens/wasm-goexport/guest"
	_ "github.com/stealthrocket/net/http"
	"github.com/stealthrocket/net/wasip1"
)

func main() {
	// Because there is no file mounted in the plugin by default, we configure insecureSkipVerify to avoid having to load rootCas
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Because there is no file mounted in the plugin by default, we configure a default resolver to 1.1.1.1
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := wasip1.Dialer{
				Timeout: time.Millisecond * time.Duration(3000),
			}

			return d.DialContext(ctx, "udp", "1.1.1.1")
		},
	}

	var config Config
	err := json.Unmarshal(handler.Host.GetConfig(), &config)
	if err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not load config %v", err))
		os.Exit(1)
	}

	mw, err := New(config)
	if err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not load config %v", err))
		os.Exit(1)
	}
	handler.HandleRequestFn = mw.handleRequest
	guest.SetExports(handler.GetExports())
}

// Config the plugin configuration.
type Config struct {
	HeaderName string `json:"headerName,omitempty"`
	Timezone   string `json:"timezone,omitempty"`
}

// Demo a Demo plugin.
type Demo struct {
	header   string
	timezone string
}

type WorldTime struct {
	Abbreviation string    `json:"abbreviation"`
	ClientIp     string    `json:"client_ip"`
	Datetime     time.Time `json:"datetime"`
	DayOfWeek    int       `json:"day_of_week"`
	DayOfYear    int       `json:"day_of_year"`
	Dst          bool      `json:"dst"`
	DstFrom      time.Time `json:"dst_from"`
	DstOffset    int       `json:"dst_offset"`
	DstUntil     time.Time `json:"dst_until"`
	RawOffset    int       `json:"raw_offset"`
	Timezone     string    `json:"timezone"`
	Unixtime     int       `json:"unixtime"`
	UtcDatetime  time.Time `json:"utc_datetime"`
	UtcOffset    string    `json:"utc_offset"`
	WeekNumber   int       `json:"week_number"`
}

// New created a new Demo plugin.
func New(config Config) (*Demo, error) {
	if len(config.HeaderName) == 0 {
		return nil, fmt.Errorf("header name cannot be empty")
	}

	timezone := "Europe/Paris"
	if len(config.Timezone) > 0 {
		timezone = config.Timezone
	}
	return &Demo{
		header:   config.HeaderName,
		timezone: timezone,
	}, nil
}

func (a *Demo) handleRequest(req api.Request, resp api.Response) (next bool, reqCtx uint32) {
	response, err := http.Get("http://worldtimeapi.org/api/timezone/" + a.timezone)
	if err != nil {
		resp.SetStatusCode(http.StatusInternalServerError)
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not fetch time %v", err))
		return false, reqCtx
	}

	if response.StatusCode != 200 {
		resp.SetStatusCode(http.StatusInternalServerError)
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not fetch time %d instead of 200", response.Status))
		return false, reqCtx
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		resp.SetStatusCode(http.StatusInternalServerError)
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not read body %v", err))
		return
	}

	worldTime := WorldTime{}
	err = json.Unmarshal(body, &worldTime)
	if err != nil {
		resp.SetStatusCode(http.StatusInternalServerError)
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not parse body %v", err))
		return
	}
	req.Headers().Set(a.header, worldTime.Datetime.String())

	return true, 0
}
