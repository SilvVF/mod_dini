package modpreview

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

func ParseMod(
	modPath string,
	baseModelFolder *string,
	quality *string,
) (*ViewerPayload, error) {
	return ParseModWithContext(context.Background(), modPath, baseModelFolder, quality)
}

func ParseModWithContext(
	ctx context.Context,
	modPath string,
	baseModelFolder *string,
	quality *string,
) (*ViewerPayload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	modDir := filepath.Clean(modPath)

	info, err := os.Stat(modDir)
	if err != nil || !info.IsDir() {
		return nil, errors.New("mod directory does not exist")
	}

	ownGeometryCount := countGeometryFilesRecursive(modDir, 0, 6)
	isRetexture := ownGeometryCount == 0

	iniFiles := findActiveINIFiles(modDir, 5)
	if len(iniFiles) == 0 {
		return nil, errors.New("no INI files found in mod folder")
	}

	var allSections []IniSection

	for _, iniPath := range iniFiles {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		sections, err := parseINISections(iniPath)
		if err != nil {
			log.Printf("Failed to parse INI %s: %v", iniPath, err)
			continue
		}

		allSections = append(allSections, sections...)
	}

	clues := []string{}
	candidates := []BaseModelCandidate{}

	if isRetexture {
		clues = ExtractModClues(modDir, allSections)
		candidates, _ = DiscoverBaseModelCandidates(modDir, allSections)
	}

	chosenBasePath := resolveBaseModelPath(
		modDir,
		baseModelFolder,
		isRetexture,
		candidates,
	)

	var baseModelSource *string
	if chosenBasePath != "" {
		name := filepath.Base(chosenBasePath)
		baseModelSource = &name
	}

	combinedSections := allSections

	if chosenBasePath != "" {
		baseINI := findActiveINIFiles(chosenBasePath, 5)

		var baseSections []IniSection
		for _, path := range baseINI {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			sections, err := parseINISections(path)
			if err != nil {
				continue
			}

			baseSections = append(baseSections, sections...)
		}

		combinedSections = make([]IniSection, 0, len(baseSections)+len(allSections))
		combinedSections = append(combinedSections, baseSections...)
		combinedSections = append(combinedSections, allSections...)
	}

	toggles := extractINIToggles(combinedSections)

	groups := extractDrawGroups(
		combinedSections,
		modDir,
		chosenBasePath,
	)

	if len(groups) == 0 {
		if isRetexture {
			return &ViewerPayload{
				Meshes:              []MeshData{},
				Textures:            map[string]string{},
				IsRetexture:         true,
				BaseModelSource:     nil,
				Clues:               clues,
				BaseModelCandidates: candidates,
				Toggles:             toggles,
				SuggestedTags:       computeSuggestedTags(true, nil, nil, toggles),
			}, nil
		}

		return nil, errors.New("no renderable draw groups found in INI files")
	}

	isPotato, maxTextureDimension := textureQualitySettings(quality)

	textures, textureFileMap, err := convertGroupTextures(
		ctx,
		groups,
		isPotato,
		maxTextureDimension,
	)
	if err != nil {
		return nil, err
	}

	meshes, err := buildMeshes(
		ctx,
		modDir,
		groups,
		isPotato,
		textureFileMap,
		textures,
	)
	if err != nil {
		return nil, err
	}

	if len(meshes) == 0 {
		if isRetexture {
			return &ViewerPayload{
				Meshes:              []MeshData{},
				Textures:            textures,
				IsRetexture:         true,
				BaseModelSource:     nil,
				Clues:               clues,
				BaseModelCandidates: candidates,
				Toggles:             toggles,
				SuggestedTags:       computeSuggestedTags(true, nil, textures, toggles),
			}, nil
		}

		return nil, errors.New(
			"could not extract any renderable meshes; the mod may have an unsupported layout",
		)
	}

	suggestedTags := computeSuggestedTags(
		isRetexture,
		meshes,
		textures,
		toggles,
	)

	return &ViewerPayload{
		Meshes:              meshes,
		Textures:            textures,
		IsRetexture:         isRetexture,
		BaseModelSource:     baseModelSource,
		Clues:               clues,
		BaseModelCandidates: candidates,
		Toggles:             toggles,
		SuggestedTags:       suggestedTags,
	}, nil
}

func countGeometryFilesRecursive(dir string, depth, maxDepth int) int {
	if depth > maxDepth {
		return 0
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	count := 0

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			count += countGeometryFilesRecursive(path, depth+1, maxDepth)
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".buf" || ext == ".ib" {
			count++
		}
	}

	return count
}

func findActiveINIFiles(root string, maxDepth int) []string {
	var files []string

	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}

		depth := 0
		if rel != "." {
			depth = strings.Count(rel, string(filepath.Separator)) + 1
		}

		if entry.IsDir() {
			if depth > maxDepth {
				return filepath.SkipDir
			}
			if path != root && isDisabledINIPath(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		if strings.EqualFold(filepath.Ext(entry.Name()), ".ini") &&
			!isDisabledINIPath(entry.Name()) {
			files = append(files, path)
		}

		return nil
	})

	sort.Strings(files)
	return files
}

func isDisabledINIPath(name string) bool {
	lower := strings.ToLower(name)

	if strings.HasPrefix(lower, "disabled ") ||
		strings.HasPrefix(lower, "disabled_") ||
		strings.HasPrefix(lower, "disabled-") {
		return true
	}

	if strings.HasSuffix(lower, ".bak") ||
		strings.HasSuffix(lower, ".backup") ||
		strings.HasSuffix(lower, ".old") ||
		strings.HasSuffix(lower, ".orig") {
		return true
	}

	return false
}

func resolveBaseModelPath(
	modDir string,
	requested *string,
	isRetexture bool,
	candidates []BaseModelCandidate,
) string {
	if !isRetexture {
		return ""
	}

	parent := filepath.Dir(modDir)
	grandparent := filepath.Dir(parent)

	if requested != nil {
		folder := strings.TrimSpace(*requested)

		if folder == "" || strings.EqualFold(folder, "none") {
			return ""
		}

		if filepath.IsAbs(folder) {
			if isDirectory(folder) {
				return filepath.Clean(folder)
			}
			return ""
		}

		candidate := filepath.Join(parent, folder)
		if isDirectory(candidate) {
			return filepath.Clean(candidate)
		}

		candidate = filepath.Join(grandparent, folder)
		if isDirectory(candidate) {
			return filepath.Clean(candidate)
		}

		return ""
	}

	if len(candidates) == 0 {
		return ""
	}

	top := candidates[0]
	if !strings.EqualFold(top.Confidence, "exact") &&
		top.MatchedBuffersCount == 0 {
		return ""
	}

	candidate := filepath.Join(parent, top.FolderName)
	if isDirectory(candidate) {
		return filepath.Clean(candidate)
	}

	candidate = filepath.Join(grandparent, top.FolderName)
	if isDirectory(candidate) {
		return filepath.Clean(candidate)
	}

	return ""
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func textureQualitySettings(quality *string) (bool, uint32) {
	if quality == nil {
		return true, 512
	}

	q := strings.ToLower(strings.TrimSpace(*quality))

	switch q {
	case "high", "full":
		return false, 2048
	case "balanced":
		return false, 1024
	default:
		return true, 512
	}
}

func convertGroupTextures(
	ctx context.Context,
	groups []DrawGroup,
	isPotato bool,
	maxTextureDimension uint32,
) (map[string]string, map[string]string, error) {
	uniquePaths := make(map[string]struct{})

	addPath := func(path *string) {
		if path == nil || *path == "" || !isFile(*path) {
			return
		}

		uniquePaths[filepath.Clean(*path)] = struct{}{}
	}

	for _, group := range groups {
		addPath(group.DiffuseFile)

		if !isPotato {
			addPath(group.NormalFile)
			addPath(group.LightFile)
			addPath(group.MaterialFile)
		}
	}

	paths := make([]string, 0, len(uniquePaths))
	for path := range uniquePaths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	type result struct {
		path string
		uri  string
	}

	workerCount := min(runtime.GOMAXPROCS(0), len(paths))
	if workerCount == 0 {
		return map[string]string{}, map[string]string{}, nil
	}

	jobs := make(chan string)
	results := make(chan result, len(paths))

	var workers sync.WaitGroup

	for range workerCount {
		workers.Add(1)

		go func() {
			defer workers.Done()

			for path := range jobs {
				if ctx.Err() != nil {
					continue
				}

				uri := ConvertTextureToDataURI(path, maxTextureDimension)
				if uri == "" {
					continue
				}

				results <- result{
					path: path,
					uri:  uri,
				}
			}
		}()
	}

	go func() {
		defer close(jobs)

		for _, path := range paths {
			if ctx.Err() != nil {
				return
			}
			jobs <- path
		}
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	fileMap := make(map[string]string, len(paths))

	for result := range results {
		fileMap[result.path] = result.uri
	}

	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	return map[string]string{}, fileMap, nil
}

func buildMeshes(
	ctx context.Context,
	modDir string,
	groups []DrawGroup,
	isPotato bool,
	textureFileMap map[string]string,
	textures map[string]string,
) ([]MeshData, error) {
	meshes := make([]MeshData, 0)
	bufferCache := make(map[string][]byte)

	readCached := func(path string) ([]byte, error) {
		path = filepath.Clean(path)

		if data, ok := bufferCache[path]; ok {
			return data, nil
		}

		data, err := ReadBuffer(path)
		if err != nil {
			return nil, err
		}

		bufferCache[path] = data
		return data, nil
	}

	for _, group := range groups {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		posPath := group.PositionFile
		ibPath := group.IBFile

		if !isFile(posPath) || !isFile(ibPath) {
			continue
		}

		posData, err := readCached(posPath)
		if err != nil {
			return nil, fmt.Errorf("read position buffer %q: %w", posPath, err)
		}

		var tcData []byte
		if group.TexcoordFile != nil && isFile(*group.TexcoordFile) {
			tcData, _ = readCached(*group.TexcoordFile)
		}

		ibData, err := readCached(ibPath)
		if err != nil {
			return nil, fmt.Errorf("read index buffer %q: %w", ibPath, err)
		}

		allPositions := ReadPositions(posData, group.PositionStride)

		var allUVs []float32
		if len(tcData) > 0 {
			uvOffset, uvFormat := DetectUVBest(tcData, group.TexcoordStride)
			allUVs = ReadTexcoords(
				tcData,
				group.TexcoordStride,
				uvOffset,
				uvFormat,
			)
		}

		texKey := resolveTextureKey(
			"diffuse",
			group.DiffuseFile,
			modDir,
			textureFileMap,
			textures,
		)

		var normalKey *string
		var lightKey *string
		var materialKey *string

		if !isPotato {
			normalKey = resolveTextureKey(
				"normal",
				group.NormalFile,
				modDir,
				textureFileMap,
				textures,
			)

			lightKey = resolveTextureKey(
				"light",
				group.LightFile,
				modDir,
				textureFileMap,
				textures,
			)

			materialKey = resolveTextureKey(
				"material",
				group.MaterialFile,
				modDir,
				textureFileMap,
				textures,
			)
		}

		for drawIndex, draw := range group.Draws {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			rawIndices := ReadIndices(
				ibData,
				draw.Start,
				draw.Count,
				group.IndexSize,
			)

			if len(rawIndices) == 0 {
				continue
			}

			compacted, ok := CompactVertices(
				allPositions,
				allUVs,
				rawIndices,
				draw.Base,
			)
			if !ok || compacted == nil {
				continue
			}

			if len(compacted.Positions) == 0 || len(compacted.Indices) == 0 {
				continue
			}

			name := group.Name
			if len(group.Draws) > 1 {
				name = fmt.Sprintf("%s #%d", group.Name, drawIndex+1)
			}

			meshes = append(meshes, MeshData{
				Name:        name,
				Positions:   EncodeF32Array(compacted.Positions),
				UVs:         encodeOptionalF32Array(compacted.UVs, compacted.HasUVs),
				Indices:     EncodeU32Array(compacted.Indices),
				TexKey:      texKey,
				NormalKey:   normalKey,
				LightKey:    lightKey,
				MaterialKey: materialKey,
				Component:   stringPointer(group.Name),
				Condition:   draw.Condition,
			})
		}
	}

	return meshes, nil
}

func resolveTextureKey(
	kind string,
	path *string,
	modDir string,
	fileMap map[string]string,
	textures map[string]string,
) *string {
	if path == nil || *path == "" {
		return nil
	}

	cleanPath := filepath.Clean(*path)
	uri, ok := fileMap[cleanPath]
	if !ok {
		return nil
	}

	relPath, err := filepath.Rel(modDir, cleanPath)
	if err != nil || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) || relPath == ".." {
		relPath = filepath.Base(cleanPath)
	}

	relPath = filepath.ToSlash(relPath)
	key := kind + "::" + relPath

	if _, exists := textures[key]; !exists {
		textures[key] = uri
	}

	return &key
}

func encodeOptionalF32Array(values []float32, hasValues bool) *string {
	if !hasValues {
		return nil
	}

	encoded := EncodeF32Array(values)
	return &encoded
}

func stringPointer(value string) *string {
	return &value
}

func isFile(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
