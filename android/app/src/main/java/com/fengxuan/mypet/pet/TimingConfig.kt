package com.fengxuan.mypet.pet

data class TimingConfig(
    val idleFps: Double = 3.0,
    val hoverFps: Double = 4.0,
    val hitFps: Double = 6.0,
    val hitDurationMs: Long = 1350,
)

object TimingParser {
    private const val MIN_FPS = 0.2
    private const val MAX_FPS = 60.0
    private const val MIN_HIT = 0.1
    private const val MAX_HIT = 30.0

    fun parse(text: String): TimingConfig {
        val values = parseValues(text)
        var config = TimingConfig()
        fps(values, "idle_fps")?.let { config = config.copy(idleFps = it) }
        fps(values, "hover_fps")?.let { config = config.copy(hoverFps = it) }
        fps(values, "hit_fps")?.let { config = config.copy(hitFps = it) }
        duration(values, "hit_duration")?.let { config = config.copy(hitDurationMs = (it * 1000).toLong()) }
        return config
    }

    private fun fps(values: Map<String, String>, key: String): Double? {
        val raw = values[key] ?: return null
        val parsed = raw.toDoubleOrNull() ?: return null
        if (parsed.isNaN() || parsed !in MIN_FPS..MAX_FPS) return null
        return parsed
    }

    private fun duration(values: Map<String, String>, key: String): Double? {
        val raw = values[key] ?: return null
        val parsed = raw.toDoubleOrNull() ?: return null
        if (parsed.isNaN() || parsed !in MIN_HIT..MAX_HIT) return null
        return parsed
    }

    private fun parseValues(text: String): Map<String, String> {
        val trimmed = text.trim()
        if (trimmed.startsWith("{")) {
            return parseJsonish(trimmed)
        }
        val values = mutableMapOf<String, String>()
        for (line in trimmed.lineSequence()) {
            val clean = line.trim()
            if (clean.isEmpty() || clean.startsWith("#") || clean.startsWith(";") || clean.startsWith("//")) {
                continue
            }
            val sep = clean.indexOfAny(charArrayOf('=', ':'))
            if (sep <= 0) continue
            val key = clean.substring(0, sep).trim().lowercase().replace("-", "_").replace(" ", "_")
            val value = clean.substring(sep + 1).trim().trim('"', '\'')
            if (key.isNotEmpty()) values[key] = value
        }
        return values
    }

    private fun parseJsonish(text: String): Map<String, String> {
        val values = mutableMapOf<String, String>()
        val regex = Regex("\"([^\"]+)\"\\s*:\\s*(\"[^\"]*\"|-?\\d+(?:\\.\\d+)?|true|false)")
        for (match in regex.findAll(text)) {
            val key = match.groupValues[1].trim().lowercase().replace("-", "_").replace(" ", "_")
            values[key] = match.groupValues[2].trim().trim('"')
        }
        return values
    }
}
