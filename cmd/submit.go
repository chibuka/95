package cmd

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/client"
	"github.com/chibuka/95-cli/internal/config"
)

var (
	passStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	failStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	dimStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type streamEvent struct {
	StageNumber int    `json:"stage_number"`
	Passed      bool   `json:"passed"`
	Logs        string `json:"logs"`
	Done        bool   `json:"done"`
}

func submitToServer(stageUuid string) error {
	endpoint := fmt.Sprintf("/api/stages/%s/submit", stageUuid)
	successHint := "→ Check your browser for live progress updates and stage completion!"
	return sendAndStream(stageUuid, endpoint, "submission", successHint)
}

func testOnServer(stageUuid string) error {
	endpoint := fmt.Sprintf("/api/stages/%s/test", stageUuid)
	successHint := "→ Run '95 run' to submit and save your progress."
	return sendAndStream(stageUuid, endpoint, "test", successHint)
}

// sendAndStream handles the shared logic for submit and test:
// loads config, packages the archive, posts to the given endpoint path,
// refreshes the token on 401, streams SSE results, and prints a summary.
func sendAndStream(stageUuid, endpointPath, opName, successHint string) error {
	projectCfg, err := config.LoadProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}
	if projectCfg.RunCommand == "" {
		return fmt.Errorf("no run command found — run '95 init' first")
	}

	globalCfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	if globalCfg.AccessToken == "" {
		return fmt.Errorf("not logged in. Run '95 login' first")
	}

	fmt.Println("Packaging submission...")
	archive, err := createArchive(".")
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	apiURL := globalCfg.GetAPIURL()
	endpoint := apiURL + endpointPath

	resp, err := postSubmission(endpoint, globalCfg.AccessToken, archive, projectCfg.RunCommand, projectCfg.Language)
	if err != nil {
		return fmt.Errorf("%s failed: %w", opName, err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		fmt.Println("Access token expired. Refreshing...")
		authResp, err := client.RefreshToken(globalCfg.RefreshToken, apiURL)
		if err != nil {
			return fmt.Errorf("token refresh failed: %w\n\n→ Run '95 login' to re-authenticate", err)
		}
		globalCfg.AccessToken = authResp.AccessToken
		globalCfg.RefreshToken = authResp.RefreshToken
		if err := globalCfg.Save(); err != nil {
			return fmt.Errorf("failed to save refreshed tokens: %w", err)
		}
		fmt.Println("✓ Token refreshed successfully!")

		resp, err = postSubmission(endpoint, globalCfg.AccessToken, archive, projectCfg.RunCommand, projectCfg.Language)
		if err != nil {
			return fmt.Errorf("%s failed: %w", opName, err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		switch resp.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("stage '%s' not found\n\n→ Check the UUID and try again", stageUuid)
		case http.StatusForbidden:
			return fmt.Errorf("access denied - you don't have permission to access this stage")
		default:
			return fmt.Errorf("%s failed: HTTP %d - %s", opName, resp.StatusCode, strings.TrimSpace(string(body)))
		}
	}

	fmt.Println()
	allPassed := true
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1 MB buffer for large log payloads
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")

		var event streamEvent
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			continue
		}
		if event.Done {
			break
		}

		printStageResult(event)
		if !event.Passed {
			allPassed = false
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading response stream: %w", err)
	}

	fmt.Println()
	if allPassed {
		fmt.Println(passStyle.Render("✓ All stages passed!"))
		fmt.Println()
		fmt.Println(dimStyle.Render(successHint))
	} else {
		fmt.Println(failStyle.Render("✗ Some stages failed."))
	}
	fmt.Println()

	return nil
}

func printStageResult(event streamEvent) {
	var icon string
	if event.Passed {
		icon = passStyle.Render("✓")
	} else {
		icon = failStyle.Render("✗")
	}
	fmt.Printf("Stage %02d  %s\n", event.StageNumber+1, icon)
	if !event.Passed && event.Logs != "" {
		for _, line := range strings.Split(strings.TrimRight(event.Logs, "\n"), "\n") {
			if line != "" {
				fmt.Println(dimStyle.Render("   " + line))
			}
		}
		fmt.Println()
	}
}

func postSubmission(endpoint, token string, archive []byte, runCmd, language string) (*http.Response, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	if err := mw.WriteField("run_command", runCmd); err != nil {
		return nil, err
	}
	if err := mw.WriteField("language", language); err != nil {
		return nil, err
	}
	fw, err := mw.CreateFormFile("code", "submission.tar.gz")
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(archive); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	return http.DefaultClient.Do(req)
}

func createArchive(dir string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	baseDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories (e.g. .git, .DS_Store)
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = rel
		if info.IsDir() {
			header.Name += "/"
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(tw, f)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
