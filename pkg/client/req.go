package client

import (
	"fmt"
	"net/http"
	"os"

	"github.com/imroc/req/v3"
)

// ReqClient ...
type ReqClient struct {
	*req.Client
}

// Get ...
func (rc *ReqClient) Get(url string) (*http.Response, error) {
	r, err := rc.Client.R().Get(url)
	return r.Response, err
}

// GetJSON ...
func (rc *ReqClient) GetJSON(url string, v interface{}) error {
	rc.Client.SetAutoDecodeContentType("json")
	c := os.Getenv("cookies")
	//expiration, _ := time.Parse(time.RFC1123, "Sun, 23-Mar-25 15:03:41 GMT")

	if c != "" {
		cookie := &http.Cookie{
			Name:  "cf_clearance",
			Value: c,
			Path:  "/",
			//Expires:  expiration,
			Domain:   ".konachan.com",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		}

		rc.Client.SetCommonCookies(cookie)
		fmt.Println("test cookies")
		fmt.Println(c)
	}

	_, err := rc.Client.R().SetResult(v).Get(url)
	return err
}

// Post ...
func (rc *ReqClient) Post(url string, v interface{}) (*http.Response, error) {
	r, err := rc.Client.R().SetBody(v).Post(url)
	return r.Response, err
}

// Download ...
func (rc *ReqClient) Download(url, file string, callback req.DownloadCallback) error {
	_, err := rc.Client.R().
		SetOutputFile(file).
		SetDownloadCallback(callback).
		Get(url)
	return err
}

func (rc *ReqClient) CheckDownloadUrl(url string) (bool, error) {
	r, err := rc.Client.R().Head(url)
	if err != nil {
		return false, err
	}
	if r.Response.Header.Get("Content-Type") == "image/jpeg" {
		return true, nil
	}
	return false, nil
}

// SetProxyUrl ...
func (rc *ReqClient) SetProxyUrl(proxyUrl string) error {
	rc.Client = rc.SetProxyURL(proxyUrl)
	rc.Client = rc.SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36")
	return nil
}
