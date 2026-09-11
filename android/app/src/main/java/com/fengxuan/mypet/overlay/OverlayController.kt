package com.fengxuan.mypet.overlay

import android.content.Context
import android.graphics.PixelFormat
import android.os.Build
import android.view.Gravity
import android.view.WindowManager
import com.fengxuan.mypet.pet.PetFrames
import com.fengxuan.mypet.pet.PetInfo
import com.fengxuan.mypet.pet.PetRepository

/**
 * Service
 *   ↓
 * WindowManager
 *   ↓
 * 悬浮窗口
 *   ↓
 * PetView
 */
class OverlayController(
    context: Context,
    private val repository: PetRepository,
) {
    private val appContext = context.applicationContext
    private val windowManager = appContext.getSystemService(Context.WINDOW_SERVICE) as WindowManager
    private var petView: PetView? = null
    private var pickerView: PetPickerView? = null
    private var petParams: WindowManager.LayoutParams? = null
    private var pickerParams: WindowManager.LayoutParams? = null
    private var pets: List<PetInfo> = emptyList()
    private var petIndex = 0
    private var frames: PetFrames? = null
    private var attached = false

    fun show(): Boolean {
        if (attached) return true
        pets = repository.allPets()
        if (pets.isEmpty()) return false
        petIndex = repository.indexOf(repository.savedName(), pets)
        val view = PetView(appContext)
        view.callbacks = object : PetView.Callbacks {
            override fun onDrag(dx: Int, dy: Int) = movePet(dx, dy)
            override fun onTap() {}
            override fun onLongPress() = togglePicker()
        }
        val params = baseParams().apply {
            width = WindowManager.LayoutParams.WRAP_CONTENT
            height = WindowManager.LayoutParams.WRAP_CONTENT
            gravity = Gravity.TOP or Gravity.START
            x = 80
            y = 400
        }
        windowManager.addView(view, params)
        petView = view
        petParams = params
        attached = true
        loadCurrent()
        return true
    }

    fun hide() {
        hidePicker()
        petView?.let { runCatching { windowManager.removeView(it) } }
        petView = null
        petParams = null
        frames?.recycle()
        frames = null
        attached = false
    }

    private fun loadCurrent() {
        val pet = pets.getOrNull(petIndex) ?: return
        frames?.recycle()
        val loaded = repository.loadFrames(pet)
        frames = loaded
        petView?.setPet(loaded, repository.timing)
        repository.saveName(pet.name)
    }

    private fun movePet(dx: Int, dy: Int) {
        val view = petView ?: return
        val params = petParams ?: return
        params.x += dx
        params.y += dy
        windowManager.updateViewLayout(view, params)
        pickerView?.let { picker ->
            val pp = pickerParams ?: return@let
            pp.x = params.x
            pp.y = params.y + view.height + 12
            windowManager.updateViewLayout(picker, pp)
        }
    }

    private fun togglePicker() {
        if (pickerView != null) {
            hidePicker()
        } else {
            showPicker()
        }
    }

    private fun showPicker() {
        if (pickerView != null || pets.size < 2) return
        val pet = petView ?: return
        val origin = petParams ?: return
        val picker = PetPickerView(appContext)
        val thumbs = pets.associate { it.name to repository.loadThumb(it) }
        picker.setPets(pets, petIndex, thumbs)
        picker.listener = PetPickerView.Listener { index ->
            petIndex = index
            loadCurrent()
            hidePicker()
        }
        val params = baseParams().apply {
            width = WindowManager.LayoutParams.WRAP_CONTENT
            height = WindowManager.LayoutParams.WRAP_CONTENT
            gravity = Gravity.TOP or Gravity.START
            x = origin.x
            y = origin.y + pet.height + 12
            flags = flags or WindowManager.LayoutParams.FLAG_WATCH_OUTSIDE_TOUCH
        }
        picker.setOnTouchListener { _, event ->
            if (event.action == android.view.MotionEvent.ACTION_OUTSIDE) {
                hidePicker()
                true
            } else {
                false
            }
        }
        windowManager.addView(picker, params)
        pickerView = picker
        pickerParams = params
    }

    private fun hidePicker() {
        pickerView?.let { runCatching { windowManager.removeView(it) } }
        pickerView = null
        pickerParams = null
    }

    private fun baseParams(): WindowManager.LayoutParams {
        val type = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY
        } else {
            @Suppress("DEPRECATION")
            WindowManager.LayoutParams.TYPE_PHONE
        }
        return WindowManager.LayoutParams(
            WindowManager.LayoutParams.WRAP_CONTENT,
            WindowManager.LayoutParams.WRAP_CONTENT,
            type,
            WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE or
                WindowManager.LayoutParams.FLAG_LAYOUT_NO_LIMITS or
                WindowManager.LayoutParams.FLAG_HARDWARE_ACCELERATED,
            PixelFormat.TRANSLUCENT,
        )
    }
}
