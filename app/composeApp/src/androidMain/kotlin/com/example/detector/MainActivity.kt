package com.example.detector

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.Composable
import androidx.compose.ui.tooling.preview.Preview
import com.shepeliev.webrtckmp.WebRtc
import org.webrtc.Logging

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge()
        super.onCreate(savedInstanceState)

        val initializationOptionsBuilder = WebRtc.createInitializationOptionsBuilder()
//            .setInjectableLogger(WebRtcLogger, Logging.Severity.LS_ERROR)
        val peerConnectionFactoryBuilder = WebRtc.createPeerConnectionFactoryBuilder(initializationOptionsBuilder = initializationOptionsBuilder)
        WebRtc.configure(peerConnectionFactoryBuilder = peerConnectionFactoryBuilder)

        setContent {
            App()
        }
    }
}

@Preview
@Composable
fun AppAndroidPreview() {
    App()
}