package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type MeshData struct {
	VertexCount int      `json:"vertexCount"`
	Indices     []uint16 `json:"indices"`

	Positions    []float32 `json:"positions"`
	Normals      []float32 `json:"normals"`
	Tangents     []float32 `json:"tangents"`
	BlendWeights []float32 `json:"blendWeights"`
	BlendIndices []int32   `json:"blendIndices"`
	Colors       []float32 `json:"colors"`
	UVs          []float32 `json:"uvs"`
}

const TEST_VB = "C:\\Users\\david\\dev\\GI-Model-Importer-Assets-main\\PlayerCharacterData\\HuTao\\HuTaoBody-vb0=dd16576c.txt"
const TEST_IB = "C:\\Users\\david\\dev\\GI-Model-Importer-Assets-main\\PlayerCharacterData\\HuTao\\HuTaoBody-ib=3de1efe2.txt"

func (ms *ModService) GetParsedData() (*MeshData, error) {
	vbf, err := os.Open(TEST_VB)
	if err != nil {
		panic(err)
	}
	defer vbf.Close()

	ibf, err := os.Open(TEST_IB)
	if err != nil {
		panic(err)
	}
	defer ibf.Close()

	data, err := parseMesh(vbf, ibf, 14427)

	return data, nil
}

func parseMesh(vb, ib io.Reader, vertexCount int) (*MeshData, error) {
	mesh, err := parseVertexDump(vb, vertexCount)
	if err != nil {
		return nil, err
	}

	return mesh, nil
}

// vb0[0]+000 POSITION: 0.2308814972639084, 1.137562870979309, -0.24797916412353516
var vertexLine = regexp.MustCompile(
	`^vb0\[(\d+)\]\+\d+\s+(\w+):\s*(.+)$`,
)

func parseVertexDump(r io.Reader, vertexCount int) (*MeshData, error) {
	mesh := &MeshData{
		VertexCount:  vertexCount,
		Positions:    make([]float32, vertexCount*3),
		Normals:      make([]float32, vertexCount*3),
		Tangents:     make([]float32, vertexCount*4),
		BlendWeights: make([]float32, vertexCount*4),
		BlendIndices: make([]int32, vertexCount*4),
		Colors:       make([]float32, vertexCount*4),
		UVs:          make([]float32, vertexCount*2),
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 1024*1024)

	for scanner.Scan() {
		matches := vertexLine.FindStringSubmatch(scanner.Text())
		if matches == nil {
			continue
		}

		index, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, err
		}

		if index < 0 || index >= vertexCount {
			return nil, fmt.Errorf("vertex index %d out of range", index)
		}

		values, err := parseValues(matches[3])
		if err != nil {
			return nil, fmt.Errorf("vertex %d: %w", index, err)
		}

		var dst []float32
		components := 0

		switch matches[2] {
		case "POSITION":
			dst, components = mesh.Positions, 3
		case "NORMAL":
			dst, components = mesh.Normals, 3
		case "TANGENT":
			dst, components = mesh.Tangents, 4
		case "BLENDWEIGHT":
			dst, components = mesh.BlendWeights, 4
		case "COLOR":
			dst, components = mesh.Colors, 4
		case "TEXCOORD":
			dst, components = mesh.UVs, 2
		case "BLENDINDICES":
			if len(values) != 4 {
				return nil, fmt.Errorf("vertex %d: expected 4 blend indices", index)
			}

			for i, value := range values {
				mesh.BlendIndices[index*4+i] = int32(value)
			}
			continue
		default:
			continue
		}

		if len(values) != components {
			return nil, fmt.Errorf(
				"vertex %d %s: expected %d components, got %d",
				index, matches[2], components, len(values),
			)
		}

		copy(dst[index*components:], values)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return mesh, nil
}

func parseIndexDump(r io.Reader) ([]uint16, error) {
	var indices []uint16

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 3 {
			continue
		}

		for _, field := range fields {
			value, err := strconv.ParseUint(field, 10, 16)
			if err != nil {
				return nil, err
			}

			indices = append(indices, uint16(value))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return indices, nil
}

func parseValues(s string) ([]float32, error) {
	parts := strings.Split(s, ",")
	values := make([]float32, len(parts))

	for i, part := range parts {
		value, err := strconv.ParseFloat(strings.TrimSpace(part), 32)
		if err != nil {
			return nil, err
		}
		values[i] = float32(value)
	}

	return values, nil
}
