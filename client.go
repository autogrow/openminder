package openminder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Client is an API client for the OpenMinder API
type Client struct {
	*http.Client
	baseURL string
}

// NewClient returns a new OpenMinder API client
func NewClient(baseURL string) *Client {
	return &Client{
		&http.Client{Timeout: time.Minute / 2},
		baseURL,
	}
}

// SetCalibration will set the calibration scale and offset of a specific reading field
func (cl *Client) SetCalibration(field string, scale, offset float64) error {
	url := fmt.Sprintf("%s/readings/calibrate/%s/%0.2f/%0.2f", cl.baseURL, field, scale, offset)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		panic(err)
	}

	res, err := cl.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 204 {
		return fmt.Errorf("unexpected http status: %d", res.StatusCode)
	}

	return nil
}

// Readings returns the readings from the API
func (cl *Client) Readings() (Readings, error) {
	r := Readings{}
	url := cl.baseURL + "/readings"
	res, err := http.Get(url)
	if err != nil {
		return r, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return r, err
	}

	err = json.Unmarshal(data, &r)

	return r, err
}
