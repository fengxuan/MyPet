package com.fengxuan.mypet.pet

import org.junit.Assert.assertEquals
import org.junit.Test

class TimingParserTest {
    @Test
    fun defaultsOnGarbage() {
        assertEquals(TimingConfig(), TimingParser.parse("this is not a config :)"))
    }

    @Test
    fun partialInvalidKeepsDefaults() {
        val got = TimingParser.parse(
            """
            idle_fps = 2
            hover_fps = no
            hit_fps = -3
            hit_duration = 2.5
            """.trimIndent(),
        )
        assertEquals(2.0, got.idleFps, 0.0)
        assertEquals(4.0, got.hoverFps, 0.0)
        assertEquals(6.0, got.hitFps, 0.0)
        assertEquals(2500L, got.hitDurationMs)
    }

    @Test
    fun jsonValues() {
        val got = TimingParser.parse("""{"idle_fps": 1.5, "hover_fps": "5", "hit_fps": 1000, "hit_duration": true}""")
        assertEquals(1.5, got.idleFps, 0.0)
        assertEquals(5.0, got.hoverFps, 0.0)
        assertEquals(6.0, got.hitFps, 0.0)
        assertEquals(1350L, got.hitDurationMs)
    }
}
