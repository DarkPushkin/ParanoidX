package com.example.isle_app

import android.os.Bundle
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine

class MainActivity : FlutterActivity() {
    private var paranoidx: ParanoidXController? = null

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        paranoidx = ParanoidXController(this, flutterEngine)
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        paranoidx?.start()
    }

    override fun onDestroy() {
        paranoidx?.stop()
        super.onDestroy()
    }
}
