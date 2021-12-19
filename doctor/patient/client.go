package patient

import (
	"fmt"
	"net/http"
	"time"
)

const timeout = 10 * time.Second

type Client struct {
	targetHost string
	client     *http.Client
}

func NewClient(targetHost string) *Client {
	return &Client{
		targetHost: targetHost,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (cl *Client) Start() error {
	url := cl.buildURL("start")
	return cl.get(url)
}

func (cl *Client) Stop() error {
	url := cl.buildURL("stop")
	return cl.get(url)
}

func (cl *Client) get(url string) error {
	resp, err := cl.client.Get(url)
	if err != nil {
		return err
	}
	return readResp(resp)
}

func readResp(resp *http.Response) error {
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		return nil
	default:
		return fmt.Errorf("failed with %s", resp.Status)
	}
}

func (cl *Client) buildURL(endpoint string) string {
	return fmt.Sprintf("http://%s/api/%s", cl.targetHost, endpoint)
}
