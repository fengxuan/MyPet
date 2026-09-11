package com.fengxuan.mypet.pet

data class PetInfo(
    val name: String,
    val dir: String,
)

object PetCatalog {
    private val idleKeys = setOf("idle", "stand", "sleep", "sleeping", "walk", "sitting", "laying")
    private val hoverKeys = setOf("hover", "meow", "alert", "look")
    private val hitKeys = setOf("hit", "itch", "lick", "licking", "slap", "attack", "action")

    fun discover(list: (String) -> List<String>): List<PetInfo> {
        val pets = scan("pets", "", list)
        return pets.sortedWith(
            compareBy<PetInfo> { it.name != "orange" }.thenBy { it.name },
        )
    }

    fun isPng(name: String): Boolean = name.endsWith(".png", ignoreCase = true)

    fun animationDirName(kind: Kind, names: List<String>): String? {
        val keys = when (kind) {
            Kind.IDLE -> idleKeys
            Kind.HOVER -> hoverKeys
            Kind.HIT -> hitKeys
        }
        return names.firstOrNull { keys.contains(it.lowercase()) }
    }

    fun otherPetIndices(current: Int, total: Int): List<Int> {
        if (total <= 0) return emptyList()
        return (0 until total).filter { it != current }
    }

    private fun scan(dir: String, prefix: String, list: (String) -> List<String>): List<PetInfo> {
        val names = list(dir).filter { !it.startsWith(".") }.sorted()
        val pets = mutableListOf<PetInfo>()
        for (name in names) {
            val child = "$dir/$name"
            val display = if (prefix.isEmpty()) name else "$prefix/$name"
            val children = list(child)
            if (isPetDir(children, child, list)) {
                pets += PetInfo(display, child)
            } else if (children.isNotEmpty() && children.none { isPng(it) }) {
                pets += scan(child, display, list)
            }
        }
        return pets
    }

    private fun isPetDir(children: List<String>, dir: String, list: (String) -> List<String>): Boolean {
        if (children.any { isPng(it) }) return true
        val sub = children.firstOrNull { idleKeys.contains(it.lowercase()) } ?: return false
        return list("$dir/$sub").any { isPng(it) }
    }

    enum class Kind { IDLE, HOVER, HIT }
}
