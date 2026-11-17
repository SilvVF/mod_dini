package modpreview

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ExtractModClues extracts information that may help identify the base model
// used by a retexture mod.
func ExtractModClues(modDir string, sections []IniSection) []string {
	var clues []string

	// 1. Read top-level .txt readme files.
	entries, err := os.ReadDir(modDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(
				strings.ToLower(entry.Name()),
				".txt",
			) {
				continue
			}

			path := filepath.Join(modDir, entry.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			lines := strings.Split(string(content), "\n")
			lineCount := 0

			for _, line := range lines {
				line = strings.TrimSpace(line)

				if line == "" ||
					strings.HasPrefix(line, "-") ||
					strings.HasPrefix(line, "=") {
					continue
				}

				if lineCount >= 4 {
					break
				}

				lineCount++

				if len(line) > 3 && len(line) < 120 {
					clues = append(
						clues,
						fmt.Sprintf("%s: %s", entry.Name(), line),
					)
				}
			}
		}
	}

	// 2. Check INI CreditInfo or Author data.
	for _, section := range sections {
		for _, entry := range section.Entries {
			if entry.Key != "data" {
				continue
			}

			if !strings.Contains(entry.Val, "Created by") &&
				!strings.Contains(entry.Val, "Credit") &&
				!strings.Contains(entry.Val, "Author") &&
				!strings.Contains(entry.Val, "VoidHero") {
				continue
			}

			clean := strings.TrimSpace(strings.Trim(entry.Val, `"`))
			if clean == "" {
				continue
			}

			alreadyExists := false
			for _, clue := range clues {
				if strings.Contains(clue, clean) {
					alreadyExists = true
					break
				}
			}

			if !alreadyExists {
				clues = append(clues, "INI Credit: "+clean)
			}
		}
	}

	// 3. Collect referenced mesh buffers.
	referencedBuffers := make([]string, 0)

	for _, section := range sections {
		if !strings.HasPrefix(
			strings.ToLower(section.Name),
			"resource",
		) {
			continue
		}

		for _, entry := range section.Entries {
			if entry.Key != "filename" {
				continue
			}

			if !strings.HasSuffix(entry.Val, ".buf") &&
				!strings.HasSuffix(entry.Val, ".ib") {
				continue
			}

			if !containsString(referencedBuffers, entry.Val) {
				referencedBuffers = append(referencedBuffers, entry.Val)
			}
		}
	}

	if len(referencedBuffers) > 0 {
		clues = append(
			clues,
			"Required Mesh Buffers: "+strings.Join(referencedBuffers, ", "),
		)
	}

	// 4. Character category from the parent directory.
	parent := filepath.Dir(filepath.Clean(modDir))
	parentName := filepath.Base(parent)

	if parentName != "." && parentName != string(filepath.Separator) && parentName != "" {
		clues = append(clues, "Character Category: "+parentName)
	}

	return clues
}

// DiscoverBaseModelCandidates searches sibling mod folders and other
// categories belonging to the same character for matching mesh buffers.
func DiscoverBaseModelCandidates(
	modDir string,
	sections []IniSection,
) ([]BaseModelCandidate, []string) {
	var candidates []BaseModelCandidate
	var requiredBuffers []string

	// Collect requested .buf and .ib filenames.
	for _, section := range sections {
		if !strings.HasPrefix(
			strings.ToLower(section.Name),
			"resource",
		) {
			continue
		}

		for _, entry := range section.Entries {
			if entry.Key != "filename" {
				continue
			}

			if !strings.HasSuffix(entry.Val, ".buf") &&
				!strings.HasSuffix(entry.Val, ".ib") {
				continue
			}

			if !containsString(requiredBuffers, entry.Val) {
				requiredBuffers = append(requiredBuffers, entry.Val)
			}
		}
	}

	parent := filepath.Dir(filepath.Clean(modDir))
	if parent == "." || parent == string(filepath.Separator) {
		return candidates, requiredBuffers
	}

	// 1. Search sibling mod folders.
	entries, err := os.ReadDir(parent)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			siblingPath := filepath.Join(parent, entry.Name())
			if samePath(siblingPath, modDir) {
				continue
			}

			folderName := entry.Name()
			displayName := cleanDisabledPrefix(folderName)

			siblingBuffers := collectBufferNames(siblingPath)
			if len(siblingBuffers) == 0 {
				continue
			}

			candidate := buildBaseModelCandidate(
				folderName,
				displayName,
				siblingBuffers,
				requiredBuffers,
				"",
			)

			candidates = append(candidates, candidate)
		}
	}

	// 2. Search other skin categories belonging to the same character.
	grandparent := filepath.Dir(parent)
	if grandparent == parent || grandparent == "." {
		sortBaseModelCandidates(candidates)
		return candidates, requiredBuffers
	}

	parentFolderName := filepath.Base(parent)
	charPrefix := strings.TrimSpace(
		strings.SplitN(parentFolderName, " - ", 2)[0],
	)

	categoryEntries, err := os.ReadDir(grandparent)
	if err == nil {
		for _, categoryEntry := range categoryEntries {
			if !categoryEntry.IsDir() {
				continue
			}

			categoryPath := filepath.Join(grandparent, categoryEntry.Name())
			if samePath(categoryPath, parent) {
				continue
			}

			categoryName := categoryEntry.Name()

			if !strings.HasPrefix(
				strings.ToLower(categoryName),
				strings.ToLower(charPrefix),
			) {
				continue
			}

			otherModEntries, err := os.ReadDir(categoryPath)
			if err != nil {
				continue
			}

			for _, otherModEntry := range otherModEntries {
				if !otherModEntry.IsDir() {
					continue
				}

				rawModName := otherModEntry.Name()
				otherModPath := filepath.Join(categoryPath, rawModName)

				otherBuffers := collectBufferNames(otherModPath)
				if len(otherBuffers) == 0 {
					continue
				}

				folderName := filepath.Join(categoryName, rawModName)
				displayName := fmt.Sprintf(
					"%s (%s)",
					cleanDisabledPrefix(rawModName),
					categoryName,
				)

				candidate := buildBaseModelCandidate(
					folderName,
					displayName,
					otherBuffers,
					requiredBuffers,
					categoryName,
				)

				candidates = append(candidates, candidate)
			}
		}
	}

	sortBaseModelCandidates(candidates)

	return candidates, requiredBuffers
}

// buildBaseModelCandidate evaluates how many requested buffers exist in
// a candidate mod folder.
func buildBaseModelCandidate(
	folderName string,
	displayName string,
	candidateBuffers []string,
	requiredBuffers []string,
	sourceCategory string,
) BaseModelCandidate {
	matchedCount := 0
	matchedNames := make([]string, 0)

	for _, required := range requiredBuffers {
		for _, candidate := range candidateBuffers {
			if strings.EqualFold(candidate, required) {
				matchedCount++
				matchedNames = append(matchedNames, required)
				break
			}
		}
	}

	totalRequired := len(requiredBuffers)

	confidence := "low"
	var matchDetails string

	if totalRequired > 0 && matchedCount == totalRequired {
		confidence = "exact"

		if sourceCategory != "" {
			matchDetails = fmt.Sprintf(
				"Exact Match from %s: %d/%d required mesh buffers found",
				sourceCategory,
				matchedCount,
				totalRequired,
			)
		} else {
			matchDetails = fmt.Sprintf(
				"Exact Match: %d/%d required mesh buffers found",
				matchedCount,
				totalRequired,
			)
		}
	} else if matchedCount > 0 {
		confidence = "partial"

		if sourceCategory != "" {
			matchDetails = fmt.Sprintf(
				"Partial Match from %s: %d/%d buffers matched (%s)",
				sourceCategory,
				matchedCount,
				totalRequired,
				strings.Join(matchedNames, ", "),
			)
		} else {
			matchDetails = fmt.Sprintf(
				"Partial Match: %d/%d buffers matched (%s)",
				matchedCount,
				totalRequired,
				strings.Join(matchedNames, ", "),
			)
		}
	} else {
		if sourceCategory != "" {
			matchDetails = fmt.Sprintf(
				"From %s: Contains %d 3D mesh buffers",
				sourceCategory,
				len(candidateBuffers),
			)
		} else {
			matchDetails = fmt.Sprintf(
				"Contains %d 3D mesh buffers (unmatched filenames)",
				len(candidateBuffers),
			)
		}
	}

	return BaseModelCandidate{
		FolderName:           folderName,
		DisplayName:          displayName,
		Confidence:           confidence,
		MatchedBuffersCount:  matchedCount,
		TotalRequiredBuffers: totalRequired,
		MatchDetails:         matchDetails,
	}
}

// collectBufferNames collects .buf and .ib files from a directory and
// its immediate subdirectories.
func collectBufferNames(dir string) []string {
	var buffers []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return buffers
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			nestedEntries, err := os.ReadDir(path)
			if err != nil {
				continue
			}

			for _, nested := range nestedEntries {
				if nested.IsDir() {
					continue
				}

				if isMeshBufferFile(nested.Name()) {
					buffers = append(buffers, nested.Name())
				}
			}

			continue
		}

		if isMeshBufferFile(entry.Name()) {
			buffers = append(buffers, entry.Name())
		}
	}

	return buffers
}

func isMeshBufferFile(name string) bool {
	return strings.HasSuffix(name, ".buf") ||
		strings.HasSuffix(name, ".ib")
}

func cleanDisabledPrefix(name string) string {
	name = strings.TrimPrefix(name, "DISABLED ")
	name = strings.TrimPrefix(name, "DISABLED_")
	return name
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}

func samePath(a, b string) bool {
	a, errA := filepath.Abs(a)
	b, errB := filepath.Abs(b)

	if errA != nil || errB != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}

	return strings.EqualFold(
		filepath.Clean(a),
		filepath.Clean(b),
	)
}

func sortBaseModelCandidates(candidates []BaseModelCandidate) {
	confidenceRank := func(confidence string) int {
		switch confidence {
		case "exact":
			return 3
		case "partial":
			return 2
		case "low":
			return 1
		default:
			return 0
		}
	}

	isEnabled := func(name string) bool {
		baseName := filepath.Base(name)
		return !strings.HasPrefix(baseName, "DISABLED ") &&
			!strings.HasPrefix(baseName, "DISABLED_")
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		a := candidates[i]
		b := candidates[j]

		if confidenceRank(a.Confidence) != confidenceRank(b.Confidence) {
			return confidenceRank(a.Confidence) > confidenceRank(b.Confidence)
		}

		if isEnabled(a.FolderName) != isEnabled(b.FolderName) {
			return isEnabled(a.FolderName)
		}

		return a.MatchedBuffersCount > b.MatchedBuffersCount
	})
}

// ArchiveOperations provides the project-specific extraction functions
// needed for ZIP, RAR, and 7z archives.
type ArchiveOperations struct {
	ExtractZIP func(archivePath, destDir string) error

	FindArchiveTool func(winrarPath string) (
		toolPath string,
		toolType string,
		found bool,
	)

	ExtractRAROr7Z func(
		archivePath string,
		destDir string,
		toolPath string,
		toolType string,
	) error
}

// ExtractArchiveGeneric extracts ZIP, RAR, or 7z archives.
//
// ZIP extraction can use the standard library when ExtractZIP is nil.
// RAR and 7z extraction require the configured archive-tool callbacks.
func ExtractArchiveGeneric(
	archivePath string,
	destDir string,
	winrarPath string,
	operations ArchiveOperations,
) error {
	ext := strings.ToLower(filepath.Ext(archivePath))
	ext = strings.TrimPrefix(ext, ".")

	switch ext {
	case "zip":
		if operations.ExtractZIP != nil {
			return operations.ExtractZIP(archivePath, destDir)
		}

		return extractZIP(archivePath, destDir)

	case "rar", "7z":
		if operations.FindArchiveTool == nil ||
			operations.ExtractRAROr7Z == nil {
			return fmt.Errorf(
				"WinRAR or 7-Zip is required to extract .%s files; configure it in Settings",
				ext,
			)
		}

		toolPath, toolType, found := operations.FindArchiveTool(winrarPath)
		if !found {
			return fmt.Errorf(
				"WinRAR or 7-Zip is required to extract .%s files; configure it in Settings",
				ext,
			)
		}

		return operations.ExtractRAROr7Z(
			archivePath,
			destDir,
			toolPath,
			toolType,
		)

	default:
		return fmt.Errorf("unsupported archive format: .%s", ext)
	}
}

// extractZIP extracts a ZIP archive while preventing path traversal.
func extractZIP(archivePath, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	root, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		target := filepath.Join(root, filepath.FromSlash(file.Name))

		targetAbs, err := filepath.Abs(target)
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(root, targetAbs)
		if err != nil {
			return err
		}

		if relative == ".." ||
			strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("invalid ZIP entry path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetAbs, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetAbs), 0755); err != nil {
			return err
		}

		if err := extractZIPFile(file, targetAbs); err != nil {
			return err
		}
	}

	return nil
}

func extractZIPFile(file *zip.File, target string) error {
	source, err := file.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	mode := file.Mode().Perm()
	if mode == 0 {
		mode = 0644
	}

	destination, err := os.OpenFile(
		target,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		mode,
	)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(destination, source)
	closeErr := destination.Close()

	if copyErr != nil {
		return copyErr
	}

	return closeErr
}

// FindModRootInExtracted locates the directory containing the first INI
// file, searching up to five directory levels.
func FindModRootInExtracted(destDir string) string {
	iniFiles := make([]string, 0)
	findINIFiles(destDir, &iniFiles, 0, 5)

	if len(iniFiles) > 0 {
		return filepath.Dir(iniFiles[0])
	}

	entries, err := os.ReadDir(destDir)
	if err == nil {
		var subdirs []string

		for _, entry := range entries {
			if entry.IsDir() {
				subdirs = append(
					subdirs,
					filepath.Join(destDir, entry.Name()),
				)
			}
		}

		if len(subdirs) == 1 {
			return subdirs[0]
		}
	}

	return destDir
}

// findINIFiles recursively finds INI files up to maxDepth.
func findINIFiles(
	dir string,
	results *[]string,
	depth int,
	maxDepth int,
) {
	if depth > maxDepth {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			findINIFiles(path, results, depth+1, maxDepth)
			continue
		}

		if strings.EqualFold(filepath.Ext(entry.Name()), ".ini") {
			*results = append(*results, path)
		}
	}
}
