package com.fengxuan.mypet

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.os.Build
import android.os.IBinder
import android.provider.Settings
import java.util.concurrent.atomic.AtomicBoolean
import androidx.core.app.NotificationCompat
import com.fengxuan.mypet.overlay.OverlayController
import com.fengxuan.mypet.pet.PetRepository

class PetService : Service() {
    private var overlay: OverlayController? = null

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onCreate() {
        super.onCreate()
        createChannel()
        startForeground(NOTIFICATION_ID, buildNotification())
        if (!canDrawOverlays()) {
            stopSelf()
            return
        }
        val repository = PetRepository(this)
        val controller = OverlayController(this, repository)
        if (!controller.show()) {
            stopSelf()
            return
        }
        overlay = controller
        running.set(true)
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        if (intent?.action == ACTION_STOP) {
            stopSelf()
            return START_NOT_STICKY
        }
        return START_STICKY
    }

    override fun onDestroy() {
        overlay?.hide()
        overlay = null
        running.set(false)
        super.onDestroy()
    }

    private fun canDrawOverlays(): Boolean {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            Settings.canDrawOverlays(this)
        } else {
            true
        }
    }

    private fun createChannel() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return
        val manager = getSystemService(NotificationManager::class.java)
        val channel = NotificationChannel(
            CHANNEL_ID,
            getString(R.string.notification_channel),
            NotificationManager.IMPORTANCE_LOW,
        )
        manager.createNotificationChannel(channel)
    }

    private fun buildNotification(): Notification {
        val launch = PendingIntent.getActivity(
            this,
            0,
            Intent(this, MainActivity::class.java),
            PendingIntent.FLAG_UPDATE_CURRENT or immutableFlag(),
        )
        val stop = PendingIntent.getService(
            this,
            1,
            Intent(this, PetService::class.java).setAction(ACTION_STOP),
            PendingIntent.FLAG_UPDATE_CURRENT or immutableFlag(),
        )
        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_stat_pet)
            .setContentTitle(getString(R.string.notification_title))
            .setContentText(getString(R.string.notification_text))
            .setContentIntent(launch)
            .addAction(0, getString(R.string.stop_pet), stop)
            .setOngoing(true)
            .build()
    }

    private fun immutableFlag(): Int {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) PendingIntent.FLAG_IMMUTABLE else 0
    }

    companion object {
        const val ACTION_STOP = "com.fengxuan.mypet.STOP"
        val running = AtomicBoolean(false)
        private const val CHANNEL_ID = "pet_overlay"
        private const val NOTIFICATION_ID = 7

        fun start(context: Context) {
            val intent = Intent(context, PetService::class.java)
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                context.startForegroundService(intent)
            } else {
                context.startService(intent)
            }
        }

        fun stop(context: Context) {
            context.stopService(Intent(context, PetService::class.java))
        }
    }
}
