package mod3d

import (
	"encoding/base64"
	"encoding/binary"
	"math"
	"slices"
	"sort"
)

type CompactedMesh struct {
	Positions []float32
	UVs       []float32
	Indices   []uint32
	HasUVs    bool
}

// ReadPositions reads float32 XYZ positions from a vertex buffer.
func ReadPositions(data []byte, stride int) []float32 {
	effectiveStride := stride
	if effectiveStride < 12 {
		effectiveStride = DefaultPositionStride
	}

	if effectiveStride <= 0 {
		return nil
	}

	positions := make([]float32, 0, (len(data)/effectiveStride)*3)

	for offset := DefaultPositionOffset; offset+12 <= len(data); offset += effectiveStride {
		x := math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
		y := math.Float32frombits(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		z := math.Float32frombits(binary.LittleEndian.Uint32(data[offset+8 : offset+12]))

		if !isFinite32(x) || !isFinite32(y) || !isFinite32(z) {
			positions = append(positions, 0, 0, 0)
			continue
		}

		positions = append(positions, x, y, z)
	}

	return positions
}

type uvCandidate struct {
	bothLive     bool
	inRangeRatio float64
	spread       float64
	uvOffset     int
	format       UvFormat
}

// DetectUVBest detects the most likely UV format and offset by sampling the buffer.
func DetectUVBest(data []byte, stride int) (int, UvFormat) {
	if stride <= 0 {
		return DefaultUVOffset, UvFormatFloat16
	}

	total := len(data) / stride
	if total == 0 {
		return DefaultUVOffset, UvFormatFloat16
	}

	sampleCount := min(4096, total)
	step := max(total/sampleCount, 1)

	candidates := make([]uvCandidate, 0, 4)

	offsets := []int{0, 4}
	formats := []struct {
		format UvFormat
		size   int
	}{
		{UvFormatFloat16, 4},
		{UvFormatFloat32, 8},
	}

	for _, uvOffset := range offsets {
		for _, candidateFormat := range formats {
			if uvOffset+candidateFormat.size > stride {
				continue
			}

			us := make([]float64, 0, sampleCount)
			vs := make([]float64, 0, sampleCount)
			sampled := 0

			for i := 0; i < total; i += step {
				offset := i*stride + uvOffset
				if offset+candidateFormat.size > len(data) {
					break
				}

				sampled++

				var u, v float64

				switch candidateFormat.format {
				case UvFormatFloat16:
					uBits := binary.LittleEndian.Uint16(data[offset : offset+2])
					vBits := binary.LittleEndian.Uint16(data[offset+2 : offset+4])

					u = float64(halfToFloat32(uBits))
					v = float64(halfToFloat32(vBits))

				case UvFormatFloat32:
					u = float64(math.Float32frombits(
						binary.LittleEndian.Uint32(data[offset : offset+4]),
					))
					v = float64(math.Float32frombits(
						binary.LittleEndian.Uint32(data[offset+4 : offset+8]),
					))
				}

				if isFinite64(u) && isFinite64(v) &&
					u >= -0.01 && u <= 2.0 &&
					v >= -0.01 && v <= 2.0 {
					us = append(us, u)
					vs = append(vs, v)
				}
			}

			if sampled == 0 || len(us) == 0 {
				continue
			}

			inRangeRatio := float64(len(us)) / float64(sampled)
			if inRangeRatio < 0.95 {
				continue
			}

			minU, maxU := minMax64(us)
			minV, maxV := minMax64(vs)

			du := maxU - minU
			dv := maxV - minV

			candidates = append(candidates, uvCandidate{
				bothLive:     du >= 1e-4 && dv >= 1e-4,
				inRangeRatio: inRangeRatio,
				spread:       du + dv,
				uvOffset:     uvOffset,
				format:       candidateFormat.format,
			})
		}
	}

	if len(candidates) == 0 {
		return DefaultUVOffset, UvFormatFloat16
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		a := candidates[i]
		b := candidates[j]

		if a.bothLive != b.bothLive {
			return a.bothLive
		}

		if a.inRangeRatio != b.inRangeRatio {
			return a.inRangeRatio > b.inRangeRatio
		}

		return a.spread > b.spread
	})

	best := candidates[0]
	return best.uvOffset, best.format
}

// ReadTexcoords reads UV coordinates and flips V for Three.js.
func ReadTexcoords(
	data []byte,
	stride int,
	uvOffset int,
	format UvFormat,
) []float32 {
	formatSize := 4
	if format == UvFormatFloat32 {
		formatSize = 8
	}

	effectiveStride := stride
	if effectiveStride < formatSize {
		effectiveStride = 24
	}

	if effectiveStride <= 0 {
		return nil
	}

	uvs := make([]float32, 0, (len(data)/effectiveStride)*2)

	for offset := uvOffset; offset >= 0 && offset+formatSize <= len(data); offset += effectiveStride {
		var u, v float32

		switch format {
		case UvFormatFloat16:
			uBits := binary.LittleEndian.Uint16(data[offset : offset+2])
			vBits := binary.LittleEndian.Uint16(data[offset+2 : offset+4])

			u = halfToFloat32(uBits)
			v = halfToFloat32(vBits)

		case UvFormatFloat32:
			u = math.Float32frombits(
				binary.LittleEndian.Uint32(data[offset : offset+4]),
			)
			v = math.Float32frombits(
				binary.LittleEndian.Uint32(data[offset+4 : offset+8]),
			)
		}

		if isFinite32(u) && isFinite32(v) {
			uvs = append(uvs, u, 1.0-v)
		} else {
			uvs = append(uvs, 0, 0)
		}
	}

	return uvs
}

// ReadIndices reads uint16 or uint32 index buffer values.
func ReadIndices(
	data []byte,
	start int,
	count int,
	indexSize IndexSize,
) []uint32 {
	if indexSize != IndexSizeU16 && indexSize != IndexSizeU32 {
		return nil
	}

	if start < 0 {
		return nil
	}

	safeIndexSize := int(indexSize)
	total := len(data) / safeIndexSize
	if start >= total {
		return nil
	}

	effectiveCount := count
	if effectiveCount == 0 {
		effectiveCount = total - start
	}

	if effectiveCount < 0 {
		return nil
	}

	end := total
	if effectiveCount < total-start {
		end = start + effectiveCount
	}

	indices := make([]uint32, 0, end-start)

	for i := start; i < end; i++ {
		offset := i * safeIndexSize

		var value uint32

		switch indexSize {

		case IndexSizeU16:
			value = uint32(binary.LittleEndian.Uint16(data[offset : offset+2]))
		case IndexSizeU32:
			value = binary.LittleEndian.Uint32(data[offset : offset+4])
		default:
			panic("unexpected modpreview.IndexSize")
		}
		indices = append(indices, value)
	}

	return indices
}

// CompactVertices exports only vertices referenced by valid triangles.
//
// Primitive restart markers and out-of-bounds triangles are discarded.
// UVs are optional; pass nil when the mesh has no UV data.
func CompactVertices(
	allPositions []float32,
	allUVs []float32,
	rawIndices []uint32,
	base int32,
) (*CompactedMesh, bool) {
	vertexCount := len(allPositions) / 3

	if vertexCount == 0 || len(rawIndices) < 3 {
		return nil, false
	}

	validAdjusted := make([]uint32, 0, len(rawIndices))

	for i := 0; i+2 < len(rawIndices); i += 3 {
		r0 := rawIndices[i]
		r1 := rawIndices[i+1]
		r2 := rawIndices[i+2]

		if isPrimitiveRestart(r0) ||
			isPrimitiveRestart(r1) ||
			isPrimitiveRestart(r2) {
			continue
		}

		a0 := int64(r0) + int64(base)
		a1 := int64(r1) + int64(base)
		a2 := int64(r2) + int64(base)

		if !validVertexIndex(a0, vertexCount) ||
			!validVertexIndex(a1, vertexCount) ||
			!validVertexIndex(a2, vertexCount) {
			continue
		}

		validAdjusted = append(
			validAdjusted,
			uint32(a0),
			uint32(a1),
			uint32(a2),
		)
	}

	if len(validAdjusted) == 0 {
		return nil, false
	}

	// Collect unique referenced vertices.
	used := append([]uint32(nil), validAdjusted...)
	slices.Sort(used)

	unique := used[:0]
	for _, index := range used {
		if len(unique) == 0 || unique[len(unique)-1] != index {
			unique = append(unique, index)
		}
	}
	used = unique

	// Build the old-index to compact-index mapping.
	remap := make(map[uint32]uint32, len(used))

	for newIndex, oldIndex := range used {
		remap[oldIndex] = uint32(newIndex)
	}

	// Compact positions.
	compactPositions := make([]float32, 0, len(used)*3)

	for _, vertexIndex := range used {
		offset := int(vertexIndex) * 3

		compactPositions = append(
			compactPositions,
			allPositions[offset],
			allPositions[offset+1],
			allPositions[offset+2],
		)
	}

	// Compact UVs when provided.
	var compactUVs []float32
	hasUVs := allUVs != nil

	if hasUVs {
		uvVertexCount := len(allUVs) / 2
		compactUVs = make([]float32, 0, len(used)*2)

		for _, vertexIndex := range used {
			if int(vertexIndex) < uvVertexCount {
				offset := int(vertexIndex) * 2

				compactUVs = append(
					compactUVs,
					allUVs[offset],
					allUVs[offset+1],
				)
			} else {
				compactUVs = append(compactUVs, 0, 0)
			}
		}
	}

	// Remap indices.
	remappedIndices := make([]uint32, 0, len(validAdjusted))

	for _, index := range validAdjusted {
		remappedIndices = append(remappedIndices, remap[index])
	}

	return &CompactedMesh{
		Positions: compactPositions,
		UVs:       compactUVs,
		Indices:   remappedIndices,
		HasUVs:    hasUVs,
	}, true
}

// EncodeF32Array encodes float32 values as little-endian Base64.
func EncodeF32Array(data []float32) string {
	bytes := make([]byte, len(data)*4)

	for i, value := range data {
		binary.LittleEndian.PutUint32(
			bytes[i*4:],
			math.Float32bits(value),
		)
	}

	return base64.StdEncoding.EncodeToString(bytes)
}

// EncodeU32Array encodes uint32 values as little-endian Base64.
func EncodeU32Array(data []uint32) string {
	bytes := make([]byte, len(data)*4)

	for i, value := range data {
		binary.LittleEndian.PutUint32(
			bytes[i*4:],
			value,
		)
	}

	return base64.StdEncoding.EncodeToString(bytes)
}

func isPrimitiveRestart(index uint32) bool {
	return index == 0xFFFF || index == 0xFFFFFFFF
}

func validVertexIndex(index int64, vertexCount int) bool {
	return index >= 0 && index < int64(vertexCount)
}

func isFinite32(value float32) bool {
	return !float32IsNaN(value) && !float32IsInf(value)
}

func isFinite64(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func float32IsNaN(value float32) bool {
	return math.IsNaN(float64(value))
}

func float32IsInf(value float32) bool {
	return math.IsInf(float64(value), 0)
}

func minMax64(values []float64) (float64, float64) {
	minValue := math.Inf(1)
	maxValue := math.Inf(-1)

	for _, value := range values {
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}

	return minValue, maxValue
}

// halfToFloat32 converts an IEEE 754 binary16 value to float32.
func halfToFloat32(value uint16) float32 {
	sign := uint32(value>>15) << 31
	exponent := uint32((value >> 10) & 0x1F)
	mantissa := uint32(value & 0x03FF)

	var bits uint32

	switch exponent {
	case 0:
		if mantissa == 0 {
			bits = sign
		} else {
			// Normalize subnormal values.
			exponent = 1

			for mantissa&0x0400 == 0 {
				mantissa <<= 1
				exponent--
			}

			mantissa &= 0x03FF

			bits = sign |
				((exponent + (127 - 15)) << 23) |
				(mantissa << 13)
		}

	case 0x1F:
		bits = sign | 0x7F800000 | (mantissa << 13)

	default:
		bits = sign |
			((exponent + (127 - 15)) << 23) |
			(mantissa << 13)
	}

	return math.Float32frombits(bits)
}
