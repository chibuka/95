package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	githubLatestAPI = "https://api.github.com/repos/chibuka/95-cli/releases/latest"
	checkInterval   = 24 * time.Hour
	cacheFileName   = "update-check.json"
)

type cacheEntry struct {
	CheckedAtUnix int64  `json:"checked_at"`
	LatestTag     string `json:"latest_tag"`
}

func cachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".95cli", cacheFileName), nil
}

func readCache() (cacheEntry, error) {
	p, err := cachePath()
	if err != nil {
		return cacheEntry{}, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cacheEntry{}, nil
		}
		return cacheEntry{}, err
	}
	var c cacheEntry
	if err := json.Unmarshal(data, &c); err != nil {
		return cacheEntry{}, err
	}
	return c, nil
}

func writeCache(c cacheEntry) error {
	p, err := cachePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// FetchLatestTag returns the tag_name of the latest GitHub release.
func FetchLatestTag() (string, error) {
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Get(githubLatestAPI)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("github API status %d: %s", resp.StatusCode, string(body))
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	if rel.TagName == "" {
		return "", errors.New("empty tag_name in release response")
	}
	return rel.TagName, nil
}

// MaybeUpdateNotice returns a short user-facing line when the installed
// version is behind the latest GitHub release. Checks the network at most
// once per checkInterval; otherwise uses the cached latest tag.
func MaybeUpdateNotice(current string) string {
	c, err := readCache()
	if err != nil {
		c = cacheEntry{}
	}

	now := time.Now().Unix()
	stale := c.LatestTag == "" || time.Unix(c.CheckedAtUnix, 0).Add(checkInterval).Before(time.Now())

	if stale {
		tag, err := FetchLatestTag()
		if err == nil && tag != "" {
			c.LatestTag = tag
			c.CheckedAtUnix = now
			_ = writeCache(c)
		} else {
			if c.LatestTag == "" {
				return ""
			}
			// Keep last known tag but back off on errors so we do not hammer GitHub.
			c.CheckedAtUnix = now
			_ = writeCache(c)
		}
	}

	if c.LatestTag == "" || c.LatestTag == current {
		return ""
	}
	return fmt.Sprintf("update available: %s → run `95 update`", c.LatestTag)
}

// RecordAppliedVersion updates the local cache after a successful self-update
// so we do not nag until the next release (or until the cache TTL refreshes).
func RecordAppliedVersion(tag string) {
	if tag == "" {
		return
	}
	_ = writeCache(cacheEntry{
		CheckedAtUnix: time.Now().Unix(),
		LatestTag:     tag,
	})
}
