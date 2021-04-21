package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/json-iterator/go"
)

const DefaultURL string = "http://localhost:8080"

type Client struct {
	HTTPClient *http.Client
	Url        string
	Username   *string
	Password   *string
}

func NewClient(url string, username *string, password *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Url:        DefaultURL,
	}

	c.Url = url

	if (username != nil) && (password != nil) {
		c.Username = username
		c.Password = password
	}

	return &c, nil
}

func (c *Client) request(method, url string, body map[string]interface{}) (interface{}, error) {
	var jsonReader io.Reader

	if body != nil {
		jsonBody, err := jsoniter.Marshal(body)
		if err != nil {
			return nil, err
		}

		jsonReader = bytes.NewReader(jsonBody)

		log.Printf("[DEBUG] Starting request %s %s >> '%s'\n", method, c.Url+url, jsonBody)
	} else {
		log.Printf("[DEBUG] Starting request %s %s\n", method, c.Url+url)
	}

	req, err := http.NewRequest(method, fmt.Sprintf(c.Url+url), jsonReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if (c.Username != nil) && (c.Password != nil) {
		req.SetBasicAuth(
			*c.Username,
			*c.Password,
		)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyResult, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if (res.StatusCode != http.StatusOK) && (res.StatusCode != http.StatusNoContent) {
		return nil, fmt.Errorf("status: %d, method: %s, body: %s", res.StatusCode, method, bodyResult)
	}

	var jsonDecoded interface{}
	if string(bodyResult) != "" {
		decoder := json.NewDecoder(bytes.NewReader(bodyResult))
		decoder.UseNumber()

		err = decoder.Decode(&jsonDecoded)
		if err != nil {
			return nil, err
		}
	}

	return jsonDecoded, nil
}
