package com.fengxuan.mypet

import android.content.Intent
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.provider.Settings
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import com.google.android.material.button.MaterialButton

class MainActivity : AppCompatActivity() {
    private lateinit var status: TextView
    private lateinit var grant: MaterialButton
    private lateinit var start: MaterialButton
    private lateinit var stop: MaterialButton

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)
        status = findViewById(R.id.status)
        grant = findViewById(R.id.grantOverlay)
        start = findViewById(R.id.startPet)
        stop = findViewById(R.id.stopPet)

        grant.setOnClickListener { requestOverlayPermission() }
        start.setOnClickListener {
            if (!canDrawOverlays()) {
                requestOverlayPermission()
                return@setOnClickListener
            }
            requestNotificationPermission()
            PetService.start(this)
            refreshStatus()
        }
        stop.setOnClickListener {
            PetService.stop(this)
            refreshStatus()
        }
    }

    override fun onResume() {
        super.onResume()
        refreshStatus()
    }

    private fun refreshStatus() {
        val permitted = canDrawOverlays()
        grant.isEnabled = !permitted
        start.isEnabled = permitted
        status.setText(
            when {
                !permitted -> R.string.status_need_permission
                PetService.running.get() -> R.string.status_running
                else -> R.string.status_stopped
            },
        )
    }

    private fun canDrawOverlays(): Boolean {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            Settings.canDrawOverlays(this)
        } else {
            true
        }
    }

    private fun requestOverlayPermission() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.M) return
        val intent = Intent(
            Settings.ACTION_MANAGE_OVERLAY_PERMISSION,
            Uri.parse("package:$packageName"),
        )
        startActivity(intent)
    }

    private fun requestNotificationPermission() {
        if (Build.VERSION.SDK_INT >= 33) {
            ActivityCompat.requestPermissions(
                this,
                arrayOf(android.Manifest.permission.POST_NOTIFICATIONS),
                100,
            )
        }
    }
}
