package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func FetchRuntimes() (map[string]string, error) {
	url := fmt.Sprintf("%s/api/runtimes", getAPIURL())
	c := &http.Client{Timeout: 5 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("runtimes: %d", resp.StatusCode)
	}

	var runtimes map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&runtimes); err != nil {
		return nil, err
	}
	return runtimes, nil
}
