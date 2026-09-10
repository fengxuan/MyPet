package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultIdleFPS     = 3.0
	defaultHoverFPS    = 4.0
	defaultHitFPS      = 6.0
	defaultHitDuration = 1.35
	minFPS             = 0.2
	maxFPS             = 60.0
	minHitDuration     = 0.1
	maxHitDuration     = 30.0
	configFileName     = "config.txt"
)

type TimingConfig struct {
	IdleFPS     float64
	HoverFPS    float64
	HitFPS      float64
	HitDuration float64
}

func defaultTiming() TimingConfig {
	return TimingConfig{
		IdleFPS:     defaultIdleFPS,
		HoverFPS:    defaultHoverFPS,
		HitFPS:      defaultHitFPS,
		HitDuration: defaultHitDuration,
	}
}

func configPath(assetsRoot string) string {
	return filepath.Join(assetsRoot, configFileName)
}

func loadTimingConfig(assetsRoot string) TimingConfig {
	path := configPath(assetsRoot)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			_ = os.WriteFile(path, []byte(defaultConfigText), 0o644)
		}
		return defaultTiming()
	}
	return parseTimingConfig(string(data))
}

func parseTimingConfig(text string) TimingConfig {
	timing := defaultTiming()
	values, ok := parseConfigValues(text)
	if !ok {
		return timing
	}
	if fps, ok := positiveFloatInRange(values, "idle_fps", minFPS, maxFPS); ok {
		timing.IdleFPS = fps
	}
	if fps, ok := positiveFloatInRange(values, "hover_fps", minFPS, maxFPS); ok {
		timing.HoverFPS = fps
	}
	if fps, ok := positiveFloatInRange(values, "hit_fps", minFPS, maxFPS); ok {
		timing.HitFPS = fps
	}
	if duration, ok := positiveFloatInRange(values, "hit_duration", minHitDuration, maxHitDuration); ok {
		timing.HitDuration = duration
	}
	return timing
}

func parseConfigValues(text string) (map[string]string, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, false
	}
	if strings.HasPrefix(trimmed, "{") {
		return parseJSONConfig(trimmed)
	}
	return parseKeyValueConfig(text)
}

func parseJSONConfig(text string) (map[string]string, bool) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return nil, false
	}
	values := make(map[string]string, len(raw))
	for key, value := range raw {
		normalized := normalizeConfigKey(key)
		switch typed := value.(type) {
		case string:
			values[normalized] = typed
		case float64:
			values[normalized] = strconv.FormatFloat(typed, 'f', -1, 64)
		case json.Number:
			values[normalized] = typed.String()
		default:
			// Skip bools, arrays and objects so one bad field does not drop the file.
		}
	}
	return values, true
}

func parseKeyValueConfig(text string) (map[string]string, bool) {
	values := make(map[string]string)
	found := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "//") {
			continue
		}
		sep := strings.IndexAny(line, "=:")
		if sep <= 0 {
			continue
		}
		key := normalizeConfigKey(line[:sep])
		value := strings.TrimSpace(line[sep+1:])
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		values[key] = value
		found = true
	}
	if !found {
		return nil, false
	}
	return values, true
}

func normalizeConfigKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, " ", "_")
	return key
}

func positiveFloatInRange(values map[string]string, key string, min, max float64) (float64, bool) {
	raw, exists := values[key]
	if !exists {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) || parsed < min || parsed > max {
		return 0, false
	}
	return parsed, true
}

const defaultConfigText = `# MyPet 动画速度
# 改完后重新启动，或对小猫右键弹出「下一只」时会重新读取。
# 某一项写错、留空或超出范围时，只该项使用默认值，其它项仍生效。

# 每秒播放多少帧，数字越小动作越慢。可用范围 0.2 ~ 60
idle_fps = 3
hover_fps = 4
hit_fps = 6

# 拍打动作持续秒数。可用范围 0.1 ~ 30
hit_duration = 1.35
`
