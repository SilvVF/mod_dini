package mod3d

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mholt/archives"
)

// BaseModelMod represents a mod supplied by the database.
//
// Path must point to either the mod directory or its archive.
type BaseModelMod struct {
	Provider string
	ModID    string
	FileID   string
	Name     string
	Path     string
}

// DiscoverBaseModelCandidates searches the supplied mods for matching
// mesh buffers.
//
// mods should contain all mods belonging to the same category, retrieved
// from the database.
//
// Each mod can point to either an extracted directory or an archive.
//
// FolderName contains the candidate's full local path.
func DiscoverBaseModelCandidates(
	modDir string,
	sections []IniSection,
	mods []BaseModelMod,
) ([]BaseModelCandidate, []string) {
	requiredBuffers := collectRequiredBuffers(sections)

	var candidates []BaseModelCandidate

	if len(requiredBuffers) == 0 {
		return candidates, requiredBuffers
	}

	modDir = filepath.Clean(modDir)

	for _, mod := range mods {
		mod.Path = strings.TrimSpace(mod.Path)

		if mod.Path == "" {
			continue
		}

		// Exclude the currently selected mod.
		if samePath(mod.Path, modDir) {
			continue
		}

		candidateBuffers := collectBufferNames(mod.Path)
		if len(candidateBuffers) == 0 {
			continue
		}

		displayName := strings.TrimSpace(mod.Name)
		if displayName == "" {
			displayName = filepath.Base(mod.Path)
			displayName = strings.TrimSuffix(
				displayName,
				filepath.Ext(displayName),
			)
		}

		displayName = cleanDisabledPrefix(displayName)

		candidates = append(
			candidates,
			buildBaseModelCandidate(
				mod.Path,
				displayName,
				candidateBuffers,
				requiredBuffers,
			),
		)
	}

	sortBaseModelCandidates(candidates)

	return candidates, requiredBuffers
}

// collectRequiredBuffers extracts unique .buf and .ib references from INI sections.
func collectRequiredBuffers(sections []IniSection) []string {
	var buffers []string

	for _, section := range sections {
		if !strings.HasPrefix(strings.ToLower(section.Name), "resource") {
			continue
		}

		for _, entry := range section.Entries {
			if !strings.EqualFold(entry.Key, "filename") ||
				!isMeshBufferFile(entry.Val) {
				continue
			}

			if !containsStringFold(buffers, entry.Val) {
				buffers = append(buffers, entry.Val)
			}
		}
	}

	return buffers
}

// buildBaseModelCandidate evaluates how many requested buffers exist in
// a candidate mod.
func buildBaseModelCandidate(
	folderName string,
	displayName string,
	candidateBuffers []string,
	requiredBuffers []string,
) BaseModelCandidate {
	matchedCount := 0
	matchedNames := make([]string, 0)

	for _, required := range requiredBuffers {
		if containsStringFold(candidateBuffers, required) {
			matchedCount++
			matchedNames = append(matchedNames, required)
		}
	}

	totalRequired := len(requiredBuffers)

	confidence := "low"
	var matchDetails string

	switch {
	case totalRequired > 0 && matchedCount == totalRequired:
		confidence = "exact"

		matchDetails = fmt.Sprintf(
			"Exact Match from %s: %d/%d required mesh buffers found",
			displayName,
			matchedCount,
			totalRequired,
		)

	case matchedCount > 0:
		confidence = "partial"

		matchDetails = fmt.Sprintf(
			"Partial Match from %s: %d/%d buffers matched (%s)",
			displayName,
			matchedCount,
			totalRequired,
			strings.Join(matchedNames, ", "),
		)

	default:
		matchDetails = fmt.Sprintf(
			"From %s: Contains %d 3D mesh buffers",
			displayName,
			len(candidateBuffers),
		)
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

// collectBufferNames collects .buf and .ib files from a directory or archive.
// archives are inspected through archives.FileSystem without extracting them.
func collectBufferNames(source string) []string {
	fsys, err := archives.FileSystem(context.Background(), source, nil)
	if err != nil {
		return nil
	}

	var buffers []string

	err = fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if entry.IsDir() || !isMeshBufferFile(entry.Name()) {
			return nil
		}

		buffer := path.Base(name)

		if !containsStringFold(buffers, buffer) {
			buffers = append(buffers, buffer)
		}

		return nil
	})

	if err != nil {
		return buffers
	}

	return buffers
}

func isMeshBufferFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".buf" || ext == ".ib"
}

func cleanDisabledPrefix(name string) string {
	for _, prefix := range []string{"DISABLED ", "DISABLED_"} {
		if len(name) >= len(prefix) &&
			strings.EqualFold(name[:len(prefix)], prefix) {
			return name[len(prefix):]
		}
	}

	return name
}

func containsStringFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
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
		for _, part := range strings.FieldsFunc(
			filepath.ToSlash(name),
			func(r rune) bool { return r == '/' },
		) {
			upper := strings.ToUpper(part)

			if strings.HasPrefix(upper, "DISABLED ") ||
				strings.HasPrefix(upper, "DISABLED_") {
				return false
			}
		}

		return true
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
		filePath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			findINIFiles(filePath, results, depth+1, maxDepth)
			continue
		}

		if strings.EqualFold(filepath.Ext(entry.Name()), ".ini") {
			*results = append(*results, filePath)
		}
	}
}
