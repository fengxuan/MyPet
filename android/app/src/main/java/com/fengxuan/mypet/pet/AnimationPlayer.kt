package com.fengxuan.mypet.pet

class AnimationPlayer(
    private val idleCount: Int,
    private val hoverCount: Int,
    private val hitCount: Int,
    private val timing: TimingConfig,
) {
    enum class Mode { IDLE, HOVER, HIT }

    var hovering: Boolean = false
    private var mode: Mode = Mode.IDLE
    private var hitStartMs: Long = 0L

    fun triggerHit(nowMs: Long) {
        if (hitCount <= 0) return
        mode = Mode.HIT
        hitStartMs = nowMs
    }

    fun current(nowMs: Long): FrameRef {
        if (mode == Mode.HIT) {
            if (nowMs - hitStartMs >= timing.hitDurationMs) {
                mode = if (hovering) Mode.HOVER else Mode.IDLE
            }
        } else {
            mode = if (hovering && hoverCount > 0) Mode.HOVER else Mode.IDLE
        }
        val kind = when (mode) {
            Mode.HIT -> PetCatalog.Kind.HIT
            Mode.HOVER -> PetCatalog.Kind.HOVER
            Mode.IDLE -> PetCatalog.Kind.IDLE
        }
        val count = countFor(kind)
        if (count <= 0) {
            return FrameRef(PetCatalog.Kind.IDLE, 0)
        }
        val fps = fpsFor(kind)
        val elapsed = if (mode == Mode.HIT) nowMs - hitStartMs else nowMs
        var index = ((elapsed / 1000.0) * fps).toInt()
        index = if (mode == Mode.HIT) {
            index.coerceIn(0, count - 1)
        } else {
            Math.floorMod(index, count)
        }
        return FrameRef(kind, index)
    }

    private fun countFor(kind: PetCatalog.Kind): Int = when (kind) {
        PetCatalog.Kind.IDLE -> idleCount
        PetCatalog.Kind.HOVER -> if (hoverCount > 0) hoverCount else idleCount
        PetCatalog.Kind.HIT -> if (hitCount > 0) hitCount else idleCount
    }

    private fun fpsFor(kind: PetCatalog.Kind): Double = when (kind) {
        PetCatalog.Kind.IDLE -> timing.idleFps
        PetCatalog.Kind.HOVER -> timing.hoverFps
        PetCatalog.Kind.HIT -> timing.hitFps
    }

    data class FrameRef(val kind: PetCatalog.Kind, val index: Int)
}
