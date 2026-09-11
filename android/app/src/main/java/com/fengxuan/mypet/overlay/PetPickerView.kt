package com.fengxuan.mypet.overlay

import android.content.Context
import android.graphics.Bitmap
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.FrameLayout
import android.widget.ImageView
import android.widget.TextView
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import com.fengxuan.mypet.R
import com.fengxuan.mypet.pet.PetCatalog
import com.fengxuan.mypet.pet.PetInfo

class PetPickerView(context: Context) : FrameLayout(context) {
    fun interface Listener {
        fun onPetChosen(index: Int)
    }

    private val recycler: RecyclerView
    private val adapter = Adapter()
    var listener: Listener? = null

    init {
        val view = LayoutInflater.from(context).inflate(R.layout.view_pet_picker, this, true)
        recycler = view.findViewById(R.id.petList)
        recycler.layoutManager = LinearLayoutManager(context)
        recycler.adapter = adapter
        val itemHeight = (56 * resources.displayMetrics.density).toInt()
        recycler.layoutParams = recycler.layoutParams.apply {
            height = itemHeight * 5
        }
    }

    fun setPets(pets: List<PetInfo>, currentIndex: Int, thumbs: Map<String, Bitmap?>) {
        val others = PetCatalog.otherPetIndices(currentIndex, pets.size).map { pets[it] to it }
        adapter.submit(others, thumbs)
        recycler.scrollToPosition(0)
    }

    private inner class Adapter : RecyclerView.Adapter<Holder>() {
        private var items: List<Pair<PetInfo, Int>> = emptyList()
        private var thumbs: Map<String, Bitmap?> = emptyMap()

        fun submit(items: List<Pair<PetInfo, Int>>, thumbs: Map<String, Bitmap?>) {
            this.items = items
            this.thumbs = thumbs
            notifyDataSetChanged()
        }

        override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): Holder {
            val view = LayoutInflater.from(parent.context).inflate(R.layout.item_pet_picker, parent, false)
            return Holder(view)
        }

        override fun onBindViewHolder(holder: Holder, position: Int) {
            val (pet, index) = items[position]
            holder.name.text = pet.name.replace('/', ' ')
            holder.thumb.setImageBitmap(thumbs[pet.name])
            holder.itemView.setOnClickListener { listener?.onPetChosen(index) }
        }

        override fun getItemCount(): Int = items.size
    }

    private class Holder(view: View) : RecyclerView.ViewHolder(view) {
        val thumb: ImageView = view.findViewById(R.id.thumb)
        val name: TextView = view.findViewById(R.id.name)
    }
}
