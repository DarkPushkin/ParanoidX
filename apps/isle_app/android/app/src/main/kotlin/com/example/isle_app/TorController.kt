package com.example.isle_app

import android.content.Context
import android.util.Log
import java.io.File
import java.io.FileOutputStream

class TorController(private val context: Context, private val dataDir: File) {
    companion object {
        private const val TAG = "ParanoidX:Tor"
        const val SOCKS_PORT = 9050
        private const val CONTROL_PORT = 9051
        private const val BOOT_TIMEOUT_MS = 60000L
    }

    private var process: Process? = null
    private var logThread: Thread? = null

    fun start() {
        if (process?.isAlive == true) return
        Log.i(TAG, "Starting Tor daemon...")

        val torDir = File(dataDir, "tor").apply { mkdirs() }
        val torBinary = extractTorBinary(torDir)
        if (torBinary == null) {
            Log.e(TAG, "FATAL: Could not extract Tor binary")
            throw Exception("Tor binary extraction failed")
        }

        val torrc = File(torDir, "torrc").apply {
            writeText("""
SocksPort 127.0.0.1:$SOCKS_PORT
ControlPort $CONTROL_PORT
DataDirectory ${torDir.absolutePath}/data
AvoidDiskWrites 1
Log notice stdout
            """.trimIndent())
        }
        File(torDir, "data").mkdirs()

        val pb = ProcessBuilder(
            torBinary.absolutePath, "-f", torrc.absolutePath
        )
        pb.environment()["HOME"] = torDir.absolutePath
        pb.environment()["LD_LIBRARY_PATH"] = torBinary.parentFile?.absolutePath ?: ""
        pb.redirectErrorStream(true)

        process = pb.start()
        Log.i(TAG, "Tor PID: ${process?.pid()}")

        logThread = Thread { readLogs() }.apply { start() }
    }

    fun stop() {
        process?.let { p ->
            Log.i(TAG, "Stopping Tor")
            p.destroy()
            p.waitFor(5, java.util.concurrent.TimeUnit.SECONDS)
            if (p.isAlive) p.destroyForcibly()
        }
        process = null
    }

    fun isRunning(): Boolean = process?.isAlive == true

    private fun readLogs() {
        try {
            process?.inputStream?.bufferedReader()?.use { reader ->
                var line = reader.readLine()
                while (line != null) {
                    if (line.contains("Bootstrapped 100")) {
                        Log.i(TAG, "Tor bootstrapped 100% — ready")
                    }
                    line = reader.readLine()
                }
            }
        } catch (_: Exception) {}
    }

    private fun extractTorBinary(torDir: File): File? {
        val nativeLibDir = context.applicationInfo.nativeLibraryDir
        val libTor = File(nativeLibDir, "libtor.so")
        if (!libTor.exists()) {
            Log.w(TAG, "libtor.so not found in $nativeLibDir")
            return null
        }
        val target = File(torDir, "tor")
        if (!target.exists() || libTor.lastModified() > target.lastModified()) {
            Log.i(TAG, "Copying libtor.so (${libTor.length()} bytes) -> $target")
            libTor.inputStream().use { input ->
                FileOutputStream(target).use { output ->
                    input.copyTo(output)
                }
            }
        }
        target.setExecutable(true, true)
        return if (target.exists() && target.canExecute()) target else null
    }
}
