package mod3d

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const maxResourceSearchDepth = 6

// parseINISections parses an INI file and preserves conditional expressions.
// readINIToString should handle UTF-8 and UTF-16 files.
func parseINISections(path string) ([]IniSection, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	iniDir := filepath.Dir(path)

	var sections []IniSection
	var current *IniSection
	var conditions []string

	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))

		if line == "" ||
			strings.HasPrefix(line, ";") ||
			strings.HasPrefix(line, "//") {
			continue
		}

		if index := strings.IndexByte(line, ';'); index >= 0 {
			line = strings.TrimSpace(line[:index])
		}

		if line == "" {
			continue
		}

		lower := strings.ToLower(line)

		switch {
		case strings.HasPrefix(lower, "if "):
			conditions = append(conditions, strings.TrimSpace(line[3:]))
			continue

		case lower == "endif":
			if len(conditions) > 0 {
				conditions = conditions[:len(conditions)-1]
			}
			continue

		case strings.HasPrefix(lower, "else"):
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if current != nil {
				sections = append(sections, *current)
			}

			conditions = nil

			current = &IniSection{
				Name:    strings.TrimSpace(line[1 : len(line)-1]),
				Entries: make([]IniEntry, 0),
				IniDir:  iniDir,
			}
			continue
		}

		if current == nil {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		var condition *string
		if len(conditions) > 0 {
			joined := strings.Join(conditions, " && ")
			condition = &joined
		}

		current.Entries = append(current.Entries, IniEntry{
			Key:       strings.ToLower(strings.TrimSpace(key)),
			Val:       strings.TrimSpace(value),
			Condition: condition,
		})
	}

	if current != nil {
		sections = append(sections, *current)
	}

	return sections, nil
}

func findFileInDirRecursive(
	dir string,
	fileName string,
	depth int,
	maxDepth int,
) (string, bool) {
	if depth > maxDepth {
		return "", false
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if !entry.IsDir() {
			if strings.EqualFold(entry.Name(), fileName) {
				return path, true
			}
			continue
		}

		found, ok := findFileInDirRecursive(
			path,
			fileName,
			depth+1,
			maxDepth,
		)
		if ok {
			return found, true
		}
	}

	return "", false
}

func safeResourcePath(
	baseDir string,
	modDir string,
	relativePath string,
	baseModelDir string,
) (string, bool) {
	relativePath = strings.TrimSpace(relativePath)

	if relativePath == "" || filepath.IsAbs(relativePath) {
		return "", false
	}

	rel := filepath.Clean(filepath.FromSlash(relativePath))

	// Prevent resource paths from escaping their search directories.
	if rel == ".." ||
		strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}

	candidate := filepath.Join(baseDir, rel)
	if fileExists(candidate) {
		return candidate, true
	}

	candidate = filepath.Join(modDir, rel)
	if fileExists(candidate) {
		return candidate, true
	}

	fileName := filepath.Base(rel)

	if found, ok := findFileInDirRecursive(
		modDir,
		fileName,
		0,
		maxResourceSearchDepth,
	); ok {
		return found, true
	}

	if baseModelDir != "" {
		candidate = filepath.Join(baseModelDir, fileName)
		if fileExists(candidate) {
			return candidate, true
		}

		if found, ok := findFileInDirRecursive(
			baseModelDir,
			fileName,
			0,
			4,
		); ok {
			return found, true
		}
	}

	// Keep the unresolved path for diagnostics and later resolution.
	return filepath.Join(baseDir, rel), true
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func cleanComponentName(value string) string {
	name := strings.ToLower(value)

	prefixes := []string{
		"textureoverride",
		"shaderoverride",
		"resource",
		"commandlist",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
			break
		}
	}

	if index := strings.LastIndex(name, "lod"); index >= 0 {
		suffix := name[index+3:]
		if isASCIIDigits(suffix) {
			name = name[:index]
		}
	}

	suffixes := []string{
		"position",
		"texcoord",
		"blend",
		"diffuse1k",
		"normalmap1k",
		"lightmap1k",
		"materialmap1k",
		"diffuse",
		"normalmap",
		"lightmap",
		"materialmap",
		"vertexlimitraise",
		"ib",
		"vb",
	}

	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			name = strings.TrimSuffix(name, suffix)
			break
		}
	}

	if len(name) > 2 {
		last := name[len(name)-1]

		if last >= 'a' && last <= 'z' {
			prefix := name[:len(name)-1]

			knownComponents := []string{
				"body",
				"hair",
				"face",
				"head",
				"dress",
				"legs",
				"arms",
				"mesh",
				"weapon",
			}

			remove := len(prefix) >= 3
			for _, component := range knownComponents {
				if strings.HasSuffix(prefix, component) {
					remove = true
					break
				}
			}

			if remove {
				name = prefix
			}
		}
	}

	return name
}

func isASCIIDigits(value string) bool {
	if value == "" {
		return false
	}

	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}

	return true
}

func stringPtr(value string) *string {
	return &value
}

func cleanResourceReference(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSpace(strings.TrimPrefix(value, "ref "))
	return strings.ToLower(value)
}

func cleanHash(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(strings.ToLower(value), "0x")
	return value
}

func isNullResource(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "null") ||
		strings.TrimSpace(value) == ""
}

func resourceNameMatches(record ResourceRecord, component string) bool {
	name := strings.ToLower(record.Name)
	file := strings.ToLower(record.Filename)

	cleanName := cleanComponentName(record.Name)

	return strings.Contains(name, component) ||
		strings.Contains(file, component) ||
		strings.Contains(component, cleanName)
}

func sortedResourceRecords(resources map[string]ResourceRecord) []ResourceRecord {
	result := make([]ResourceRecord, 0, len(resources))

	for _, record := range resources {
		result = append(result, record)
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) <
			strings.ToLower(result[j].Name)
	})

	return result
}

func resourceHasName(record ResourceRecord, terms ...string) bool {
	name := strings.ToLower(record.Name)
	file := strings.ToLower(record.Filename)

	for _, term := range terms {
		term = strings.ToLower(term)

		if strings.Contains(name, term) ||
			strings.Contains(file, term) {
			return true
		}
	}

	return false
}

func resourceIsTexture(record ResourceRecord) bool {
	ext := strings.ToLower(filepath.Ext(record.Filename))

	switch ext {
	case ".dds", ".png", ".jpg", ".jpeg":
		return true
	default:
		return false
	}
}

func resourceIsIndexBuffer(record ResourceRecord) bool {
	return strings.HasSuffix(strings.ToLower(record.Filename), ".ib") ||
		strings.HasSuffix(strings.ToLower(record.Name), "ib") ||
		(record.Format != nil &&
			strings.Contains(strings.ToUpper(*record.Format), "UINT"))
}

func resourceIsPositionBuffer(record ResourceRecord) bool {
	return resourceHasName(record, "position")
}

func resourceIsTexcoordBuffer(record ResourceRecord) bool {
	return resourceHasName(record, "texcoord")
}

func resourceIndexSize(record ResourceRecord) IndexSize {
	if record.Format != nil &&
		strings.Contains(strings.ToUpper(*record.Format), "R16") {
		return IndexSizeU16
	}

	return IndexSizeU32
}

func resourceStride(record ResourceRecord, minimum int, fallback int) int {
	if record.Stride != nil && *record.Stride >= minimum {
		return *record.Stride
	}

	return fallback
}

func findMatchingResource(
	resources map[string]ResourceRecord,
	component string,
	predicate func(ResourceRecord) bool,
) *ResourceRecord {
	records := sortedResourceRecords(resources)

	for _, record := range records {
		if predicate(record) && resourceNameMatches(record, component) {
			result := record
			return &result
		}
	}

	return nil
}

func extractDrawGroups(
	sections []IniSection,
	modDir string,
	baseModelDir string,
) []DrawGroup {
	resources := make(map[string]ResourceRecord)

	// 1. Collect resource definitions.
	for _, section := range sections {
		lowerName := strings.ToLower(section.Name)
		if !strings.HasPrefix(lowerName, "resource") {
			continue
		}

		var filename string
		var stride *int
		var format *string

		for _, entry := range section.Entries {
			switch entry.Key {
			case "filename":
				filename = entry.Val

			case "stride":
				if value, err := strconv.Atoi(entry.Val); err == nil && value >= 0 {
					stride = &value
				}

			case "format":
				format = stringPtr(entry.Val)
			}
		}

		if filename == "" {
			continue
		}

		fullPath, ok := safeResourcePath(
			section.IniDir,
			modDir,
			filename,
			baseModelDir,
		)
		if !ok {
			continue
		}

		resources[lowerName] = ResourceRecord{
			Name:     section.Name,
			Filename: filename,
			FullPath: fullPath,
			Stride:   stride,
			Format:   format,
		}
	}

	// Resolve a resource name or a direct file reference.
	findResource := func(resourceName string) *ResourceRecord {
		clean := cleanResourceReference(resourceName)
		if clean == "" {
			return nil
		}

		directKey := clean
		if !strings.HasPrefix(directKey, "resource") {
			directKey = "resource" + directKey
		}

		if record, ok := resources[directKey]; ok {
			result := record
			return &result
		}

		if record, ok := resources[clean]; ok {
			result := record
			return &result
		}

		ext := strings.ToLower(filepath.Ext(clean))
		switch ext {
		case ".buf", ".ib", ".dds", ".png", ".jpg", ".jpeg":
			path, ok := safeResourcePath(
				modDir,
				modDir,
				resourceName,
				baseModelDir,
			)
			if !ok {
				return nil
			}

			return &ResourceRecord{
				Name:     resourceName,
				Filename: resourceName,
				FullPath: path,
			}
		}

		return nil
	}

	// 2. Collect component and hash bindings.
	compPos := make(map[string]string)
	compTC := make(map[string]string)

	hashPos := make(map[string]string)
	hashTC := make(map[string]string)
	hashDiffuse := make(map[string]string)

	diffuseMap := make(map[string]string)
	normalMap := make(map[string]string)
	lightMap := make(map[string]string)
	materialMap := make(map[string]string)

	for _, section := range sections {
		sectionLower := strings.ToLower(section.Name)

		if !strings.HasPrefix(sectionLower, "textureoverride") &&
			!strings.HasPrefix(sectionLower, "commandlist") {
			continue
		}

		component := cleanComponentName(section.Name)

		// Collect hash first so bindings work regardless of INI entry order.
		var sectionHash string
		for _, entry := range section.Entries {
			if entry.Key == "hash" {
				sectionHash = cleanHash(entry.Val)
				break
			}
		}

		for _, entry := range section.Entries {
			key := strings.ToLower(entry.Key)
			value := strings.TrimSpace(entry.Val)

			if isNullResource(value) {
				continue
			}

			switch key {
			case "hash":
				continue

			case "vb0":
				if component != "" {
					compPos[component] = value
				}
				if sectionHash != "" {
					hashPos[sectionHash] = value
				}

			case "vb1":
				if component != "" {
					compTC[component] = value
				}
				if sectionHash != "" {
					hashTC[sectionHash] = value
				}

			case "vb2":
				record := findResource(value)
				if record != nil && (record.Stride == nil || *record.Stride != 32) {
					if component != "" {
						if _, exists := compTC[component]; !exists {
							compTC[component] = value
						}
					}
					if sectionHash != "" {
						hashTC[sectionHash] = value
					}
				}

			case "this", "ps-t0":
				isDiffuse := strings.Contains(sectionLower, "diffuse") ||
					strings.Contains(strings.ToLower(value), "diffuse") ||
					key == "ps-t0"

				if isDiffuse {
					if component != "" {
						diffuseMap[component] = value
					}
					if sectionHash != "" {
						hashDiffuse[sectionHash] = value
					}

					rawName := strings.TrimPrefix(section.Name, "TextureOverride")
					diffuseMap[strings.ToLower(rawName)] = value
				} else if strings.Contains(sectionLower, "normal") {
					normalMap[component] = value
					normalMap[sectionLower] = value
				} else if strings.Contains(sectionLower, "light") {
					lightMap[component] = value
					lightMap[sectionLower] = value
				} else if strings.Contains(sectionLower, "material") {
					materialMap[component] = value
					materialMap[sectionLower] = value
				}

			case "ps-t1":
				normalMap[component] = value
				normalMap[sectionLower] = value

			case "ps-t2":
				lightMap[component] = value
				lightMap[sectionLower] = value

			case "ps-t3":
				materialMap[component] = value
				materialMap[sectionLower] = value

			default:
				switch {
				case strings.Contains(key, "diffuse"):
					if component != "" {
						diffuseMap[component] = value
					}
					if sectionHash != "" {
						hashDiffuse[sectionHash] = value
					}

					rawName := strings.TrimPrefix(section.Name, "TextureOverride")
					rawName = strings.TrimPrefix(rawName, "CommandList")
					diffuseMap[strings.ToLower(rawName)] = value

				case strings.Contains(key, "normal"):
					normalMap[component] = value
					normalMap[sectionLower] = value

				case strings.Contains(key, "light"):
					lightMap[component] = value
					lightMap[sectionLower] = value

				case strings.Contains(key, "material"):
					materialMap[component] = value
					materialMap[sectionLower] = value
				}
			}
		}
	}

	// 3. Build a case-insensitive section lookup.
	sectionMap := make(map[string]*IniSection)

	for i := range sections {
		key := strings.ToLower(sections[i].Name)
		sectionMap[key] = &sections[i]
	}

	type scannedSection struct {
		ib           *string
		vb0          *string
		vb1          *string
		diffuse      *string
		normal       *string
		light        *string
		material     *string
		handlingSkip bool
		hash         *string
		draws        []DrawIndexed
		runs         []string
	}

	var scanSectionRecursive func(
		sectionName string,
		out *scannedSection,
		visiting map[string]bool,
	)

	scanSectionRecursive = func(
		sectionName string,
		out *scannedSection,
		visiting map[string]bool,
	) {
		sectionKey := cleanResourceReference(sectionName)
		if visiting[sectionKey] {
			return
		}

		section, ok := sectionMap[sectionKey]
		if !ok {
			return
		}

		visiting[sectionKey] = true
		defer delete(visiting, sectionKey)

		for _, entry := range section.Entries {
			key := strings.ToLower(entry.Key)
			value := strings.TrimSpace(entry.Val)

			if isNullResource(value) {
				continue
			}

			switch key {
			case "ib":
				if out.ib == nil {
					out.ib = stringPtr(value)
				}

			case "vb0":
				if out.vb0 == nil {
					out.vb0 = stringPtr(value)
				}

			case "vb1", "vb2":
				if out.vb1 == nil {
					out.vb1 = stringPtr(value)
				}

			case "ps-t0":
				if out.diffuse == nil {
					out.diffuse = stringPtr(value)
				}

			case "ps-t1":
				if out.normal == nil {
					out.normal = stringPtr(value)
				}

			case "ps-t2":
				if out.light == nil {
					out.light = stringPtr(value)
				}

			case "ps-t3":
				if out.material == nil {
					out.material = stringPtr(value)
				}

			case "this":
				sectionLower := strings.ToLower(section.Name)

				switch {
				case strings.Contains(sectionLower, "diffuse"):
					if out.diffuse == nil {
						out.diffuse = stringPtr(value)
					}
				case strings.Contains(sectionLower, "normal"):
					if out.normal == nil {
						out.normal = stringPtr(value)
					}
				case strings.Contains(sectionLower, "light"):
					if out.light == nil {
						out.light = stringPtr(value)
					}
				case strings.Contains(sectionLower, "material"):
					if out.material == nil {
						out.material = stringPtr(value)
					}
				}

			case "handling":
				if strings.EqualFold(value, "skip") {
					out.handlingSkip = true
				}

			case "hash":
				if out.hash == nil {
					out.hash = stringPtr(cleanHash(value))
				}

			case "drawindexed":
				parts := strings.Split(value, ",")
				for i := range parts {
					parts[i] = strings.TrimSpace(parts[i])
				}

				if len(parts) >= 3 {
					count, errCount := strconv.Atoi(parts[0])
					start, errStart := strconv.Atoi(parts[1])
					base, errBase := strconv.ParseInt(parts[2], 10, 32)

					if errCount == nil &&
						errStart == nil &&
						errBase == nil &&
						count >= 0 &&
						start >= 0 {
						out.draws = append(out.draws, DrawIndexed{
							Count:     count,
							Start:     start,
							Base:      int32(base),
							Condition: entry.Condition,
						})
					}
				} else if len(parts) == 1 && strings.EqualFold(parts[0], "auto") {
					out.draws = append(out.draws, DrawIndexed{
						Count:     0,
						Start:     0,
						Base:      0,
						Condition: entry.Condition,
					})
				}

			case "run":
				target := strings.TrimSpace(strings.TrimPrefix(value, "ref "))
				out.runs = append(out.runs, target)
				scanSectionRecursive(target, out, visiting)
			}
		}
	}

	var groups []DrawGroup

	type drawKey struct {
		ib        string
		start     int
		count     int
		base      int32
		condition string
		position  string
	}

	seenDrawKeys := make(map[drawKey]bool)

	// Iterate in original INI order to preserve section priority.
	for _, section := range sections {
		sectionLower := strings.ToLower(section.Name)

		if !strings.HasPrefix(sectionLower, "textureoverride") &&
			!strings.HasPrefix(sectionLower, "commandlist") {
			continue
		}

		scanned := &scannedSection{
			draws: make([]DrawIndexed, 0),
		}

		scanSectionRecursive(
			section.Name,
			scanned,
			make(map[string]bool),
		)

		if scanned.handlingSkip && len(scanned.draws) == 0 {
			continue
		}

		if scanned.ib == nil && len(scanned.draws) == 0 {
			continue
		}

		componentKey := cleanComponentName(section.Name)

		// Resolve index buffer.
		var ibRecord *ResourceRecord

		if scanned.ib != nil {
			ibRecord = findResource(*scanned.ib)
		}

		if ibRecord == nil {
			ibRecord = findMatchingResource(
				resources,
				componentKey,
				func(record ResourceRecord) bool {
					return strings.HasSuffix(
						strings.ToLower(record.Name),
						"ib",
					) || strings.HasSuffix(
						strings.ToLower(record.Filename),
						".ib",
					)
				},
			)
		}

		if ibRecord == nil || !fileExists(ibRecord.FullPath) {
			continue
		}

		effectiveComponent := cleanComponentName(ibRecord.Name)
		lookupComponent := componentKey

		if effectiveComponent != "" {
			lookupComponent = effectiveComponent
		}

		// Resolve position buffer.
		var positionRecord *ResourceRecord

		if scanned.vb0 != nil {
			positionRecord = findResource(*scanned.vb0)
		}

		if positionRecord == nil {
			if value, ok := compPos[lookupComponent]; ok {
				positionRecord = findResource(value)
			}
		}

		if positionRecord == nil {
			for key, value := range compPos {
				if strings.Contains(lookupComponent, key) ||
					strings.Contains(key, lookupComponent) {
					positionRecord = findResource(value)
					if positionRecord != nil {
						break
					}
				}
			}
		}

		if positionRecord == nil && scanned.hash != nil {
			if value, ok := hashPos[*scanned.hash]; ok {
				positionRecord = findResource(value)
			}
		}

		if positionRecord == nil {
			positionRecord = findMatchingResource(
				resources,
				lookupComponent,
				resourceIsPositionBuffer,
			)
		}

		if positionRecord == nil {
			var candidates []ResourceRecord

			for _, record := range sortedResourceRecords(resources) {
				if resourceIsPositionBuffer(record) {
					candidates = append(candidates, record)
				}
			}

			if len(candidates) == 1 {
				positionRecord = &candidates[0]
			}
		}

		if positionRecord == nil || !fileExists(positionRecord.FullPath) {
			continue
		}

		// Resolve texcoord buffer.
		var texcoordRecord *ResourceRecord

		if scanned.vb1 != nil {
			texcoordRecord = findResource(*scanned.vb1)
		}

		if texcoordRecord == nil {
			if value, ok := compTC[lookupComponent]; ok {
				texcoordRecord = findResource(value)
			}
		}

		if texcoordRecord == nil {
			for key, value := range compTC {
				if strings.Contains(lookupComponent, key) ||
					strings.Contains(key, lookupComponent) {
					texcoordRecord = findResource(value)
					if texcoordRecord != nil {
						break
					}
				}
			}
		}

		if texcoordRecord == nil && scanned.hash != nil {
			if value, ok := hashTC[*scanned.hash]; ok {
				texcoordRecord = findResource(value)
			}
		}

		if texcoordRecord == nil {
			texcoordRecord = findMatchingResource(
				resources,
				lookupComponent,
				resourceIsTexcoordBuffer,
			)
		}

		if texcoordRecord == nil {
			var candidates []ResourceRecord

			for _, record := range sortedResourceRecords(resources) {
				if resourceIsTexcoordBuffer(record) {
					candidates = append(candidates, record)
				}
			}

			if len(candidates) == 1 {
				texcoordRecord = &candidates[0]
			}
		}

		if texcoordRecord != nil && !fileExists(texcoordRecord.FullPath) {
			texcoordRecord = nil
		}

		// Resolve texture resources.
		submeshKey := strings.TrimPrefix(section.Name, "TextureOverride")
		submeshKey = strings.TrimPrefix(submeshKey, "CommandList")
		submeshKey = strings.ToLower(submeshKey)

		resolveTexture := func(
			direct *string,
			mapping map[string]string,
			resourceTerms ...string,
		) *ResourceRecord {
			if direct != nil {
				if record := findResource(*direct); record != nil {
					return record
				}
			}

			if scanned.hash != nil {
				if value, ok := hashDiffuse[*scanned.hash]; ok &&
					resourceTerms[0] == "diffuse" {
					if record := findResource(value); record != nil {
						return record
					}
				}
			}

			if value, ok := mapping[submeshKey]; ok {
				if record := findResource(value); record != nil {
					return record
				}
			}

			if value, ok := mapping[lookupComponent]; ok {
				if record := findResource(value); record != nil {
					return record
				}
			}

			// Match resources by both texture role and component.
			records := sortedResourceRecords(resources)
			for _, record := range records {
				if !resourceIsTexture(record) ||
					!resourceHasName(record, resourceTerms...) {
					continue
				}

				if strings.Contains(strings.ToLower(record.Name), submeshKey) ||
					strings.Contains(strings.ToLower(record.Filename), submeshKey) ||
					strings.Contains(strings.ToLower(record.Name), lookupComponent) ||
					strings.Contains(strings.ToLower(record.Filename), lookupComponent) {
					return &record
				}
			}

			return nil
		}

		diffuseRecord := resolveTexture(
			scanned.diffuse,
			diffuseMap,
			"diffuse",
		)

		normalRecord := resolveTexture(
			scanned.normal,
			normalMap,
			"normal",
		)

		lightRecord := resolveTexture(
			scanned.light,
			lightMap,
			"light",
		)

		materialRecord := resolveTexture(
			scanned.material,
			materialMap,
			"material",
		)

		// If no explicit drawindexed exists, use the entire index buffer.
		draws := scanned.draws
		if len(draws) == 0 {
			draws = []DrawIndexed{{
				Count: 0,
				Start: 0,
				Base:  0,
			}}
		}

		uniqueDraws := make([]DrawIndexed, 0, len(draws))

		for _, draw := range draws {
			condition := ""
			if draw.Condition != nil {
				condition = *draw.Condition
			}

			key := drawKey{
				ib:        ibRecord.FullPath,
				start:     draw.Start,
				count:     draw.Count,
				base:      draw.Base,
				condition: condition,
				position:  positionRecord.FullPath,
			}

			if seenDrawKeys[key] {
				continue
			}

			seenDrawKeys[key] = true
			uniqueDraws = append(uniqueDraws, draw)
		}

		if len(uniqueDraws) == 0 {
			continue
		}

		displayName := strings.TrimPrefix(section.Name, "TextureOverride")
		displayName = strings.TrimPrefix(displayName, "CommandList")

		var texcoordPath *string
		if texcoordRecord != nil {
			texcoordPath = stringPtr(texcoordRecord.FullPath)
		}

		var diffusePath *string
		if diffuseRecord != nil {
			diffusePath = stringPtr(diffuseRecord.FullPath)
		}

		var normalPath *string
		if normalRecord != nil {
			normalPath = stringPtr(normalRecord.FullPath)
		}

		var lightPath *string
		if lightRecord != nil {
			lightPath = stringPtr(lightRecord.FullPath)
		}

		var materialPath *string
		if materialRecord != nil {
			materialPath = stringPtr(materialRecord.FullPath)
		}

		groups = append(groups, DrawGroup{
			Name:         displayName,
			PositionFile: positionRecord.FullPath,
			TexcoordFile: texcoordPath,
			IBFile:       ibRecord.FullPath,
			PositionStride: resourceStride(
				*positionRecord,
				12,
				DefaultPositionStride,
			),
			TexcoordStride: func() int {
				if texcoordRecord == nil {
					return 24
				}
				return resourceStride(*texcoordRecord, 4, 24)
			}(),
			IndexSize:    resourceIndexSize(*ibRecord),
			DiffuseFile:  diffusePath,
			NormalFile:   normalPath,
			LightFile:    lightPath,
			MaterialFile: materialPath,
			Draws:        uniqueDraws,
		})
	}

	// 4. Resource-based synthesis if no INI draw sections were resolved.
	if len(groups) == 0 {
		records := sortedResourceRecords(resources)

		var indexBuffers []ResourceRecord
		for _, record := range records {
			if resourceIsIndexBuffer(record) {
				indexBuffers = append(indexBuffers, record)
			}
		}

		for _, ibRecord := range indexBuffers {
			component := cleanComponentName(ibRecord.Name)

			positionRecord := findMatchingResource(
				resources,
				component,
				func(record ResourceRecord) bool {
					return strings.HasSuffix(
						strings.ToLower(record.Filename),
						".buf",
					) || resourceIsPositionBuffer(record)
				},
			)

			texcoordRecord := findMatchingResource(
				resources,
				component,
				func(record ResourceRecord) bool {
					return strings.HasSuffix(
						strings.ToLower(record.Filename),
						".buf",
					) || resourceIsTexcoordBuffer(record)
				},
			)

			if positionRecord == nil || texcoordRecord == nil {
				continue
			}

			if !fileExists(positionRecord.FullPath) ||
				!fileExists(texcoordRecord.FullPath) ||
				!fileExists(ibRecord.FullPath) {
				continue
			}

			diffuseRecord := findMatchingResource(
				resources,
				component,
				func(record ResourceRecord) bool {
					return resourceIsTexture(record) &&
						resourceHasName(record, "diffuse")
				},
			)

			normalRecord := findMatchingResource(
				resources,
				component,
				func(record ResourceRecord) bool {
					return resourceIsTexture(record) &&
						resourceHasName(record, "normal")
				},
			)

			lightRecord := findMatchingResource(
				resources,
				component,
				func(record ResourceRecord) bool {
					return resourceIsTexture(record) &&
						resourceHasName(record, "light")
				},
			)

			materialRecord := findMatchingResource(
				resources,
				component,
				func(record ResourceRecord) bool {
					return resourceIsTexture(record) &&
						resourceHasName(record, "material")
				},
			)

			var diffusePath *string
			if diffuseRecord != nil {
				diffusePath = stringPtr(diffuseRecord.FullPath)
			}

			var normalPath *string
			if normalRecord != nil {
				normalPath = stringPtr(normalRecord.FullPath)
			}

			var lightPath *string
			if lightRecord != nil {
				lightPath = stringPtr(lightRecord.FullPath)
			}

			var materialPath *string
			if materialRecord != nil {
				materialPath = stringPtr(materialRecord.FullPath)
			}

			groups = append(groups, DrawGroup{
				Name:         component,
				PositionFile: positionRecord.FullPath,
				TexcoordFile: stringPtr(texcoordRecord.FullPath),
				IBFile:       ibRecord.FullPath,
				PositionStride: resourceStride(
					*positionRecord,
					12,
					DefaultPositionStride,
				),
				TexcoordStride: resourceStride(
					*texcoordRecord,
					4,
					24,
				),
				IndexSize:    resourceIndexSize(ibRecord),
				DiffuseFile:  diffusePath,
				NormalFile:   normalPath,
				LightFile:    lightPath,
				MaterialFile: materialPath,
				Draws: []DrawIndexed{{
					Count: 0,
					Start: 0,
					Base:  0,
				}},
			})
		}
	}

	return groups
}

func cleanVariableName(variable string) string {
	raw := strings.TrimPrefix(strings.TrimSpace(variable), "$")
	raw = strings.ReplaceAll(raw, "_", " ")

	words := strings.Fields(raw)
	if len(words) == 0 {
		return variable
	}

	for i, word := range words {
		if len(word) == 0 {
			continue
		}

		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}

	return strings.Join(words, " ")
}

func extractINIToggles(sections []IniSection) []ViewerToggle {
	defaults := make(map[string]int)
	var toggles []ViewerToggle
	seenVariables := make(map[string]bool)

	// 1. Read default values from [Constants].
	for _, section := range sections {
		if !strings.EqualFold(section.Name, "constants") {
			continue
		}

		for _, entry := range section.Entries {
			key := strings.ToLower(strings.TrimSpace(entry.Key))

			variable := ""
			for _, part := range strings.Fields(key) {
				if strings.HasPrefix(part, "$") {
					variable = part
					break
				}
			}

			if variable == "" {
				continue
			}

			value, err := strconv.Atoi(strings.TrimSpace(entry.Val))
			if err == nil {
				defaults[variable] = value
			}
		}
	}

	// 2. Read key bindings and their variable cycles.
	for _, section := range sections {
		if !strings.HasPrefix(strings.ToLower(section.Name), "key") {
			continue
		}

		var keyBind *string

		for _, entry := range section.Entries {
			if entry.Key == "key" {
				keyBind = stringPtr(strings.TrimSpace(entry.Val))
				break
			}
		}

		for _, entry := range section.Entries {
			if !strings.HasPrefix(entry.Key, "$") {
				continue
			}

			variable := strings.TrimSpace(entry.Key)
			if seenVariables[variable] {
				continue
			}

			var values []int
			for _, part := range strings.Split(entry.Val, ",") {
				value, err := strconv.Atoi(strings.TrimSpace(part))
				if err != nil {
					continue
				}

				values = append(values, value)
			}

			if len(values) == 0 {
				continue
			}

			// Deduplicate while preserving order.
			uniqueValues := make([]int, 0, len(values))
			seenValues := make(map[int]bool)

			for _, value := range values {
				if seenValues[value] {
					continue
				}

				seenValues[value] = true
				uniqueValues = append(uniqueValues, value)
			}

			currentValue, ok := defaults[variable]
			if !ok {
				currentValue = uniqueValues[0]
			}

			var labels []string

			switch {
			case equalIntSlices(uniqueValues, []int{0, 1}):
				labels = []string{"Off", "On"}

			case equalIntSlices(uniqueValues, []int{1, 0}):
				labels = []string{"On", "Off"}

			default:
				for i := range uniqueValues {
					labels = append(
						labels,
						fmt.Sprintf("Style %d", i+1),
					)
				}
			}

			toggles = append(toggles, ViewerToggle{
				ID:           section.Name + "_" + strings.TrimPrefix(variable, "$"),
				Name:         cleanVariableName(variable),
				Key:          keyBind,
				Variable:     variable,
				CurrentValue: currentValue,
				Values:       uniqueValues,
				Labels:       labels,
			})

			seenVariables[variable] = true
		}
	}

	return toggles
}

func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func computeSuggestedTags(
	isRetexture bool,
	meshes []MeshData,
	textures map[string]string,
	toggles []ViewerToggle,
) []string {
	var tags []string

	if isRetexture || (len(meshes) == 0 && len(textures) > 0) {
		tags = append(tags, "Retexture")
	}

	if len(toggles) > 0 {
		tags = append(tags, "Toggle")
	}

	hasOutfit := false
	hasWeapon := false

	for _, mesh := range meshes {
		name := strings.ToLower(mesh.Name)

		component := ""
		if mesh.Component != nil {
			component = strings.ToLower(*mesh.Component)
		}

		combined := name + " " + component

		outfitTerms := []string{
			"body",
			"dress",
			"hair",
			"head",
			"cloth",
			"suit",
			"outfit",
			"jacket",
		}

		weaponTerms := []string{
			"weapon",
			"wpn",
			"sword",
			"gun",
			"blade",
			"spear",
			"shield",
			"hammer",
		}

		for _, term := range outfitTerms {
			if strings.Contains(combined, term) {
				hasOutfit = true
				break
			}
		}

		for _, term := range weaponTerms {
			if strings.Contains(combined, term) {
				hasWeapon = true
				break
			}
		}
	}

	if hasOutfit {
		tags = append(tags, "Outfit")
	}

	if hasWeapon {
		tags = append(tags, "Weapon")
	}

	if len(meshes) >= 4 {
		tags = append(tags, "MultiMesh")
	}

	sort.Strings(tags)

	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		if len(result) == 0 || result[len(result)-1] != tag {
			result = append(result, tag)
		}
	}

	return result
}
