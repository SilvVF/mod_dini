package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

type HuiHuiProvider struct{}

const BaseURL = "https://cloudreve.huihui648.org"

type File struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
	Size int64  `json:"size"`
	Path string `json:"path"`
}

type ListResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`

	Data struct {
		Files []File `json:"files"`
	} `json:"data"`
}

type DownloadURLResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`

	Data struct {
		URLs []struct {
			URL string `json:"url"`
		} `json:"urls"`
	} `json:"data"`
}

func (h *HuiHuiProvider) getDownloadURL(
	ctx context.Context,
	fileURI string,
) (string, error) {
	body, err := json.Marshal(map[string]any{
		"uris": []string{fileURI},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		BaseURL+"/api/v4/file/url",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result DownloadURLResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", fmt.Errorf("cloudreve: %s", result.Msg)
	}

	if len(result.Data.URLs) == 0 {
		return "", fmt.Errorf("cloudreve returned no download URLs")
	}

	return result.Data.URLs[0].URL, nil
}

func (h *HuiHuiProvider) ListFiles(ctx context.Context, shareID string) ([]File, error) {
	uri := fmt.Sprintf("cloudreve://%s@share/", shareID)

	endpoint := BaseURL + "/api/v4/file?uri=" + url.QueryEscape(uri)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ListResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("cloudreve: %s", result.Msg)
	}

	return result.Data.Files, nil
}

func (*HuiHuiProvider) ID() string {
	return "huihui"
}

func (h *HuiHuiProvider) Page(ctx context.Context, submissionID string) (SubmissionPage, error) {
	url := "https://huihui648.org/" + submissionID

	log.Println(url)

	res, err := http.Get(url)
	if err != nil {
		return SubmissionPage{}, nil
	}
	defer res.Body.Close()

	p, err := h.parseSubmissionPage(ctx, res.Body)

	return p, err
}

func (*HuiHuiProvider) List(ctx context.Context, p ListParams) (SubmissionList, error) {

	url := "https://huihui648.org/category/" + p.Cat + "?page=" + strconv.Itoa(p.Page)
	log.Println("huihui url: ", url)

	res, err := http.Get(url)
	if err != nil {
		return SubmissionList{}, nil
	}
	defer res.Body.Close()

	subs := parseSubmissions(res.Body)

	return SubmissionList{
		Submissions: subs,
		HasNext:     len(subs) > 0,
		Next:        p.Page + 1,
		Prev:        p.Page - 1,
	}, nil
}

func (h *HuiHuiProvider) parseSubmissionPage(ctx context.Context, html io.Reader) (SubmissionPage, error) {
	doc, err := goquery.NewDocumentFromReader(html)
	if err != nil {
		return SubmissionPage{}, err
	}

	layout := doc.Find(".article-layout").First()
	if layout.Length() == 0 {
		return SubmissionPage{}, err
	}

	main := layout.Find(".article-main").First()
	if main.Length() == 0 {
		return SubmissionPage{}, err
	}

	title := strings.TrimSpace(main.Find(".article-title h1").First().Text())
	if title == "" {
		return SubmissionPage{}, err
	}

	href := ""
	if canonical, ok := doc.Find(`link[rel="canonical"]`).First().Attr("href"); ok {
		href = canonical
	}

	if href == "" {
		return SubmissionPage{}, err
	}

	parsedURL, err := url.Parse(href)
	if err != nil {
		return SubmissionPage{}, err
	}

	slug := strings.Trim(path.Base(strings.TrimSuffix(parsedURL.Path, "/")), "/")
	if slug == "" || slug == "." {
		return SubmissionPage{}, err
	}

	var added, updated int64

	if published := doc.Find(`meta[property="article:published_time"]`).First().AttrOr("content", ""); published != "" {
		if t, err := time.Parse(time.RFC3339Nano, published); err == nil {
			added = t.Unix()
		}
	}

	if modified := doc.Find(`meta[property="article:modified_time"]`).First().AttrOr("content", ""); modified != "" {
		if t, err := time.Parse(time.RFC3339Nano, modified); err == nil {
			updated = t.Unix()
		}
	}

	if updated == 0 {
		updated = added
	}

	viewsText := strings.TrimSpace(main.Find(".resource-views").First().Text())
	viewsText = strings.TrimSpace(strings.TrimPrefix(viewsText, "阅读"))
	views, _ := strconv.Atoi(viewsText)

	// Extract gallery images.
	var images []string

	main.Find(".gallery-stage a.gallery-full, .gallery-thumbnails img").Each(func(_ int, s *goquery.Selection) {
		src := ""

		if s.Is("img") {
			src = s.AttrOr("src", "")
		} else {
			src = s.AttrOr("href", "")
		}

		src = strings.TrimSpace(src)
		if src == "" {
			return
		}

		for _, existing := range images {
			if existing == src {
				return
			}
		}

		images = append(images, src)
	})

	var cover *string
	if len(images) > 0 {
		cover = &images[0]
	}

	// Extract article description as HTML text.
	description := strings.TrimSpace(main.Find(".article-body .prose").First().Text())

	// Extract category.
	category := ""
	resource := layout.Find(".resource-panel").First()

	resource.Find("dl dt").Each(func(i int, dt *goquery.Selection) {
		if strings.TrimSpace(dt.Text()) == "分类" {
			category = strings.TrimSpace(dt.Next().Text())
		}
	})

	// Extract tags.
	var tags []string

	main.Find(".article-body .article-tags").Last().Find("a").Each(func(_ int, a *goquery.Selection) {
		tag := strings.TrimSpace(a.Text())
		if tag != "" {
			tags = append(tags, tag)
		}
	})

	tagsJSON, _ := json.Marshal(tags)

	// Extract author from tags.
	author := "" //guessAuthor(tags)

	// Extract download URL.
	downloadURL := strings.TrimSpace(resource.Find("a.download-button").First().AttrOr("href", ""))

	// Extract resource version and size.
	var version, size string

	resource.Find("dl dt").Each(func(i int, dt *goquery.Selection) {
		switch strings.TrimSpace(dt.Text()) {
		case "版本":
			version = strings.TrimSpace(dt.Next().Text())
		case "大小":
			size = strings.TrimSpace(dt.Next().Text())
		}
	})

	// Extract file information from the displayed resource size.
	fileSize := parseResourceSize(size)

	// Determine NSFW status.
	isNSFW := false
	for _, tag := range tags {
		if strings.EqualFold(tag, "nsfw") {
			isNSFW = true
			break
		}
	}

	// Determine category URL.
	categoryURL := main.Find(".breadcrumbs a").Last().AttrOr("href", "")
	if categoryURL != "" {
		if u, err := url.Parse(categoryURL); err == nil {
			categoryURL = strings.Trim(u.Path, "/")
			categoryURL = strings.TrimPrefix(categoryURL, "category/")
		}
	}

	submission := Submission{
		Provider:        "huihui",
		SubmissionType:  "mod",
		SubmissionID:    slug,
		Slug:            slug,
		Name:            title,
		Author:          author,
		Description:     description,
		ProfileURL:      href,
		Category:        category,
		Hero:            nil,
		IsNSFW:          isNSFW,
		IsHydrated:      true,
		HasFiles:        downloadURL != "",
		ViewCount:       views,
		RemoteAddedAt:   added,
		RemoteUpdatedAt: updated,
		Images:          images,
		Tags:            string(tagsJSON),
		ThumbnailURL:    cover,
	}

	_ = version
	_ = fileSize
	_ = utf8.RuneCountInString(title)

	subFiles := make([]SubmissionFile, 0)

	if downloadURL != "" {
		shareURL, err := url.Parse(downloadURL)
		if err != nil {
			return SubmissionPage{}, err
		}

		shareID := strings.TrimPrefix(
			strings.TrimSuffix(shareURL.Path, "/"),
			"/s/",
		)

		files, err := h.ListFiles(ctx, shareID)
		if err != nil {
			return SubmissionPage{
				Submission:      submission,
				SubmissionFiles: subFiles,
			}, err
		}

		for _, file := range files {
			dlLink, err := h.getDownloadURL(ctx, file.Path)
			if err != nil {
				continue
			}

			subFiles = append(subFiles, SubmissionFile{
				Provider:       "huihui",
				SubmissionType: "mod",
				SubmissionID:   submission.SubmissionID,
				FileID:         file.ID,
				FileName:       file.Name,
				Size:           file.Size,
				Description:    "",
				DownloadURL:    dlLink,
				AddedAt:        updated,
			})
		}
	}

	return SubmissionPage{
		ResolvedToolId:  nil,
		Submission:      submission,
		SubmissionFiles: subFiles,
	}, err
}

func parseResourceSize(size string) int64 {
	fields := strings.Fields(strings.TrimSpace(size))
	if len(fields) == 0 {
		return 0
	}

	value, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0
	}

	unit := ""
	if len(fields) > 1 {
		unit = strings.ToUpper(fields[1])
	}

	switch unit {
	case "KB":
		return int64(value * 1024)
	case "MB":
		return int64(value * 1024 * 1024)
	case "GB":
		return int64(value * 1024 * 1024 * 1024)
	default:
		return int64(value)
	}
}

func parseSubmissions(html io.Reader) []Submission {
	doc, err := goquery.NewDocumentFromReader(html)
	if err != nil {
		return nil
	}

	var submissions []Submission

	doc.Find(".post-grid .post-card").Each(func(_ int, card *goquery.Selection) {
		link := card.Find(".card-caption a").First()
		if link.Length() == 0 {
			link = card.Find("a").First()
		}

		href, exists := link.Attr("href")
		if !exists || href == "" {
			return
		}

		parsedURL, err := url.Parse(href)
		if err != nil {
			return
		}

		slug := strings.Trim(path.Base(parsedURL.Path), "/")
		if slug == "" || slug == "." {
			return
		}

		name := strings.TrimSpace(card.Find(".card-caption h3").First().Text())
		if name == "" {
			name = strings.TrimSpace(card.Find("img").First().AttrOr("alt", ""))
		}

		img := strings.TrimSpace(card.Find(".cover-frame img").First().AttrOr("src", ""))

		var thumbnail *string
		if img != "" {
			thumbnail = &img
		}

		var images []string
		if img != "" {
			images = append(images, img)
		}

		views := 0
		viewsText := strings.TrimSpace(card.Find(".resource-views").First().Text())
		viewsText = strings.TrimPrefix(viewsText, "阅读")
		viewsText = strings.TrimSpace(viewsText)

		if v, err := strconv.Atoi(viewsText); err == nil {
			views = v
		}
		submissions = append(submissions, Submission{
			Provider:       "huihui",
			SubmissionType: "mod",
			SubmissionID:   slug,
			Slug:           wpSlug(href),
			Name:           name,
			ThumbnailURL:   thumbnail,
			Hero:           nil,
			Images:         images,
			ViewCount:      views,
			HasFiles:       true,
		})
	})

	return submissions
}

func wpSlug(href string) string {
	u, err := url.Parse(href)
	if err != nil {
		return href
	}
	// u.Path is already percent-decoded, e.g. "/绝区零-蕾米埃尔-裸漏/"
	return path.Base(strings.TrimSuffix(u.Path, "/"))
}
