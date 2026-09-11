package com.fengxuan.mypet.pet

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class PetCatalogTest {
    @Test
    fun discoversNestedAndPrefersOrange() {
        val tree = mapOf(
            "pets" to listOf("pack", "tuft", "orange"),
            "pets/pack" to listOf("cat-1"),
            "pets/pack/cat-1" to listOf("idle"),
            "pets/pack/cat-1/idle" to listOf("000.png"),
            "pets/tuft" to listOf("idle"),
            "pets/tuft/idle" to listOf("000.png"),
            "pets/orange" to listOf("idle", "hover"),
            "pets/orange/idle" to listOf("000.png"),
            "pets/orange/hover" to listOf("000.png"),
        )
        val pets = PetCatalog.discover { tree[it].orEmpty() }
        assertEquals(listOf("orange", "pack/cat-1", "tuft"), pets.map { it.name })
    }

    @Test
    fun otherPetIndicesSkipCurrent() {
        assertEquals(listOf(0, 1, 3, 4), PetCatalog.otherPetIndices(2, 5))
        assertTrue(PetCatalog.otherPetIndices(0, 1).isEmpty())
    }

    @Test
    fun animationFolderNames() {
        val names = listOf("idle", "meow", "itch")
        assertEquals("idle", PetCatalog.animationDirName(PetCatalog.Kind.IDLE, names))
        assertEquals("meow", PetCatalog.animationDirName(PetCatalog.Kind.HOVER, names))
        assertEquals("itch", PetCatalog.animationDirName(PetCatalog.Kind.HIT, names))
    }
}
