package main

import "testing"

func TestOtherPetIndicesSkipsCurrent(t *testing.T) {
	got := otherPetIndices(2, 5)
	want := []int{0, 1, 3, 4}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	if otherPetIndices(0, 1) != nil && len(otherPetIndices(0, 1)) != 0 {
		t.Fatalf("single pet should have no others: %v", otherPetIndices(0, 1))
	}
}

func TestClampPickerScroll(t *testing.T) {
	if got := clampPickerScroll(-3, 12); got != 0 {
		t.Fatalf("min=%v", got)
	}
	if got := clampPickerScroll(100, 8); got != 3 {
		t.Fatalf("max=%v want 3", got)
	}
	if got := clampPickerScroll(1.5, 4); got != 0 {
		t.Fatalf("fewer than 5 items should not scroll, got %v", got)
	}
	if got := clampPickerScroll(2, 5); got != 0 {
		t.Fatalf("exactly 5 items should not scroll, got %v", got)
	}
}

func TestPickerHitIndexVisibleRows(t *testing.T) {
	x0, y0, _, _ := pickerRect()
	mx := int(x0 + 40)
	// first visible row
	my := int(y0 + pickerPadding + pickerItemHeight/2)
	if got := pickerHitIndex(mx, my, 0, 9); got != 0 {
		t.Fatalf("first row=%d", got)
	}
	// fifth visible row
	my = int(y0 + pickerPadding + pickerItemHeight*4 + pickerItemHeight/2)
	if got := pickerHitIndex(mx, my, 0, 9); got != 4 {
		t.Fatalf("fifth row=%d", got)
	}
	// scrolled down by 2 items: first visible is index 2
	my = int(y0 + pickerPadding + pickerItemHeight/2)
	if got := pickerHitIndex(mx, my, 2, 9); got != 2 {
		t.Fatalf("scrolled first row=%d", got)
	}
	if got := pickerHitIndex(0, 0, 0, 9); got != -1 {
		t.Fatalf("outside=%d", got)
	}
}

func TestPickerHitIndexIgnoresEmptyList(t *testing.T) {
	x0, y0, _, _ := pickerRect()
	if got := pickerHitIndex(int(x0+20), int(y0+20), 0, 0); got != -1 {
		t.Fatalf("empty=%d", got)
	}
}

func TestDisplayPetName(t *testing.T) {
	if got := displayPetName("pack/cat-1"); got != "pack cat-1" {
		t.Fatalf("got %q", got)
	}
	if got := displayPetName(""); got != "pet" {
		t.Fatalf("empty=%q", got)
	}
}

func TestPickerVisibleDefaultIsFive(t *testing.T) {
	if pickerVisibleCount != 5 {
		t.Fatalf("visible=%d", pickerVisibleCount)
	}
	_, height := pickerPanelSize()
	want := float64(pickerPadding*2 + 5*pickerItemHeight)
	if height != want {
		t.Fatalf("panel height=%v want %v", height, want)
	}
}
