// Package reproxied is a plugin
package reproxied

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Config the plugin configuration.
type Config struct {
	Proxy      string `json:"proxy"`
	TargetHost string `json:"targetHost"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// reProxied a Traefik plugin.
type reProxied struct {
	next          http.Handler
	client        HTTPClient
	targetHostURL *url.URL
}

// HTTPClient is a reduced interface for http.Client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// New creates a new reProxied plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	proxyURL, err := url.Parse(config.Proxy)
	if err != nil {
		return nil, fmt.Errorf("unable to parse proxy url: %w", err)
	}
	return NewWithClient(ctx, next, config, name, &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}})
}

// NewWithClient creates a new reProxied plugin.
func NewWithClient(ctx context.Context, next http.Handler, config *Config, name string, client HTTPClient) (http.Handler, error) {
	targetHostURL, err := url.Parse(config.TargetHost)
	if err != nil {
		return nil, fmt.Errorf("unable to parse target host url: %w", err)
	}
	return &reProxied{
		next:          next,
		targetHostURL: targetHostURL,
		client:        client,
	}, nil
}

func (c *reProxied) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	proxyRequest := c.createProxyRequest(req)

	resp, err := c.client.Do(proxyRequest)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		_, _ = rw.Write([]byte(err.Error()))
		return
	}

	defer func() { _ = resp.Body.Close() }()

	for key, values := range resp.Header {
		for _, value := range values {
			rw.Header().Add(key, value)
		}
	}
	rw.WriteHeader(resp.StatusCode)
	buf := make([]byte, 1024)
	_, _ = io.CopyBuffer(rw, resp.Body, buf)
}

func (c *reProxied) createProxyRequest(req *http.Request) *http.Request {
	proxyRequest := &http.Request{
		Method: req.Method,
		URL: &url.URL{
			Scheme:      c.targetHostURL.Scheme,
			Opaque:      req.URL.Opaque,
			User:        c.targetHostURL.User,
			Host:        c.targetHostURL.Host,
			Path:        req.URL.Path,
			ForceQuery:  req.URL.ForceQuery,
			RawQuery:    req.URL.RawQuery,
			Fragment:    req.URL.Fragment,
			RawFragment: req.URL.RawFragment,
		},
		Proto:            req.Proto,
		ProtoMajor:       req.ProtoMajor,
		ProtoMinor:       req.ProtoMinor,
		Header:           req.Header,
		Body:             req.Body,
		ContentLength:    req.ContentLength,
		TransferEncoding: req.TransferEncoding,
		Close:            req.Close,
		Host:             c.targetHostURL.Host,
		Form:             req.Form,
		PostForm:         req.PostForm,
		MultipartForm:    req.MultipartForm,
		Trailer:          req.Trailer,
		RemoteAddr:       req.RemoteAddr,
		TLS:              req.TLS,
		Response:         req.Response,
	}
	return proxyRequest
}
