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
	"regexp"
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

func (h *HuiHuiProvider) GetDownloadURL(
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

	p, err := h.ParseSubmissionPage(ctx, res.Body)

	return p, err
}

func (*HuiHuiProvider) List(ctx context.Context, p ListParams) (SubmissionList, error) {

	url := "https://huihui648.org/category/zzz/page/" + strconv.Itoa(p.Page)

	res, err := http.Get(url)
	if err != nil {
		return SubmissionList{}, nil
	}
	defer res.Body.Close()

	subs, next, err := ParseCategoryPage(res.Body)

	return SubmissionList{
		Submissions: subs,
		HasNext:     next != "",
		Next:        p.Page + 1,
		Prev:        p.Page - 1,
	}, nil
}

var sizeRe = regexp.MustCompile(`(?i)([\d.]+)\s*(B|KB|MB|GB|TB)`)
var postIDRe = regexp.MustCompile(`^post-\d+$`)

func (h *HuiHuiProvider) ParseSubmissionPage(ctx context.Context, r io.Reader) (SubmissionPage, error) {
	doc, err := goquery.NewDocumentFromReader(r)

	if err != nil {
		log.Println("failed to create doc")
		return SubmissionPage{}, err
	}

	article := doc.Find("article.article-layout").First()

	// --- ids / category / flags from the article's class list ---
	var id, category string
	for _, c := range strings.Fields(article.AttrOr("class", "")) {
		switch {
		case postIDRe.MatchString(c):
			id = strings.TrimPrefix(c, "post-")
		case strings.HasPrefix(c, "category-"):
			category = strings.TrimPrefix(c, "category-")
		}
	}

	canonical := doc.Find(`link[rel="canonical"]`).AttrOr("href", "")
	name := strings.TrimSpace(article.Find("h1").First().Text())

	// --- tags (real display names) + author guess ---
	var tags, authors []string
	article.Find(`.article-tags a[rel="tag"]`).Each(func(_ int, a *goquery.Selection) {
		label := strings.TrimSpace(a.Text())
		if label == "" {
			return
		}
		tags = append(tags, label)

		href := a.AttrOr("href", "")
		slug := strings.Trim(href[strings.LastIndex(strings.TrimSuffix(href, "/"), "/")+1:], "/")
		lower := strings.ToLower(label)
		// creator tags have plain ASCII slugs; generic Chinese tags are
		// percent-encoded, and nsfw/ui are generic too.
		if strings.Contains(slug, "%") || lower == "nsfw" || lower == "ui" {
			return
		}
		authors = append(authors, label)
	})

	// --- character (this is what Hero should hold) ---
	var hero *string
	if c := strings.TrimSpace(article.Find(`.article-tags a[href*="/character/"]`).First().Text()); c != "" {
		hero = &c
	}

	// --- gallery images ---
	var images []string
	article.Find("a[data-gallery-thumbnail]").Each(func(_ int, a *goquery.Selection) {
		if u := a.AttrOr("data-image-url", ""); u != "" {
			images = append(images, u)
		}
	})
	if len(images) == 0 {
		if u := article.Find("img[data-gallery-image]").AttrOr("src", ""); u != "" {
			images = append(images, u)
		}
	}
	if images == nil {
		images = []string{}
	}

	// og:image is the cover used on the cards
	var thumb *string
	if u := doc.Find(`meta[property="og:image"]`).AttrOr("content", ""); u != "" {
		thumb = &u
	} else if len(images) > 0 {
		thumb = &images[0]
	}

	// --- description: entry-content with <br> turned into newlines ---
	entry := article.Find(".entry-content").First().Clone()
	entry.Find("br").ReplaceWithHtml("\n")
	description := strings.TrimSpace(entry.Text())

	// --- timestamps ---
	added := parseMetaTime(doc, "article:published_time")
	updated := parseMetaTime(doc, "article:modified_time")

	// --- resource panel (version / size / updated / download button) ---
	panel := map[string]string{}
	article.Find(".resource-panel dl > div").Each(func(_ int, d *goquery.Selection) {
		k := strings.TrimSpace(d.Find("dt").Text())
		v := strings.TrimSpace(d.Find("dd").Text())
		if k != "" {
			panel[k] = v
		}
	})
	downloadURL := article.Find(".resource-panel a.download-button").AttrOr("href", "")

	imagesJSON, _ := json.Marshal(images)
	if tags == nil {
		tags = []string{}
	}
	tagsJSON, _ := json.Marshal(tags)

	sub := Submission{
		Provider:        "huihui",
		SubmissionType:  "mod",
		SubmissionID:    id,
		Slug:            canonical,
		Name:            name,
		Author:          strings.Join(authors, ", "),
		Description:     description,
		ProfileURL:      canonical,
		Category:        category,
		Hero:            hero,
		IsNSFW:          containsFold(tags, "nsfw"),
		IsHydrated:      true,
		RemoteAddedAt:   added,
		RemoteUpdatedAt: updated,
		FilesUpdatedAt:  updated,
		Images:          string(imagesJSON),
		Tags:            string(tagsJSON),
		ThumbnailURL:    thumb,
	}

	// "https://cloudreve.huihui648.org/s/knQ4Iq"
	shareID := strings.TrimPrefix(downloadURL, "https://cloudreve.huihui648.org/s/")

	subFiles := []SubmissionFile{}

	files, err := h.ListFiles(ctx, shareID)
	if err != nil {
		log.Println("failed to list files", shareID)
		return SubmissionPage{
			Submission:      sub,
			SubmissionFiles: subFiles,
		}, err
	}

	for _, file := range files {
		dlLink, err := h.GetDownloadURL(ctx, file.Path)
		if err != nil {
			return SubmissionPage{}, err
		}
		subFiles = append(subFiles, SubmissionFile{
			Provider:       "huihui",
			SubmissionType: "mod",
			SubmissionID:   id,
			FileID:         file.ID,
			FileName:       file.Name,
			Size:           file.Size,
			Description:    panel["版本"], // e.g. "v1.0"
			DownloadURL:    dlLink,
			AddedAt:        updated,
		})
	}

	return SubmissionPage{Submission: sub, SubmissionFiles: subFiles}, nil
}

func parseMetaTime(doc *goquery.Document, prop string) int64 {
	v := doc.Find(`meta[property="`+prop+`"]`).AttrOr("content", "")
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t.Unix() // UnixMilli() if the rest of your code uses ms
	}
	return 0
}

func containsFold(ss []string, v string) bool {
	for _, s := range ss {
		if strings.EqualFold(s, v) {
			return true
		}
	}
	return false
}

func ParseCategoryPage(r io.Reader) ([]Submission, string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, "", err
	}

	var chars []string
	doc.Find(".character-card .character-name").Each(func(_ int, s *goquery.Selection) {
		if n := strings.TrimSpace(s.Text()); n != "" {
			chars = append(chars, n)
		}
	})

	cards := doc.Find(".updates-section article.work-card")
	if cards.Length() == 0 {
		cards = doc.Find("article.work-card")
	}

	var subs []Submission
	seen := map[string]bool{}
	cards.Each(func(_ int, s *goquery.Selection) {
		sub, ok := parseCard(s, chars) // <- new arg
		if !ok || seen[sub.SubmissionID] {
			return
		}
		seen[sub.SubmissionID] = true
		subs = append(subs, sub)
	})

	next, _ := doc.Find("nav.pagination a.next").First().Attr("href")
	return subs, next, nil
}

func wpSlug(href string) string {
	u, err := url.Parse(href)
	if err != nil {
		return href
	}
	// u.Path is already percent-decoded, e.g. "/绝区零-蕾米埃尔-裸漏/"
	return path.Base(strings.TrimSuffix(u.Path, "/"))
}

func parseCard(s *goquery.Selection, chars []string) (Submission, bool) {
	var id, category string
	var tags []string
	for _, c := range strings.Fields(s.AttrOr("class", "")) {
		switch {
		case postIDRe.MatchString(c):
			id = strings.TrimPrefix(c, "post-")
		case strings.HasPrefix(c, "category-"):
			category = strings.TrimPrefix(c, "category-")
		case strings.HasPrefix(c, "tag-"):
			t := strings.TrimPrefix(c, "tag-")
			if u, err := url.PathUnescape(t); err == nil {
				t = u
			}
			tags = append(tags, t)
		}
	}
	if id == "" {
		return Submission{}, false
	}

	link := s.Find("h3 a").First()
	name := strings.TrimSpace(link.Text())
	href, _ := link.Attr("href")

	// cover image: src is the medium_large size, srcset has the rest
	var coverURL, thumb *string
	if img := s.Find("a.cover-link img").First(); img.Length() > 0 {
		if src, ok := img.Attr("src"); ok {
			coverURL = &src
		}
		small, _ := pickSrcset(img.AttrOr("srcset", ""))
		if small == "" && coverURL != nil {
			small = *coverURL
		}
		if small != "" {
			thumb = &small
		}
	}

	var added int64
	if dt, ok := s.Find("time").First().Attr("datetime"); ok {
		if t, err := time.Parse(time.RFC3339, dt); err == nil {
			added = t.Unix() // use UnixMilli() if the rest of your code uses ms
		}
	}

	images := []string{}
	if coverURL != nil {
		images = append(images, *coverURL)
	}
	imagesJSON, _ := json.Marshal(images)
	tagsJSON, _ := json.Marshal(tags)

	sub := Submission{
		Provider:        "huihui",
		SubmissionType:  "mod",
		SubmissionID:    id,
		Slug:            wpSlug(href),
		Name:            name,
		ProfileURL:      href,
		Category:        category,
		Hero:            matchCharacter(name, chars), // *string, nil if no match
		ThumbnailURL:    thumb,
		IsNSFW:          contains(tags, "nsfw"),
		RemoteAddedAt:   added,
		RemoteUpdatedAt: added,
		Images:          string(imagesJSON),
		Tags:            string(tagsJSON),
	}
	if a := guessAuthor(tags); a != "" {
		sub.Author = a
	}
	return sub, true
}

// Names that show up in the character grid but aren't real characters, or are
// too generic to match against a title.
var skipChars = map[string]bool{
	"UI": true, "UI界面": true, "界面UI": true, "功能": true, "功能性": true,
	"全角色": true, "武器": true, "滑翔机": true, "漂泊者": true, "开拓者": true,
}

// matchCharacter returns the longest character name contained in the title.
func matchCharacter(title string, chars []string) *string {
	best := ""
	for _, c := range chars {
		if skipChars[c] || utf8.RuneCountInString(c) < 2 {
			continue
		}
		if strings.Contains(title, c) && len(c) > len(best) {
			best = c
		}
	}
	if best == "" {
		return nil
	}
	return &best
}

// guessAuthor: on this site the creator shows up as one of the tag-* classes,
// mixed in with generic ones (nsfw, ui, numeric tags like 16/118). Drop the
// generic ones and take what's left. It's a heuristic, not reliable.
func guessAuthor(tags []string) string {
	generic := map[string]bool{"nsfw": true, "ui": true}
	for _, t := range tags {
		if generic[t] {
			continue
		}
		if _, err := strconv.Atoi(t); err == nil {
			continue
		}
		return t
	}
	return ""
}

// pickSrcset returns the smallest and largest URLs from a srcset string.
func pickSrcset(srcset string) (smallest, largest string) {
	minW, maxW := 1<<30, 0
	for _, part := range strings.Split(srcset, ",") {
		f := strings.Fields(strings.TrimSpace(part))
		if len(f) != 2 || !strings.HasSuffix(f[1], "w") {
			continue
		}
		w, err := strconv.Atoi(strings.TrimSuffix(f[1], "w"))
		if err != nil {
			continue
		}
		if w < minW {
			minW, smallest = w, f[0]
		}
		if w > maxW {
			maxW, largest = w, f[0]
		}
	}
	return
}

func contains(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}
