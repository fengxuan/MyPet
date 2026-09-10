package main

import (
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestInferCodexGridStandardSizes(t *testing.T) {
	cols, rows, cellW, cellH, ok := inferCodexGrid(1536, 1872, 0)
	if !ok || cols != 8 || rows != 9 || cellW != 192 || cellH != 208 {
		t.Fatalf("v1: cols=%d rows=%d cell=%dx%d ok=%v", cols, rows, cellW, cellH, ok)
	}
	cols, rows, cellW, cellH, ok = inferCodexGrid(1536, 2288, 2)
	if !ok || cols != 8 || rows != 11 || cellW != 192 || cellH != 208 {
		t.Fatalf("v2: cols=%d rows=%d cell=%dx%d ok=%v", cols, rows, cellW, cellH, ok)
	}
}

func TestInferCodexGridScaledAndTiny(t *testing.T) {
	cols, rows, cellW, cellH, ok := inferCodexGrid(16, 18, 0)
	if !ok || cols != 8 || rows != 9 || cellW != 2 || cellH != 2 {
		t.Fatalf("tiny v1: cols=%d rows=%d cell=%dx%d ok=%v", cols, rows, cellW, cellH, ok)
	}
	cols, rows, _, _, ok = inferCodexGrid(16, 22, 2)
	if !ok || rows != 11 {
		t.Fatalf("tiny v2 rows=%d ok=%v", rows, ok)
	}
	if _, _, _, _, ok = inferCodexGrid(15, 18, 0); ok {
		t.Fatal("expected reject when width is not divisible by 8")
	}
}

func TestSplitCodexAtlasStopsAtEmptyCell(t *testing.T) {
	atlas := newTestAtlas(2, 2, 8, 9)
	fillTestCell(atlas, 0, 0, color.RGBA{R: 255, A: 255})
	fillTestCell(atlas, 1, 0, color.RGBA{G: 255, A: 255})
	fillTestCell(atlas, 0, 3, color.RGBA{B: 255, A: 255})
	fillTestCell(atlas, 1, 3, color.RGBA{R: 255, G: 255, A: 255})
	fillTestCell(atlas, 2, 3, color.RGBA{R: 255, B: 255, A: 255})
	fillTestCell(atlas, 0, 4, color.RGBA{G: 255, B: 255, A: 255})

	rows := splitCodexAtlas(atlas, 0)
	if len(rows) != 9 {
		t.Fatalf("rows=%d", len(rows))
	}
	if len(rows[0]) != 2 {
		t.Fatalf("idle frames=%d", len(rows[0]))
	}
	if len(rows[1]) != 0 {
		t.Fatalf("empty row should have 0 frames, got %d", len(rows[1]))
	}
	if len(rows[3]) != 3 {
		t.Fatalf("waving frames=%d", len(rows[3]))
	}
	if len(rows[4]) != 1 {
		t.Fatalf("jumping frames=%d", len(rows[4]))
	}
}

func TestPickCodexRowFallbacks(t *testing.T) {
	rows := make([][]*image.RGBA, 9)
	rows[0] = []*image.RGBA{image.NewRGBA(image.Rect(0, 0, 1, 1))}
	rows[5] = []*image.RGBA{image.NewRGBA(image.Rect(0, 0, 1, 1))}

	idle, idleRow := pickCodexRow(rows, codexIdleCandidates)
	hover, hoverRow := pickCodexRow(rows, codexHoverCandidates)
	hit, hitRow := pickCodexRow(rows, codexHitCandidates)
	if len(idle) != 1 || idleRow != 0 {
		t.Fatalf("idle row=%d frames=%d", idleRow, len(idle))
	}
	if hoverRow != 0 || len(hover) != 1 {
		t.Fatalf("hover should fall back to idle, got row=%d frames=%d", hoverRow, len(hover))
	}
	if hitRow != 5 {
		t.Fatalf("hit should use failed row, got %d", hitRow)
	}
	if len(hit) != 1 {
		t.Fatalf("hit frames=%d", len(hit))
	}
}

func TestFPSFromDurations(t *testing.T) {
	got := fpsFromDurations([]int{280, 110, 110, 140, 140, 320}, 6, 3)
	if got < 5.4 || got > 5.5 {
		t.Fatalf("idle fps=%v", got)
	}
	got = fpsFromDurations([]int{140, 140, 140, 280}, 4, 4)
	if got < 5.7 || got > 5.8 {
		t.Fatalf("wave fps=%v", got)
	}
	if fpsFromDurations(nil, 0, 3) != 3 {
		t.Fatal("empty should use fallback")
	}
}

func TestLooksLikeCodexPet(t *testing.T) {
	dir := t.TempDir()
	if looksLikeCodexPet(dir) {
		t.Fatal("empty dir should not look like a Codex pet")
	}
	writeJSON(t, filepath.Join(dir, "pet.json"), CodexManifest{
		ID:              "codie",
		DisplayName:     "Codie",
		SpritesheetPath: "spritesheet.png",
	})
	if !looksLikeCodexPet(dir) {
		t.Fatal("pet.json should be detected")
	}
}

func TestReadCodexPetFromFolder(t *testing.T) {
	dir := t.TempDir()
	atlas := newTestAtlas(2, 2, 8, 9)
	fillTestCell(atlas, 0, 0, color.RGBA{R: 10, A: 255})
	fillTestCell(atlas, 1, 0, color.RGBA{R: 20, A: 255})
	fillTestCell(atlas, 0, 3, color.RGBA{G: 10, A: 255})
	fillTestCell(atlas, 0, 4, color.RGBA{B: 10, A: 255})
	sheet := filepath.Join(dir, "spritesheet.png")
	savePNG(t, sheet, atlas)
	writeJSON(t, filepath.Join(dir, "pet.json"), CodexManifest{
		ID:              "codie",
		DisplayName:     "Codie",
		SpritesheetPath: "spritesheet.png",
	})

	manifest, rows, err := readCodexPet(dir)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.ID != "codie" {
		t.Fatalf("id=%s", manifest.ID)
	}
	if len(rows[0]) != 2 || len(rows[3]) != 1 || len(rows[4]) != 1 {
		t.Fatalf("rows idle=%d wave=%d jump=%d", len(rows[0]), len(rows[3]), len(rows[4]))
	}
}

func TestFindCodexSpritesheetFallbacks(t *testing.T) {
	dir := t.TempDir()
	if findCodexSpritesheet(dir, "") != "" {
		t.Fatal("expected no sheet")
	}
	path := filepath.Join(dir, "spritesheet.webp")
	if err := os.WriteFile(path, []byte("not-an-image"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := findCodexSpritesheet(dir, ""); got != path {
		t.Fatalf("got %s", got)
	}
	declared := filepath.Join(dir, "custom.png")
	if err := os.WriteFile(declared, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := findCodexSpritesheet(dir, "custom.png"); got != declared {
		t.Fatalf("declared got %s", got)
	}
}

func TestCollectPetsLoadsCodexFolder(t *testing.T) {
	root := t.TempDir()
	petDir := filepath.Join(root, "codie")
	if err := os.Mkdir(petDir, 0o755); err != nil {
		t.Fatal(err)
	}
	atlas := newTestAtlas(2, 2, 8, 9)
	fillTestCell(atlas, 0, 0, color.RGBA{R: 255, A: 255})
	fillTestCell(atlas, 0, 3, color.RGBA{G: 255, A: 255})
	fillTestCell(atlas, 0, 4, color.RGBA{B: 255, A: 255})
	savePNG(t, filepath.Join(petDir, "spritesheet.png"), atlas)
	writeJSON(t, filepath.Join(petDir, "pet.json"), CodexManifest{
		ID:              "codie",
		SpritesheetPath: "spritesheet.png",
	})

	pets := collectPets(root, "", defaultTiming())
	if len(pets) != 1 {
		t.Fatalf("pets=%d", len(pets))
	}
	if pets[0].name != "codie" {
		t.Fatalf("name=%s", pets[0].name)
	}
	if len(pets[0].animations.idle.frames) != 1 {
		t.Fatalf("idle frames=%d", len(pets[0].animations.idle.frames))
	}
	if len(pets[0].animations.hover.frames) != 1 {
		t.Fatalf("hover frames=%d", len(pets[0].animations.hover.frames))
	}
	if len(pets[0].animations.hit.frames) != 1 {
		t.Fatalf("hit frames=%d", len(pets[0].animations.hit.frames))
	}
}

func TestAppendCodexHomePetsSkipsDuplicates(t *testing.T) {
	home := t.TempDir()
	petDir := filepath.Join(home, "pets", "codie")
	if err := os.MkdirAll(petDir, 0o755); err != nil {
		t.Fatal(err)
	}
	atlas := newTestAtlas(2, 2, 8, 9)
	fillTestCell(atlas, 0, 0, color.RGBA{R: 255, A: 255})
	savePNG(t, filepath.Join(petDir, "spritesheet.png"), atlas)
	writeJSON(t, filepath.Join(petDir, "pet.json"), CodexManifest{ID: "codie", SpritesheetPath: "spritesheet.png"})
	t.Setenv("CODEX_HOME", home)

	existing := []Pet{{name: "codie"}}
	got := appendCodexHomePets(existing, defaultTiming())
	if len(got) != 1 {
		t.Fatalf("duplicate should be skipped, got %d", len(got))
	}

	got = appendCodexHomePets(nil, defaultTiming())
	if len(got) != 1 || got[0].name != "codie" {
		t.Fatalf("home pet not loaded: %+v", got)
	}
}

func newTestAtlas(cellW, cellH, cols, rows int) *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, cellW*cols, cellH*rows))
}

func fillTestCell(img *image.RGBA, col, row int, c color.RGBA) {
	cellW := img.Bounds().Dx() / 8
	cellH := img.Bounds().Dy() / 9
	if img.Bounds().Dy()%9 != 0 && img.Bounds().Dy()%11 == 0 {
		cellH = img.Bounds().Dy() / 11
	}
	x0 := col * cellW
	y0 := row * cellH
	for y := 0; y < cellH; y++ {
		for x := 0; x < cellW; x++ {
			img.SetRGBA(x0+x, y0+y, c)
		}
	}
}

func savePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
