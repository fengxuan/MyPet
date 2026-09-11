package com.fengxuan.mypet.pet

import android.content.Context
import android.content.SharedPreferences
import android.graphics.Bitmap
import android.graphics.BitmapFactory

class PetRepository(context: Context) {
    private val assets = context.assets
    private val prefs: SharedPreferences =
        context.getSharedPreferences("mypet", Context.MODE_PRIVATE)

    val timing: TimingConfig by lazy {
        val text = runCatching { assets.open("config.txt").bufferedReader().readText() }.getOrDefault("")
        TimingParser.parse(text)
    }

    fun allPets(): List<PetInfo> = PetCatalog.discover { path ->
        assets.list(path)?.toList().orEmpty()
    }

    fun savedName(): String? = prefs.getString(KEY_PET, null)

    fun saveName(name: String) {
        prefs.edit().putString(KEY_PET, name).apply()
    }

    fun indexOf(name: String?, pets: List<PetInfo>): Int {
        if (name.isNullOrBlank() || pets.isEmpty()) return 0
        val found = pets.indexOfFirst { it.name == name }
        return if (found >= 0) found else 0
    }

    fun loadFrames(pet: PetInfo, maxSize: Int = 256): PetFrames {
        val idle = loadKind(pet.dir, PetCatalog.Kind.IDLE, maxSize)
        val hover = loadKind(pet.dir, PetCatalog.Kind.HOVER, maxSize).ifEmpty { idle }
        val hit = loadKind(pet.dir, PetCatalog.Kind.HIT, maxSize).ifEmpty { idle }
        return PetFrames(idle, hover, hit)
    }

    fun loadThumb(pet: PetInfo, maxSize: Int = 96): Bitmap? {
        val frames = loadKind(pet.dir, PetCatalog.Kind.IDLE, maxSize)
        return frames.firstOrNull() ?: loadKind(pet.dir, PetCatalog.Kind.HOVER, maxSize).firstOrNull()
    }

    private fun loadKind(dir: String, kind: PetCatalog.Kind, maxSize: Int): List<Bitmap> {
        val children = assets.list(dir)?.toList().orEmpty()
        val sub = PetCatalog.animationDirName(kind, children)
        val folder = if (sub != null) "$dir/$sub" else dir
        val names = assets.list(folder)?.filter { PetCatalog.isPng(it) }?.sorted().orEmpty()
        if (kind != PetCatalog.Kind.IDLE && sub == null) {
            return emptyList()
        }
        return names.mapNotNull { decode("$folder/$it", maxSize) }
    }

    private fun decode(path: String, maxSize: Int): Bitmap? {
        val bounds = BitmapFactory.Options().apply { inJustDecodeBounds = true }
        runCatching { assets.open(path).use { BitmapFactory.decodeStream(it, null, bounds) } }
        var sample = 1
        val w = bounds.outWidth
        val h = bounds.outHeight
        while (w / sample > maxSize || h / sample > maxSize) {
            sample *= 2
        }
        val opts = BitmapFactory.Options().apply {
            inSampleSize = sample
            inPreferredConfig = Bitmap.Config.ARGB_8888
        }
        return runCatching {
            assets.open(path).use { BitmapFactory.decodeStream(it, null, opts) }
        }.getOrNull()
    }

    companion object {
        private const val KEY_PET = "current_pet"
    }
}

data class PetFrames(
    val idle: List<Bitmap>,
    val hover: List<Bitmap>,
    val hit: List<Bitmap>,
) {
    fun recycle() {
        (idle + hover + hit).distinct().forEach { if (!it.isRecycled) it.recycle() }
    }
}
