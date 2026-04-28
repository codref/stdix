package github

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBase = "https://api.github.com"

// PushFile creates or updates a single file in a GitHub repository via the
// Contents API. token must have contents:write permission on repo.
// repo is in "owner/repo" format. branch defaults to the repo default branch
// when empty.
func PushFile(token, repo, branch, path, message string, content []byte) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	url := fmt.Sprintf("%s/repos/%s/contents/%s", apiBase, repo, path)

	// Fetch existing file SHA (required for updates; absent for new files).
	existingSHA, err := getFileSHA(client, token, url, branch)
	if err != nil {
		return "", fmt.Errorf("fetching existing file metadata: %w", err)
	}

	// Build request body.
	body := map[string]any{
		"message": message,
		"content": base64.StdEncoding.EncodeToString(content),
	}
	if branch != "" {
		body["branch"] = branch
	}
	if existingSHA != "" {
		body["sha"] = existingSHA
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("encoding request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling GitHub API: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("GitHub API returned %s: %s", resp.Status, truncate(respBody, 200))
	}

	// Extract HTML URL of the committed file.
	var result struct {
		Content struct {
			HTMLURL string `json:"html_url"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && result.Content.HTMLURL != "" {
		return result.Content.HTMLURL, nil
	}
	return fmt.Sprintf("https://github.com/%s/blob/%s/%s", repo, branch, path), nil
}

// getFileSHA fetches the current blob SHA of a file. Returns "" if the file
// does not exist yet (new file).
func getFileSHA(client *http.Client, token, url, branch string) (string, error) {
	reqURL := url
	if branch != "" {
		reqURL += "?ref=" + branch
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", nil // new file
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API returned %s: %s", resp.Status, truncate(body, 200))
	}

	var meta struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}
	return meta.SHA, nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
