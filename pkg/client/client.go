package client

import (
	"net/http"

	"github.com/imroc/req/v3"
)

// Client ...
type Client interface {
	SetProxyUrl(rawurl string) error
	Get(url string) (*http.Response, error)
	GetJSON(url string, v interface{}) error
	Post(url string, v interface{}) (*http.Response, error)
	Download(url, filename string, callback req.DownloadCallback) error
	CheckDownloadUrl(url string) (bool, error)
}

// New ...
func New() Client {
	return &ReqClient{
		req.C(),
	}
}
