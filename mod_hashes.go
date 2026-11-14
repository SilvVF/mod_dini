package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GitHubFile struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

func getGithubContents(owner, repo, path string) ([]GitHubContent, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/contents/%s",
		owner, repo, path,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var contents []GitHubContent
	err = json.NewDecoder(resp.Body).Decode(&contents)

	return contents, err
}

func getHashes(owner, repo, branch string) (map[string][]byte, error) {
	characterData, err := getGithubContents(owner, repo, branch)
	if err != nil {
		return nil, err
	}

	hashes := make(map[string][]byte)

	for _, characterDir := range characterData {
		if characterDir.Type != "dir" {
			continue
		}

		url := fmt.Sprintf(
			"https://raw.githubusercontent.com/%s/%s/%s/%s/hash.json",
			owner, repo, branch, characterDir.Path,
		)

		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return nil, err
		}

		hashes[characterDir.Name] = data
	}

	return hashes, nil
}
