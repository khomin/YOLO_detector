package com.example.detector

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.viewinterop.AndroidView
import com.shepeliev.webrtckmp.VideoTrack
import org.webrtc.EglBase
import org.webrtc.RendererCommon
import org.webrtc.SurfaceViewRenderer

@Composable
actual fun VideoRenderer(track: VideoTrack?, modifier: Modifier) {
    if (track == null) return

    AndroidView(
        factory = { context ->
            SurfaceViewRenderer(context).apply {
                // Use the EglBase from the webrtc-kmp library if available,
                // or create a new one for the surface.
                val eglContext = EglBase.create().eglBaseContext
                init(eglContext, null)
                setScalingType(RendererCommon.ScalingType.SCALE_ASPECT_FIT)
                setEnableHardwareScaler(true)

                // This connects the WebRTC track to the Android Surface
                track.addSink(this)
            }
        },
        modifier = modifier
    )
}