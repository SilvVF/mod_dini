package main

type InstallState int

const (
	InstallErr InstallState = iota - 1
	InstallQueued
	InstallDownloading
	InstallExtracting
	InstallReady // downloaded + extracted; deployment is the Enabled flag
)

type Pack struct {
	ID       string        `json:"id"`
	ModID    string        `json:"modId"`
	FileID   int64         `json:"fileId"`
	Name     string        `json:"name"`
	State    InstallState  `json:"state"`
	Enabled  bool          `json:"enabled"`
	Priority int           `json:"priority"`
	Manifest ManifestEntry `json:"manifest"`
}

type ModSaved struct {
	ID              string           `json:"id"`
	Submission      Submission       `json:"submission"`
	SubmissionFiles []SubmissionFile `json:"submissionFiles"`
	Packs           []Pack           `json:"packages"`
}

type DataProgress struct {
	ID        string `json:"id"`
	PackageID string `json:"packageId"`
	State     int    `json:"state"`
	Total     int64  `json:"total"`
	Progress  int64  `json:"progress"`
}

type ManifestEntry struct {
	RelPath string `json:"relPath"` // path relative to game mod root, normalized (lowercase, forward slashes)
	Size    int64  `json:"size"`
	Hash    string `json:"hash"`
}

type Submission struct {
	Provider             string  `db:"provider" json:"provider"`
	SubmissionType       string  `db:"submission_type" json:"submissionType"`
	SubmissionID         string  `db:"submission_id" json:"submissionId"`
	Slug                 string  `db:"slug" json:"slug"`
	Name                 string  `db:"name" json:"name"`
	Author               string  `db:"author" json:"author"`
	Description          string  `db:"description" json:"description"`
	ProfileURL           string  `db:"profile_url" json:"profileUrl"`
	Category             string  `db:"category" json:"category"`
	Hero                 *string `db:"hero" json:"hero"`
	IsAudio              bool    `db:"is_audio" json:"isAudio"`
	IsMap                bool    `db:"is_map" json:"isMap"`
	IsNSFW               bool    `db:"is_nsfw" json:"isNsfw"`
	IsObsolete           bool    `db:"is_obsolete" json:"isObsolete"`
	IsTombstoned         bool    `db:"is_tombstoned" json:"isTombstoned"`
	IsHydrated           bool    `db:"is_hydrated" json:"isHydrated"`
	HasFiles             bool    `db:"has_files" json:"hasFiles"`
	ViewCount            int     `db:"view_count" json:"viewCount"`
	Likes                int64   `db:"likes" json:"likes"`
	RemoteAddedAt        int64   `db:"remote_added_at" json:"remoteAddedAt"`
	RemoteUpdatedAt      int64   `db:"remote_updated_at" json:"remoteUpdatedAt"`
	FilesUpdatedAt       int64   `db:"files_updated_at" json:"filesUpdatedAt"`
	LastSeenSnapshot     *string `db:"last_seen_snapshot" json:"lastSeenSnapshot"`
	Images               string  `db:"images" json:"images"`
	AudioURL             *string `db:"audio_url" json:"audioUrl"`
	Tags                 string  `db:"tags" json:"tags"`
	DevelopmentState     *string `db:"development_state" json:"developmentState"`
	CompletionPercentage *int    `db:"completion_percentage" json:"completionPercentage"`
	AuthorRemoteID       *string `db:"author_remote_id" json:"authorRemoteId"`
	ThumbnailURL         *string `db:"thumbnail_url" json:"thumbnailUrl"`
}

type SubmissionFile struct {
	Provider       string `json:"provider"`
	SubmissionType string `json:"submissionType"`
	SubmissionID   string `json:"submissionId"`
	FileID         int64  `json:"fileId"`
	FileName       string `json:"fileName"`
	Size           int64  `json:"size"`
	Description    string `json:"description"`
	DownloadCount  int64  `json:"downloadCount"`
	AnalysisState  string `json:"analysisState"`
	AnalysisCode   string `json:"analysisCode"`
	AnalysisResult string `json:"analysisResult"`
	ContainsExe    bool   `json:"containsExe"`
	DownloadURL    string `json:"downloadUrl"`
	MD5Checksum    string `json:"md5Checksum"`
	ClamAVResult   string `json:"clamAvResult"`
	AvastAVResult  string `json:"avastAvResult"`
	AddedAt        int64  `json:"addedAt"`
}
