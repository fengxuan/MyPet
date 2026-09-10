package main

import (
	"encoding/json"
	"image"
	_ "image/gif"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	_ "golang.org/x/image/webp"
)

const (
	codexCols     = 8
	codexV1Rows   = 9
	codexV2Rows   = 11
	codexCellW    = 192
	codexCellH    = 208
	codexEmptyMax = 8
)

type CodexManifest struct {
	ID                  string `json:"id"`
	DisplayName         string `json:"displayName"`
	Description         string `json:"description"`
	SpritesheetPath     string `json:"spritesheetPath"`
	SpriteVersionNumber int    `json:"spriteVersionNumber"`
}

type codexRowMeta struct {
	name        string
	durationsMS []int
}

// Official Codex atlas rows (hatch-pet animation-rows.md).
var codexRows = []codexRowMeta{
	{name: "idle", durationsMS: []int{280, 110, 110, 140, 140, 320}},
	{name: "running-right", durationsMS: []int{120, 120, 120, 120, 120, 120, 120, 220}},
	{name: "running-left", durationsMS: []int{120, 120, 120, 120, 120, 120, 120, 220}},
	{name: "waving", durationsMS: []int{140, 140, 140, 280}},
	{name: "jumping", durationsMS: []int{140, 140, 140, 140, 280}},
	{name: "failed", durationsMS: []int{140, 140, 140, 140, 140, 140, 140, 240}},
	{name: "waiting", durationsMS: []int{150, 150, 150, 150, 150, 260}},
	{name: "running", durationsMS: []int{120, 120, 120, 120, 120, 220}},
	{name: "review", durationsMS: []int{150, 150, 150, 150, 150, 280}},
}

var (
	codexIdleCandidates  = []int{0, 6, 8}
	codexHoverCandidates = []int{3, 6, 8, 0}
	codexHitCandidates   = []int{4, 5, 7}
)

func looksLikeCodexPet(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "pet.json"))
	if err != nil {
		return false
	}
	var manifest CodexManifest
	if json.Unmarshal(data, &manifest) != nil {
		return false
	}
	if manifest.SpritesheetPath != "" || manifest.ID != "" || manifest.DisplayName != "" {
		return true
	}
	return findCodexSpritesheet(dir, "") != ""
}

func loadCodexPet(dir, name string, timing TimingConfig) (Pet, bool) {
	manifest, rows, err := readCodexPet(dir)
	if err != nil || len(rows) == 0 {
		return Pet{}, false
	}
	animations := codexRowsToAnimations(rows, timing)
	if !hasAnimationFrames(animations) {
		return Pet{}, false
	}
	if name == "" {
		name = filepath.Base(dir)
	}
	if manifest.ID != "" && name == filepath.Base(dir) {
		name = manifest.ID
	}
	return Pet{name: name, animations: animations}, true
}

func appendCodexHomePets(pets []Pet, timing TimingConfig) []Pet {
	homeDir := codexHomePetsDir()
	if homeDir == "" {
		return pets
	}
	seen := make(map[string]bool, len(pets))
	for _, pet := range pets {
		seen[strings.ToLower(pet.name)] = true
		seen[strings.ToLower(filepath.Base(pet.name))] = true
	}
	entries, err := os.ReadDir(homeDir)
	if err != nil {
		return pets
	}
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		child := filepath.Join(homeDir, entry.Name())
		if !looksLikeCodexPet(child) {
			continue
		}
		pet, ok := loadCodexPet(child, entry.Name(), timing)
		if !ok {
			continue
		}
		key := strings.ToLower(pet.name)
		if seen[key] || seen[strings.ToLower(entry.Name())] {
			continue
		}
		seen[key] = true
		pets = append(pets, pet)
	}
	return pets
}

func codexHomePetsDir() string {
	if env := strings.TrimSpace(os.Getenv("CODEX_HOME")); env != "" {
		return filepath.Join(env, "pets")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".codex", "pets")
}

func readCodexPet(dir string) (CodexManifest, [][]*image.RGBA, error) {
	var manifest CodexManifest
	data, err := os.ReadFile(filepath.Join(dir, "pet.json"))
	if err != nil {
		return manifest, nil, err
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return manifest, nil, err
	}
	sheetPath := findCodexSpritesheet(dir, manifest.SpritesheetPath)
	if sheetPath == "" {
		return manifest, nil, os.ErrNotExist
	}
	src, err := decodeImageFile(sheetPath)
	if err != nil {
		return manifest, nil, err
	}
	return manifest, splitCodexAtlas(src, manifest.SpriteVersionNumber), nil
}

func findCodexSpritesheet(dir, declared string) string {
	candidates := make([]string, 0, 8)
	if declared = strings.TrimSpace(declared); declared != "" {
		candidates = append(candidates, filepath.Join(dir, declared))
	}
	for _, name := range []string{"spritesheet.webp", "spritesheet.png", "spritesheet.gif"} {
		candidates = append(candidates, filepath.Join(dir, name))
		candidates = append(candidates, filepath.Join(dir, "assets", name))
	}
	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return ""
}

func decodeImageFile(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoded, _, err := image.Decode(file)
	return decoded, err
}

func inferCodexGrid(width, height, version int) (cols, rows, cellW, cellH int, ok bool) {
	if width <= 0 || height <= 0 || width%codexCols != 0 {
		return 0, 0, 0, 0, false
	}
	cols = codexCols
	cellW = width / cols
	order := []int{codexV1Rows, codexV2Rows}
	if version >= 2 {
		order = []int{codexV2Rows, codexV1Rows}
	}
	if width == codexCols*codexCellW {
		if height == codexV1Rows*codexCellH {
			return cols, codexV1Rows, codexCellW, codexCellH, true
		}
		if height == codexV2Rows*codexCellH {
			return cols, codexV2Rows, codexCellW, codexCellH, true
		}
	}
	for _, candidate := range order {
		if height%candidate == 0 {
			cellH = height / candidate
			if cellH > 0 {
				return cols, candidate, cellW, cellH, true
			}
		}
	}
	return 0, 0, 0, 0, false
}

func splitCodexAtlas(src image.Image, version int) [][]*image.RGBA {
	bounds := src.Bounds()
	cols, rows, cellW, cellH, ok := inferCodexGrid(bounds.Dx(), bounds.Dy(), version)
	if !ok {
		return nil
	}
	out := make([][]*image.RGBA, rows)
	for row := 0; row < rows; row++ {
		frames := make([]*image.RGBA, 0, cols)
		for col := 0; col < cols; col++ {
			rect := image.Rect(
				bounds.Min.X+col*cellW,
				bounds.Min.Y+row*cellH,
				bounds.Min.X+(col+1)*cellW,
				bounds.Min.Y+(row+1)*cellH,
			)
			cell := cropImage(src, rect)
			if codexCellEmpty(cell) {
				break
			}
			frames = append(frames, cell)
		}
		out[row] = frames
	}
	return out
}

func codexCellEmpty(img *image.RGBA) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.RGBAAt(x, y).A > codexEmptyMax {
				return false
			}
		}
	}
	return true
}

func pickCodexRow(rows [][]*image.RGBA, candidates []int) (frames []*image.RGBA, index int) {
	for _, i := range candidates {
		if i >= 0 && i < len(rows) && len(rows[i]) > 0 {
			return rows[i], i
		}
	}
	return nil, -1
}

func fpsFromDurations(durations []int, frameCount int, fallback float64) float64 {
	if frameCount <= 0 {
		return fallback
	}
	total := 0
	for i := 0; i < frameCount; i++ {
		if i < len(durations) {
			total += durations[i]
			continue
		}
		if len(durations) == 0 {
			return fallback
		}
		total += durations[len(durations)-1]
	}
	if total <= 0 {
		return fallback
	}
	return 1000.0 * float64(frameCount) / float64(total)
}

func rowFPS(rowIndex, frameCount int, fallback float64) float64 {
	if rowIndex >= 0 && rowIndex < len(codexRows) {
		return fpsFromDurations(codexRows[rowIndex].durationsMS, frameCount, fallback)
	}
	return fallback
}

func framesToEbiten(frames []*image.RGBA) []*ebiten.Image {
	out := make([]*ebiten.Image, 0, len(frames))
	for _, frame := range frames {
		if frame == nil {
			continue
		}
		out = append(out, ebiten.NewImageFromImage(frame))
	}
	return out
}

func codexRowsToAnimations(rows [][]*image.RGBA, timing TimingConfig) PetAnimations {
	idleFrames, idleRow := pickCodexRow(rows, codexIdleCandidates)
	hoverFrames, hoverRow := pickCodexRow(rows, codexHoverCandidates)
	hitFrames, hitRow := pickCodexRow(rows, codexHitCandidates)
	if len(hoverFrames) == 0 {
		hoverFrames, hoverRow = idleFrames, idleRow
	}
	if len(hitFrames) == 0 {
		hitFrames, hitRow = idleFrames, idleRow
	}
	return PetAnimations{
		idle:  Animation{frames: framesToEbiten(idleFrames), fps: rowFPS(idleRow, len(idleFrames), timing.IdleFPS)},
		hover: Animation{frames: framesToEbiten(hoverFrames), fps: rowFPS(hoverRow, len(hoverFrames), timing.HoverFPS)},
		hit:   Animation{frames: framesToEbiten(hitFrames), fps: rowFPS(hitRow, len(hitFrames), timing.HitFPS)},
	}
}
