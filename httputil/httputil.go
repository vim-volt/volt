package httputil

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
)

// caller must close reader
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

func GetContent(url string) ([]byte, error) {
	r, err := GetContentReader(url)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

func GetContentString(url string) (string, error) {
	b, err := GetContent(url)
	return string(b), err
}
