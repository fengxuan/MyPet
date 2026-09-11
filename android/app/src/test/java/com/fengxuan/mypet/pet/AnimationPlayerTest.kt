package com.fengxuan.mypet.pet

import org.junit.Assert.assertEquals
import org.junit.Test

class AnimationPlayerTest {
    private val timing = TimingConfig(idleFps = 2.0, hoverFps = 4.0, hitFps = 2.0, hitDurationMs = 1000)

    @Test
    fun idleLoops() {
        val player = AnimationPlayer(4, 2, 3, timing)
        val first = player.current(0)
        val later = player.current(2000)
        assertEquals(PetCatalog.Kind.IDLE, first.kind)
        assertEquals(0, first.index)
        assertEquals(PetCatalog.Kind.IDLE, later.kind)
        assertEquals(0, later.index)
    }

    @Test
    fun hoverWhileTouching() {
        val player = AnimationPlayer(4, 2, 3, timing)
        player.hovering = true
        val frame = player.current(0)
        assertEquals(PetCatalog.Kind.HOVER, frame.kind)
    }

    @Test
    fun hitThenReturnsToIdle() {
        val player = AnimationPlayer(4, 2, 3, timing)
        player.triggerHit(0)
        val during = player.current(200)
        player.current(1200)
        val after = player.current(1200)
        assertEquals(PetCatalog.Kind.HIT, during.kind)
        assertEquals(PetCatalog.Kind.IDLE, after.kind)
    }

    @Test
    fun hitUsesLastFrameAfterSequence() {
        val player = AnimationPlayer(4, 2, 3, TimingConfig(hitFps = 10.0, hitDurationMs = 1000))
        player.triggerHit(0)
        val frame = player.current(800)
        assertEquals(PetCatalog.Kind.HIT, frame.kind)
        assertEquals(2, frame.index)
    }
}
