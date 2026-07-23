package com.example.isle_app

import android.content.Context
import android.util.Log
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import java.io.File

class ParanoidXController(private val context: Context, private val engine: FlutterEngine) {
    companion object {
        private const val TAG = "ParanoidX"
        const val CHANNEL = "com.example.isle_app/paranoidx"
        const val TOR_SOCKS_PORT = 9050
        const val V2RAY_SOCKS_PORT = 10810
        const val V2RAY_HTTP_PORT = 10811
    }

    private var tor: TorController? = null
    private var v2ray: V2RayController? = null
    private var channel: MethodChannel? = null
    private var isRunning = false

    fun start() {
        if (isRunning) return
        Log.i(TAG, "ParanoidX bridge starting...")

        channel = MethodChannel(engine.dartExecutor.binaryMessenger, CHANNEL)

        val dataDir = File(context.filesDir, "paranoidx").apply { mkdirs() }

        tor = TorController(context, dataDir)
        v2ray = V2RayController(context, dataDir)

        try {
            Log.i(TAG, "Starting Tor daemon...")
            tor?.start()
            val torReady = waitForPort(TOR_SOCKS_PORT, 30000)
            sendStatus("tor", torReady, if (torReady) "connected" else "timeout")

            if (torReady) {
                Log.i(TAG, "Starting V2Ray daemon...")
                v2ray?.start()
                val v2rayReady = waitForPort(V2RAY_SOCKS_PORT, 15000)
                sendStatus("v2ray", v2rayReady, if (v2rayReady) "connected" else "timeout")
            }

            isRunning = torReady
            sendStatus("paranoidx", isRunning,
                if (isRunning) "V2Ray+Tor ready" else "Tor failed to start"
            )
            Log.i(TAG, "ParanoidX bridge: ${if (isRunning) "READY" else "FAILED"}")
        } catch (e: Exception) {
            Log.e(TAG, "ParanoidX start failed", e)
            sendStatus("paranoidx", false, e.message ?: "unknown error")
            stop()
        }
    }

    fun stop() {
        Log.i(TAG, "Stopping ParanoidX bridge...")
        v2ray?.stop()
        tor?.stop()
        isRunning = false
        sendStatus("paranoidx", false, "stopped")
    }

    private fun sendStatus(layer: String, healthy: Boolean, message: String) {
        channel?.invokeMethod("status", mapOf(
            "layer" to layer,
            "healthy" to healthy,
            "message" to message
        ))
    }

    private fun waitForPort(port: Int, timeoutMs: Long): Boolean {
        val deadline = System.currentTimeMillis() + timeoutMs
        while (System.currentTimeMillis() < deadline) {
            try {
                val s = java.net.Socket()
                s.connect(java.net.InetSocketAddress("127.0.0.1", port), 500)
                s.close()
                return true
            } catch (_: Exception) {
                Thread.sleep(200)
            }
        }
        return false
    }
}
