package com.example.isle_app

import android.content.Context
import android.util.Log
import java.io.File

class V2RayController(private val context: Context, private val dataDir: File) {
    companion object {
        private const val TAG = "ParanoidX:V2Ray"
        const val SOCKS_PORT = 10810
    }

    private var process: Process? = null

    fun start() {
        if (process?.isAlive == true) return
        Log.i(TAG, "Starting Xray daemon...")

        val v2rayDir = File(dataDir, "v2ray").apply { mkdirs() }
        val xrayBin = extractXray(v2rayDir) ?: run {
            Log.e(TAG, "FATAL: Xray binary not found in assets")
            throw Exception("Xray binary missing — place bin/xray-arm64 in assets/")
        }

        writeConfig(v2rayDir)

        val pb = ProcessBuilder(
            xrayBin.absolutePath, "run", "-c",
            File(v2rayDir, "config.json").absolutePath
        )
        pb.directory(v2rayDir)
        pb.redirectErrorStream(true)

        process = pb.start()
        Log.i(TAG, "Xray PID: ${process?.pid()}")
    }

    fun stop() {
        process?.let { p ->
            Log.i(TAG, "Stopping Xray")
            p.destroy()
            p.waitFor(3, java.util.concurrent.TimeUnit.SECONDS)
            if (p.isAlive) p.destroyForcibly()
        }
        process = null
    }

    fun isRunning(): Boolean = process?.isAlive == true

    private fun extractXray(dir: File): File? {
        val target = File(dir, "xray")
        if (target.exists()) return target

        try {
            context.assets.open("bin/xray-arm64").use { input ->
                target.outputStream().use { output ->
                    input.copyTo(output)
                }
            }
            target.setExecutable(true, true)
            Log.i(TAG, "Extracted xray-arm64 from assets")
            return target
        } catch (e: Exception) {
            Log.e(TAG, "Failed to extract xray from assets", e)
            return null
        }
    }

    private fun writeConfig(dir: File) {
        val config = """
{
  "log": {"loglevel": "warning"},
  "inbounds": [
    {
      "port": $SOCKS_PORT,
      "protocol": "socks",
      "settings": {"udp": true, "ip": "127.0.0.1"},
      "tag": "socks-in"
    }
  ],
  "outbounds": [
    {
      "protocol": "socks",
      "settings": {
        "servers": [{"address": "127.0.0.1", "port": ${TorController.SOCKS_PORT}}]
      },
      "tag": "tor-out"
    }
  ]
}
""".trimIndent()
        File(dir, "config.json").writeText(config)
        Log.i(TAG, "Xray config: inbound :$SOCKS_PORT -> outbound -> Tor :${TorController.SOCKS_PORT}")
    }
}
