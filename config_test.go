package main

import "testing"

func TestParseTimingConfigDefaultsOnGarbage(t *testing.T) {
	got := parseTimingConfig("this is not a config :)")
	if got != defaultTiming() {
		t.Fatalf("garbage config: %+v", got)
	}
}

func TestParseTimingConfigPartialInvalid(t *testing.T) {
	got := parseTimingConfig(`
idle_fps = 2
hover_fps = no
hit_fps = -3
hit_duration = 2.5
unknown_key = 99
`)
	if got.IdleFPS != 2 {
		t.Fatalf("idle_fps=%v", got.IdleFPS)
	}
	if got.HoverFPS != defaultHoverFPS {
		t.Fatalf("hover_fps=%v", got.HoverFPS)
	}
	if got.HitFPS != defaultHitFPS {
		t.Fatalf("hit_fps=%v", got.HitFPS)
	}
	if got.HitDuration != 2.5 {
		t.Fatalf("hit_duration=%v", got.HitDuration)
	}
}

func TestParseTimingConfigJSON(t *testing.T) {
	got := parseTimingConfig(`{"idle_fps": 1.5, "hover_fps": "5", "hit_fps": 1000, "hit_duration": true}`)
	if got.IdleFPS != 1.5 {
		t.Fatalf("idle_fps=%v", got.IdleFPS)
	}
	if got.HoverFPS != 5 {
		t.Fatalf("hover_fps=%v", got.HoverFPS)
	}
	if got.HitFPS != defaultHitFPS {
		t.Fatalf("hit_fps=%v, want default", got.HitFPS)
	}
	if got.HitDuration != defaultHitDuration {
		t.Fatalf("hit_duration=%v, want default", got.HitDuration)
	}
}

func TestParseTimingConfigEmptyUsesDefaults(t *testing.T) {
	got := parseTimingConfig("   \n# only comments\n")
	if got != defaultTiming() {
		t.Fatalf("empty config: %+v", got)
	}
}
