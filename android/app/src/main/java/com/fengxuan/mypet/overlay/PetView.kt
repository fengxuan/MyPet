package com.fengxuan.mypet.overlay

import android.content.Context
import android.graphics.Canvas
import android.graphics.Paint
import android.graphics.Rect
import android.os.SystemClock
import android.view.Choreographer
import android.view.MotionEvent
import android.view.View
import android.view.ViewConfiguration
import com.fengxuan.mypet.pet.AnimationPlayer
import com.fengxuan.mypet.pet.PetCatalog
import com.fengxuan.mypet.pet.PetFrames
import com.fengxuan.mypet.pet.TimingConfig
import kotlin.math.abs
import kotlin.math.roundToInt

class PetView(context: Context) : View(context), Choreographer.FrameCallback {
    interface Callbacks {
        fun onDrag(dx: Int, dy: Int)
        fun onTap()
        fun onLongPress()
    }

    var callbacks: Callbacks? = null
    private var frames: PetFrames? = null
    private var player: AnimationPlayer? = null
    private val paint = Paint(Paint.FILTER_BITMAP_FLAG or Paint.ANTI_ALIAS_FLAG)
    private val dest = Rect()
    private val choreographer = Choreographer.getInstance()
    private var running = false

    private val touchSlop = ViewConfiguration.get(context).scaledTouchSlop
    private val longPressTimeout = ViewConfiguration.getLongPressTimeout().toLong()
    private var downX = 0f
    private var downY = 0f
    private var lastX = 0f
    private var lastY = 0f
    private var dragging = false
    private var longPressPosted = false
    private val longPressRunnable = Runnable {
        longPressPosted = false
        player?.hovering = false
        callbacks?.onLongPress()
    }

    fun setPet(frames: PetFrames, timing: TimingConfig) {
        this.frames?.recycle()
        this.frames = frames
        player = AnimationPlayer(frames.idle.size, frames.hover.size, frames.hit.size, timing)
        invalidate()
    }

    fun triggerHit() {
        player?.triggerHit(SystemClock.uptimeMillis())
    }

    override fun onMeasure(widthMeasureSpec: Int, heightMeasureSpec: Int) {
        val size = (200 * resources.displayMetrics.density).roundToInt()
        setMeasuredDimension(size, size)
    }

    override fun onAttachedToWindow() {
        super.onAttachedToWindow()
        running = true
        choreographer.postFrameCallback(this)
    }

    override fun onDetachedFromWindow() {
        running = false
        choreographer.removeFrameCallback(this)
        removeCallbacks(longPressRunnable)
        super.onDetachedFromWindow()
    }

    override fun doFrame(frameTimeNanos: Long) {
        if (!running) return
        invalidate()
        choreographer.postFrameCallback(this)
    }

    override fun onDraw(canvas: Canvas) {
        val pet = frames ?: return
        val ref = player?.current(SystemClock.uptimeMillis()) ?: return
        val list = when (ref.kind) {
            PetCatalog.Kind.IDLE -> pet.idle
            PetCatalog.Kind.HOVER -> pet.hover
            PetCatalog.Kind.HIT -> pet.hit
        }
        if (list.isEmpty()) return
        val bmp = list[ref.index.coerceIn(0, list.lastIndex)]
        dest.set(0, 0, width, height)
        val filter = bmp.width <= 64 && bmp.height <= 64
        paint.isFilterBitmap = !filter
        canvas.drawBitmap(bmp, null, dest, paint)
    }

    override fun onTouchEvent(event: MotionEvent): Boolean {
        when (event.actionMasked) {
            MotionEvent.ACTION_DOWN -> {
                downX = event.rawX
                downY = event.rawY
                lastX = event.rawX
                lastY = event.rawY
                dragging = false
                player?.hovering = true
                longPressPosted = true
                postDelayed(longPressRunnable, longPressTimeout)
                parent?.requestDisallowInterceptTouchEvent(true)
                return true
            }
            MotionEvent.ACTION_MOVE -> {
                val dx = event.rawX - downX
                val dy = event.rawY - downY
                if (!dragging && (abs(dx) > touchSlop || abs(dy) > touchSlop)) {
                    dragging = true
                    cancelLongPressWait()
                }
                if (dragging) {
                    callbacks?.onDrag((event.rawX - lastX).roundToInt(), (event.rawY - lastY).roundToInt())
                    lastX = event.rawX
                    lastY = event.rawY
                }
                return true
            }
            MotionEvent.ACTION_UP, MotionEvent.ACTION_CANCEL -> {
                val wasDragging = dragging
                cancelLongPressWait()
                player?.hovering = false
                dragging = false
                if (event.actionMasked == MotionEvent.ACTION_UP && !wasDragging) {
                    player?.triggerHit(SystemClock.uptimeMillis())
                    callbacks?.onTap()
                }
                return true
            }
        }
        return super.onTouchEvent(event)
    }

    private fun cancelLongPressWait() {
        if (longPressPosted) {
            removeCallbacks(longPressRunnable)
            longPressPosted = false
        }
    }
}
