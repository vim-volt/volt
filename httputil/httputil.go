package httputil

import (
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
)

// GetContentReader fetches url and returns io.ReadCloser.
// Caller must close the reader.
func GetContentReader(url string) (io.ReadCloser, error) {
	// http.Get() allows up to 10 redirects
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode/100 != 2 {
		return nil, errors.New(url + " returned non-successful status: " + res.Status)
	}
	return res.Body, nil
}

// GetContent fetches url and returns []byte.
func GetContent(url string) ([]byte, error) {
	r, err := GetContentReader(url)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

// GetContentString fetches url and returns string.
func GetContentString(url string) (string, error) {
	b, err := GetContent(url)
	return string(b), err
}
