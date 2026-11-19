package modpreview

const (
	DefaultPositionStride = 40
	DefaultPositionOffset = 0
	DefaultUVOffset       = 4

	IndexSizeU32 IndexSize = 4
	IndexSizeU16 IndexSize = 2
)

type IndexSize int

type BaseModelCandidate struct {
	FolderName           string `json:"folderName"`
	DisplayName          string `json:"displayName"`
	Confidence           string `json:"confidence"`
	MatchedBuffersCount  int    `json:"matchedBuffersCount"`
	TotalRequiredBuffers int    `json:"totalRequiredBuffers"`
	MatchDetails         string `json:"matchDetails"`
}

type ViewerToggle struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Key          *string  `json:"key"`
	Variable     string   `json:"variable"`
	CurrentValue int      `json:"currentValue"`
	Values       []int    `json:"values"`
	Labels       []string `json:"labels"`
}

type ViewerPayload struct {
	Meshes              []MeshData           `json:"meshes"`
	Textures            map[string]string    `json:"textures"`
	IsRetexture         bool                 `json:"isRetexture"`
	BaseModelSource     *string              `json:"baseModelSource"`
	BaseModelCandidates []BaseModelCandidate `json:"baseModelCandidates"`
	Toggles             []ViewerToggle       `json:"toggles"`
	SuggestedTags       []string             `json:"suggestedTags"`
	TempInstallPath     *string              `json:"tempInstallPath"`
	ArchiveFileName     *string              `json:"archiveFileName"`
}

type MeshData struct {
	Name        string  `json:"name"`
	Positions   string  `json:"positions"`
	UVs         *string `json:"uvs"`
	Indices     string  `json:"indices"`
	TexKey      *string `json:"texKey"`
	NormalKey   *string `json:"normalKey"`
	LightKey    *string `json:"lightKey"`
	MaterialKey *string `json:"materialKey"`
	Component   *string `json:"component"`
	Condition   *string `json:"condition"`
}

type ResourceRecord struct {
	Name     string
	Filename string
	FullPath string
	Stride   *int
	Format   *string
}

type IniEntry struct {
	Key       string
	Val       string
	Condition *string
}

type IniSection struct {
	Name    string
	Entries []IniEntry
	IniDir  string
}

type DrawGroup struct {
	Name           string
	PositionFile   string
	TexcoordFile   *string
	IBFile         string
	PositionStride int
	TexcoordStride int
	IndexSize      IndexSize
	DiffuseFile    *string
	NormalFile     *string
	LightFile      *string
	MaterialFile   *string
	Draws          []DrawIndexed
}

type DrawIndexed struct {
	Count     int     `json:"count"`
	Start     int     `json:"start"`
	Base      int32   `json:"base"`
	Condition *string `json:"condition"`
}

type UvFormat uint8

const (
	UvFormatFloat16 UvFormat = iota
	UvFormatFloat32
)
