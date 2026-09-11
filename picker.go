package main

import (
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	pickerVisibleCount = 5
	pickerItemHeight   = 54
	pickerWidth        = 308
	pickerPadding      = 6
	pickerThumbSize    = 44
	pickerScrollSpeed  = 0.28
)

func pickerPanelSize() (width, height float64) {
	return pickerWidth, float64(pickerPadding*2 + pickerVisibleCount*pickerItemHeight)
}

func pickerRect() (x0, y0, x1, y1 float64) {
	width, height := pickerPanelSize()
	x0 = (float64(screenWidth) - width) / 2
	y0 = (float64(screenHeight) - height) / 2
	return x0, y0, x0 + width, y0 + height
}

func hitPickerPanel(mx, my int) bool {
	x0, y0, x1, y1 := pickerRect()
	x, y := float64(mx), float64(my)
	return x >= x0 && x < x1 && y >= y0 && y < y1
}

func otherPetIndices(current, total int) []int {
	if total <= 0 {
		return nil
	}
	out := make([]int, 0, total)
	for i := 0; i < total; i++ {
		if i == current {
			continue
		}
		out = append(out, i)
	}
	return out
}

func maxPickerScroll(itemCount int) float64 {
	extra := itemCount - pickerVisibleCount
	if extra < 0 {
		return 0
	}
	return float64(extra)
}

func clampPickerScroll(scroll float64, itemCount int) float64 {
	if scroll < 0 || math.IsNaN(scroll) {
		return 0
	}
	max := maxPickerScroll(itemCount)
	if scroll > max {
		return max
	}
	return scroll
}

func pickerHitIndex(mx, my int, scroll float64, itemCount int) int {
	if itemCount <= 0 || !hitPickerPanel(mx, my) {
		return -1
	}
	x0, y0, _, _ := pickerRect()
	innerX := float64(mx) - x0 - pickerPadding
	innerY := float64(my) - y0 - pickerPadding
	if innerX < 0 || innerY < 0 {
		return -1
	}
	width, _ := pickerPanelSize()
	if innerX >= width-pickerPadding*2 {
		return -1
	}
	contentY := innerY + clampPickerScroll(scroll, itemCount)*pickerItemHeight
	index := int(math.Floor(contentY / pickerItemHeight))
	if index < 0 || index >= itemCount {
		return -1
	}
	return index
}

func (g *Game) openPicker() {
	g.refreshPets()
	if len(g.pets) < 2 {
		return
	}
	g.pickerOpen = true
	g.menuOpen = false
	g.pickerScroll = 0
	g.pickerHover = -1
	g.dragging = false
}

func (g *Game) closePicker() {
	g.pickerOpen = false
	g.pickerHover = -1
}

func (g *Game) selectPet(index int) {
	if index < 0 || index >= len(g.pets) {
		return
	}
	g.petIndex = index
	g.animations = g.pets[index].animations
	g.hitActive = false
	g.hitElapsed = 0
	g.time = 0
	saveCurrentPet(g.pets[index].name)
}

func (g *Game) updatePicker(mx, my int) {
	choices := otherPetIndices(g.petIndex, len(g.pets))
	count := len(choices)
	_, wy := ebiten.Wheel()
	if wy != 0 && hitPickerPanel(mx, my) {
		g.pickerScroll = clampPickerScroll(g.pickerScroll-wy*pickerScrollSpeed, count)
	}
	g.pickerHover = pickerHitIndex(mx, my, g.pickerScroll, count)
}

func petPreviewFrame(pet Pet) *ebiten.Image {
	for _, animation := range []Animation{pet.animations.idle, pet.animations.hover, pet.animations.hit} {
		if len(animation.frames) > 0 {
			return animation.frames[0]
		}
	}
	return nil
}

func displayPetName(name string) string {
	name = strings.ReplaceAll(name, "/", " ")
	name = strings.TrimSpace(name)
	if name == "" {
		return "pet"
	}
	const maxChars = 28
	runes := []rune(name)
	if len(runes) > maxChars {
		return string(runes[:maxChars-1]) + "…"
	}
	return name
}

func (g *Game) drawPicker(screen *ebiten.Image) {
	choices := otherPetIndices(g.petIndex, len(g.pets))
	count := len(choices)
	if count == 0 {
		return
	}
	x0, y0, _, _ := pickerRect()
	width, height := pickerPanelSize()
	if g.pickerBuf == nil {
		g.pickerBuf = ebiten.NewImage(int(width), int(height))
	}
	buf := g.pickerBuf
	buf.Clear()
	vector.DrawFilledRect(buf, 0, 0, float32(width), float32(height), color.RGBA{R: 40, G: 28, B: 22, A: 230}, true)

	scroll := clampPickerScroll(g.pickerScroll, count)
	innerX := float64(pickerPadding)
	innerY := float64(pickerPadding)
	innerW := width - pickerPadding*2
	innerH := height - pickerPadding*2
	start := int(math.Floor(scroll))
	if start < 0 {
		start = 0
	}
	end := start + pickerVisibleCount + 1
	if end > count {
		end = count
	}
	for i := start; i < end; i++ {
		itemTop := innerY + (float64(i)-scroll)*pickerItemHeight
		bg := color.RGBA{R: 58, G: 40, B: 32, A: 255}
		if i == g.pickerHover {
			bg = color.RGBA{R: 92, G: 62, B: 46, A: 255}
		}
		vector.DrawFilledRect(buf, float32(innerX), float32(itemTop), float32(innerW), float32(pickerItemHeight-4), bg, true)
		pet := g.pets[choices[i]]
		thumbX := innerX + 5
		thumbY := itemTop + (pickerItemHeight-4-pickerThumbSize)/2
		drawPickerThumb(buf, pet, thumbX, thumbY, pickerThumbSize)
		name := displayPetName(pet.name)
		text.Draw(buf, name, basicfont.Face7x13, int(thumbX+pickerThumbSize+10), int(itemTop+pickerItemHeight/2+3), color.RGBA{R: 255, G: 244, B: 230, A: 255})
	}

	if count > pickerVisibleCount {
		trackX := width - 9
		trackY := innerY
		trackH := innerH
		vector.DrawFilledRect(buf, float32(trackX), float32(trackY), 4, float32(trackH), color.RGBA{R: 28, G: 18, B: 14, A: 255}, true)
		thumbH := trackH * float64(pickerVisibleCount) / float64(count)
		if thumbH < 12 {
			thumbH = 12
		}
		maxScroll := maxPickerScroll(count)
		thumbY := trackY
		if maxScroll > 0 {
			thumbY += (trackH - thumbH) * (scroll / maxScroll)
		}
		vector.DrawFilledRect(buf, float32(trackX), float32(thumbY), 4, float32(thumbH), color.RGBA{R: 246, G: 164, B: 107, A: 255}, true)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x0, y0)
	screen.DrawImage(buf, op)
}

func drawPickerThumb(screen *ebiten.Image, pet Pet, x, y, size float64) {
	frame := petPreviewFrame(pet)
	if frame == nil {
		vector.DrawFilledCircle(screen, float32(x+size/2), float32(y+size/2), float32(size/2-1), peach, true)
		return
	}
	width, height := frame.Size()
	if width <= 0 || height <= 0 {
		return
	}
	scale := math.Min(size/float64(width), size/float64(height))
	op := &ebiten.DrawImageOptions{}
	if width <= 64 && height <= 64 {
		op.Filter = ebiten.FilterNearest
	}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x+(size-float64(width)*scale)/2, y+(size-float64(height)*scale)/2)
	screen.DrawImage(frame, op)
}
