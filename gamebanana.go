package main

import (
	"context"
	"encoding/json"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type GameBanana struct {
	api *GbApi // replace with the type of your gbApi
}

func NewGameBanana(api *GbApi) *GameBanana { return &GameBanana{api: api} }

func (g *GameBanana) ID() string { return "gamebanana" }

func (g *GameBanana) List(ctx context.Context, p ListParams) (SubmissionList, error) {
	q := NewCategoryQuery()
	q.Page = max(p.Page, 1)
	q.ID = p.Cat
	q.Sort = MostLiked
	q.PerPage = 50

	res, err := g.api.CategoryContent(q)
	if err != nil {
		return SubmissionList{}, err
	}

	subs := make([]Submission, len(res.ARecords))
	for i, record := range res.ARecords {
		subs[i] = MapCategoryRecord(record)
	}

	return NewListPage(subs, q.Page, res.AMetadata.NRecordCount, res.AMetadata.NPerpage), nil
}

func (g *GameBanana) Page(ctx context.Context, submissionID string) (SubmissionPage, error) {

	id, err := strconv.Atoi(submissionID)
	if err != nil {
		return SubmissionPage{}, err
	}

	res, err := g.api.ModPage(id)
	if err != nil {
		return SubmissionPage{}, err
	}

	s, files := MapModPageResponse(res)

	return SubmissionPage{s, files}, nil
}

func extractSlug(profileURL string, id int) string {
	if parsed, err := url.Parse(profileURL); err == nil {
		slug := path.Base(strings.TrimRight(parsed.Path, "/"))
		if slug != "." && slug != "/" && slug != "" {
			return slug
		}
	}

	return strconv.Itoa(id)
}

func thumbnailWithImages(media APreviewMedia) (*string, []string) {
	var thumbnail *string
	images := []string{}

	for _, img := range media.AImages {
		baseURL := strings.TrimRight(img.SBaseURL, "/")

		if img.SFile == "" {
			continue
		}
		imageUrl := baseURL + "/" + strings.TrimLeft(img.SFile, "/")
		images = append(images, imageUrl)
	}

	if len(images) > 0 {
		thumbnail = &images[0]
	}

	return thumbnail, images
}

func MapModPageResponse(record ModPageResponse) (Submission, []SubmissionFile) {
	submissionType := "Mod"
	category := strconv.FormatInt(record.ACategory.IDRow, 10)
	tags, _ := json.Marshal(record.AContentRatings)

	var completionPercentage *int

	authorRemoteId := strconv.FormatInt(record.ASubmitter.IDRow, 10)
	thumbnail, images := thumbnailWithImages(record.APreviewMedia)

	submission := Submission{
		Provider:             "gamebanana",
		SubmissionType:       submissionType,
		SubmissionID:         strconv.Itoa(int(record.IDRow)),
		Slug:                 extractSlug(record.SProfileURL, int(record.IDRow)),
		Name:                 record.SName,
		Author:               record.ASubmitter.SName,
		Description:          record.SText,
		ProfileURL:           record.SProfileURL,
		Category:             category,
		Hero:                 nil,
		IsAudio:              false,
		IsMap:                false,
		IsNSFW:               len(record.AContentRatings) > 0,
		IsObsolete:           record.BIsObsolete,
		IsTombstoned:         record.BIsTrashed || record.BIsWithheld,
		IsHydrated:           true,
		HasFiles:             len(record.AFiles) > 0,
		ViewCount:            int(record.NViewCount),
		Likes:                record.NLikeCount,
		RemoteAddedAt:        record.TsDateAdded,
		RemoteUpdatedAt:      record.TsDateUpdated,
		FilesUpdatedAt:       0,
		Images:               images,
		Tags:                 string(tags),
		CompletionPercentage: completionPercentage,
		AuthorRemoteID:       &authorRemoteId,
		ThumbnailURL:         thumbnail,
	}

	files := make([]SubmissionFile, 0, len(record.AFiles))

	for _, file := range record.AFiles {
		files = append(files, SubmissionFile{
			Provider:       submission.Provider,
			SubmissionType: submission.SubmissionType,
			SubmissionID:   submission.SubmissionID,
			FileID:         strconv.FormatInt(file.IDRow, 10),
			FileName:       file.SFile,
			Size:           file.NFilesize,
			Description:    file.SDescription,
			DownloadCount:  file.NDownloadCount,
			AnalysisState:  file.SAnalysisState,
			AnalysisCode:   file.SAnalysisResultCode,
			AnalysisResult: file.SAnalysisResult,
			ContainsExe:    file.BContainsExe,
			DownloadURL:    file.SDownloadURL,
			MD5Checksum:    file.SMd5Checksum,
			ClamAVResult:   file.SClamAVResult,
			AvastAVResult:  file.SAvastAVResult,
			AddedAt:        file.TsDateAdded,
		})
	}

	return submission, files
}

func MapCategoryRecord(record CategoryRecord) Submission {
	submissionType := record.SModelName
	if submissionType == "" {
		submissionType = "Category"
	}

	category := record.ARootCategory.SName
	if category == "" {
		category = record.SName
	}
	tags, _ := json.Marshal(record.ATags)

	remoteUpdatedAt := record.TsDateUpdated
	if remoteUpdatedAt == 0 {
		remoteUpdatedAt = record.TsDateModified
	}
	authorRemoteId := strconv.FormatInt(record.ASubmitter.IDRow, 10)
	thumbnail, images := thumbnailWithImages(record.APreviewMedia)

	return Submission{
		Provider:        "gamebanana",
		SubmissionType:  submissionType,
		SubmissionID:    strconv.Itoa(record.IDRow),
		Slug:            extractSlug(record.SProfileURL, record.IDRow),
		Name:            record.SName,
		Author:          record.ASubmitter.SName,
		Description:     "",
		ProfileURL:      record.SProfileURL,
		Category:        category,
		Hero:            &record.ASubCategory.SName,
		IsAudio:         false,
		IsMap:           false,
		IsNSFW:          record.BHasContentRatings,
		IsObsolete:      record.BIsObsolete,
		HasFiles:        record.BHasFiles,
		ViewCount:       record.NViewCount,
		Likes:           int64(record.NLikeCount),
		RemoteAddedAt:   int64(record.TsDateAdded),
		RemoteUpdatedAt: int64(remoteUpdatedAt),
		FilesUpdatedAt:  0,
		Images:          images,
		Tags:            string(tags),
		AuthorRemoteID:  &authorRemoteId,
		ThumbnailURL:    thumbnail,
	}
}
