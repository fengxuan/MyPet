package main

import (
	"image"
	"image/color"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 360
	screenHeight = 360
)

var (
	ink      = color.RGBA{R: 255, G: 250, B: 239, A: 255}
	inkDeep  = color.RGBA{R: 255, G: 230, B: 184, A: 255}
	peach    = color.RGBA{R: 246, G: 164, B: 107, A: 255}
	peach2   = color.RGBA{R: 255, G: 191, B: 137, A: 255}
	petal    = color.RGBA{R: 255, G: 160, B: 190, A: 255}
	purple   = color.RGBA{R: 55, G: 39, B: 77, A: 255}
	shadow   = color.RGBA{R: 74, G: 49, B: 96, A: 42}
	whitePix = newWhitePixel()
)

type Game struct {
	time                   float64
	hovered                bool
	dragging               bool
	dragStartWindowX       int
	dragStartWindowY       int
	dragStartCursorScreenX int
	dragStartCursorScreenY int
	lastDragWindowX        int
	lastDragWindowY        int
	cursorToWindowX        float64
	cursorToWindowY        float64
	dragMoved              bool
	hitActive              bool
	hitElapsed             float64
	animations             PetAnimations
}

type Animation struct {
	frames []*ebiten.Image
	fps    float64
}

type PetAnimations struct {
	idle  Animation
	hover Animation
	hit   Animation
}

const hitDuration = 0.65

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.triggerHit()
	}

	mx, my := ebiten.CursorPosition()
	mouseDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if !mouseDown {
		if g.dragging && !g.dragMoved {
			g.triggerHit()
		}
		g.dragging = false
	}

	if !g.dragging {
		petX, petY := float64(screenWidth/2), 205.0
		dx, dy := float64(mx)-petX, float64(my)-(petY+math.Sin(g.time*2.2)*5)
		g.hovered = dx*dx/170.0/170.0+dy*dy/165.0/165.0 < 1
	}
	if !g.dragging && g.hovered && mouseDown {
		g.dragging = true
		g.dragStartWindowX, g.dragStartWindowY = ebiten.WindowPosition()
		g.dragStartCursorScreenX = g.dragStartWindowX + int(float64(mx)*g.cursorToWindowX)
		g.dragStartCursorScreenY = g.dragStartWindowY + int(float64(my)*g.cursorToWindowY)
		g.lastDragWindowX, g.lastDragWindowY = g.dragStartWindowX, g.dragStartWindowY
		g.dragMoved = false
	}
	if g.dragging {
		// Keep the hover pose stable while the native transparent window is moving.
		g.hovered = true
		// CursorPosition is relative to the window. Reconstruct its desktop position
		// before calculating the drag delta, so moving the window does not feed back
		// into the next cursor sample.
		windowX, windowY := ebiten.WindowPosition()
		cursorScreenX := windowX + int(float64(mx)*g.cursorToWindowX)
		cursorScreenY := windowY + int(float64(my)*g.cursorToWindowY)
		dragX := cursorScreenX - g.dragStartCursorScreenX
		dragY := cursorScreenY - g.dragStartCursorScreenY
		targetWindowX := g.dragStartWindowX + dragX
		targetWindowY := g.dragStartWindowY + dragY
		if targetWindowX != g.lastDragWindowX || targetWindowY != g.lastDragWindowY {
			ebiten.SetWindowPosition(targetWindowX, targetWindowY)
			g.lastDragWindowX, g.lastDragWindowY = targetWindowX, targetWindowY
			g.dragMoved = true
		}
	}
	if !g.dragging {
		g.time += 1.0 / 60.0
		if g.hitActive {
			g.hitElapsed += 1.0 / 60.0
			if g.hitElapsed >= hitDuration {
				g.hitActive = false
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// The transparent window can be moved by the OS while dragging. Reusing the
	// last rendered frame avoids exposing a cleared frame during that move.
	if g.dragging {
		return
	}
	screen.Clear()

	petY := 205.0 + math.Sin(g.time*2.2)*5
	if animation, elapsed, ok := g.currentAnimation(); ok {
		drawAnimationFrame(screen, animation, elapsed, screenWidth/2, screenHeight/2)
		return
	}
	drawPet(screen, screenWidth/2, petY, g.hovered, g.time, g.hitActive, g.hitElapsed)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.cursorToWindowX = float64(outsideWidth) / float64(screenWidth)
	g.cursorToWindowY = float64(outsideHeight) / float64(screenHeight)
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetScreenClearedEveryFrame(false)

	runOptions := &ebiten.RunGameOptions{
		ScreenTransparent: true,
		SkipTaskbar:       true,
	}
	game := &Game{animations: loadPetAnimations(findAssetsRoot())}
	if err := ebiten.RunGameWithOptions(game, runOptions); err != nil {
		panic(err)
	}
}

func findAssetsRoot() string {
	candidates := []string{"assets"}
	if executable, err := os.Executable(); err == nil {
		executableDir := filepath.Dir(executable)
		candidates = append(candidates,
			filepath.Join(executableDir, "assets"),
			filepath.Join(executableDir, "..", "assets"),
		)
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return "assets"
}

func loadPetAnimations(root string) PetAnimations {
	return PetAnimations{
		idle:  loadAnimation(filepath.Join(root, "idle"), 8),
		hover: loadAnimation(filepath.Join(root, "hover"), 10),
		hit:   loadAnimation(filepath.Join(root, "hit"), 12),
	}
}

func loadAnimation(directory string, fps float64) Animation {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return Animation{fps: fps}
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".png") {
			continue
		}
		paths = append(paths, filepath.Join(directory, entry.Name()))
	}
	sort.Strings(paths)

	frames := make([]*ebiten.Image, 0, len(paths))
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		decoded, _, decodeErr := image.Decode(file)
		_ = file.Close()
		if decodeErr != nil {
			continue
		}
		frames = append(frames, ebiten.NewImageFromImage(decoded))
	}

	return Animation{frames: frames, fps: fps}
}

func (g *Game) triggerHit() {
	g.hitActive = true
	g.hitElapsed = 0
}

func (g *Game) currentAnimation() (Animation, float64, bool) {
	if g.hitActive && len(g.animations.hit.frames) > 0 {
		return g.animations.hit, g.hitElapsed, true
	}
	if g.hovered && len(g.animations.hover.frames) > 0 {
		return g.animations.hover, g.time, true
	}
	if len(g.animations.idle.frames) > 0 {
		return g.animations.idle, g.time, true
	}
	return Animation{}, 0, false
}

func drawAnimationFrame(screen *ebiten.Image, animation Animation, elapsed float64, cx, cy float64) {
	frameIndex := int(elapsed*animation.fps) % len(animation.frames)
	frame := animation.frames[frameIndex]
	width, height := frame.Size()
	scale := math.Min(330/float64(width), 330/float64(height))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(cx-float64(width)*scale/2, cy-float64(height)*scale/2)
	screen.DrawImage(frame, op)
}

func drawPet(screen *ebiten.Image, cx, cy float64, hovered bool, t float64, hitting bool, hitElapsed float64) {
	// The shadow is intentionally made of overlapping circles so it stays crisp at every scale.
	for i := 0; i < 7; i++ {
		vector.DrawFilledCircle(screen, float32(cx-45+float64(i)*15), float32(cy+143), 25, shadow, true)
	}

	// Tail behind the body.
	tailWave := math.Sin(t*4.0) * 11
	vector.StrokeLine(screen, float32(cx+66), float32(cy+50), float32(cx+116), float32(cy+37+tailWave), 22, purple, true)
	vector.StrokeLine(screen, float32(cx+66), float32(cy+50), float32(cx+116), float32(cy+37+tailWave), 12, peach2, true)
	vector.StrokeLine(screen, float32(cx+116), float32(cy+37+tailWave), float32(cx+135), float32(cy-3+tailWave), 22, purple, true)
	vector.StrokeLine(screen, float32(cx+116), float32(cy+37+tailWave), float32(cx+135), float32(cy-3+tailWave), 12, peach2, true)
	vector.DrawFilledCircle(screen, float32(cx+135), float32(cy-3+tailWave), 11, purple, true)
	vector.DrawFilledCircle(screen, float32(cx+135), float32(cy-3+tailWave), 6, peach2, true)

	// Body and belly.
	vector.DrawFilledCircle(screen, float32(cx), float32(cy+54), 101, purple, true)
	vector.DrawFilledCircle(screen, float32(cx), float32(cy+54), 92, inkDeep, true)
	vector.DrawFilledCircle(screen, float32(cx), float32(cy+74), 59, peach2, true)
	vector.DrawFilledCircle(screen, float32(cx-18), float32(cy+57), 16, color.RGBA{R: 255, G: 212, B: 165, A: 255}, true)

	// Ears.
	drawTriangle(screen, cx-83, cy-75, cx-61, cy-151, cx-18, cy-91, purple)
	drawTriangle(screen, cx+83, cy-75, cx+61, cy-151, cx+18, cy-91, purple)
	drawTriangle(screen, cx-70, cy-88, cx-60, cy-130, cx-37, cy-96, petal)
	drawTriangle(screen, cx+70, cy-88, cx+60, cy-130, cx+37, cy-96, petal)

	// Head.
	vector.DrawFilledCircle(screen, float32(cx), float32(cy-55), 91, purple, true)
	vector.DrawFilledCircle(screen, float32(cx), float32(cy-55), 83, ink, true)
	vector.DrawFilledCircle(screen, float32(cx-51), float32(cy-20), 19, peach2, true)
	vector.DrawFilledCircle(screen, float32(cx+51), float32(cy-20), 19, peach2, true)

	// Eyes change from sleepy lines to bright, attentive eyes on hover.
	if hovered {
		vector.DrawFilledCircle(screen, float32(cx-31), float32(cy-64), 11, purple, true)
		vector.DrawFilledCircle(screen, float32(cx+31), float32(cy-64), 11, purple, true)
		vector.DrawFilledCircle(screen, float32(cx-27), float32(cy-68), 3.5, color.White, true)
		vector.DrawFilledCircle(screen, float32(cx+35), float32(cy-68), 3.5, color.White, true)
	} else {
		vector.StrokeLine(screen, float32(cx-42), float32(cy-61), float32(cx-29), float32(cy-55), 5, purple, true)
		vector.StrokeLine(screen, float32(cx-29), float32(cy-55), float32(cx-17), float32(cy-61), 5, purple, true)
		vector.StrokeLine(screen, float32(cx+17), float32(cy-61), float32(cx+29), float32(cy-55), 5, purple, true)
		vector.StrokeLine(screen, float32(cx+29), float32(cy-55), float32(cx+42), float32(cy-61), 5, purple, true)
	}

	// Muzzle, nose and smile.
	vector.DrawFilledCircle(screen, float32(cx-15), float32(cy-38), 18, color.RGBA{R: 255, G: 235, B: 214, A: 255}, true)
	vector.DrawFilledCircle(screen, float32(cx+15), float32(cy-38), 18, color.RGBA{R: 255, G: 235, B: 214, A: 255}, true)
	drawTriangle(screen, cx-8, cy-43, cx+8, cy-43, cx, cy-32, petal)
	vector.StrokeLine(screen, float32(cx), float32(cy-32), float32(cx), float32(cy-24), 3, purple, true)
	vector.StrokeLine(screen, float32(cx), float32(cy-24), float32(cx-11), float32(cy-20), 3, purple, true)
	vector.StrokeLine(screen, float32(cx), float32(cy-24), float32(cx+11), float32(cy-20), 3, purple, true)

	// Whiskers and cheeks.
	for _, side := range []float64{-1, 1} {
		vector.DrawFilledCircle(screen, float32(cx+side*51), float32(cy-35), 7, petal, true)
		vector.StrokeLine(screen, float32(cx+side*39), float32(cy-30), float32(cx+side*83), float32(cy-24), 2, purple, true)
		vector.StrokeLine(screen, float32(cx+side*39), float32(cy-24), float32(cx+side*85), float32(cy-7), 2, purple, true)
	}

	// Paws wave up when the mouse is over the pet.
	pawLift := 0.0
	pawOffset := 77.0
	if hovered {
		pawLift = -39 + math.Sin(t*7)*5
	}
	if hitting {
		// The fallback pose mimics two paws striking the invisible desktop.
		pawLift = 15 + math.Abs(math.Sin(hitElapsed*18))*10
		pawOffset = 47
	}
	for _, side := range []float64{-1, 1} {
		px := cx + side*pawOffset
		py := cy + 93 + pawLift
		vector.DrawFilledCircle(screen, float32(px), float32(py), 29, purple, true)
		vector.DrawFilledCircle(screen, float32(px), float32(py-2), 21, peach2, true)
		for i := -1; i <= 1; i++ {
			vector.DrawFilledCircle(screen, float32(px+float64(i)*8), float32(py-7), 3, purple, true)
		}
	}

	if hovered {
		// Little attention marks above the ears.
		vector.StrokeLine(screen, float32(cx-27), float32(cy-160), float32(cx-36), float32(cy-177), 4, peach, true)
		vector.StrokeLine(screen, float32(cx+27), float32(cy-160), float32(cx+36), float32(cy-177), 4, peach, true)
	}
	if hitting {
		vector.StrokeLine(screen, float32(cx-57), float32(cy+128), float32(cx-72), float32(cy+143), 4, peach, true)
		vector.StrokeLine(screen, float32(cx+57), float32(cy+128), float32(cx+72), float32(cy+143), 4, peach, true)
	}
}

func drawTriangle(screen *ebiten.Image, x1, y1, x2, y2, x3, y3 float64, c color.Color) {
	r, g, b, a := c.RGBA()
	vertices := []ebiten.Vertex{
		{DstX: float32(x1), DstY: float32(y1), SrcX: 0.5, SrcY: 0.5, ColorR: float32(r) / 0xffff, ColorG: float32(g) / 0xffff, ColorB: float32(b) / 0xffff, ColorA: float32(a) / 0xffff},
		{DstX: float32(x2), DstY: float32(y2), SrcX: 0.5, SrcY: 0.5, ColorR: float32(r) / 0xffff, ColorG: float32(g) / 0xffff, ColorB: float32(b) / 0xffff, ColorA: float32(a) / 0xffff},
		{DstX: float32(x3), DstY: float32(y3), SrcX: 0.5, SrcY: 0.5, ColorR: float32(r) / 0xffff, ColorG: float32(g) / 0xffff, ColorB: float32(b) / 0xffff, ColorA: float32(a) / 0xffff},
	}
	screen.DrawTriangles(vertices, []uint16{0, 1, 2}, whitePix, &ebiten.DrawTrianglesOptions{AntiAlias: true, ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha})
}

func newWhitePixel() *ebiten.Image {
	pixel := ebiten.NewImage(1, 1)
	pixel.Fill(color.White)
	return pixel
}
