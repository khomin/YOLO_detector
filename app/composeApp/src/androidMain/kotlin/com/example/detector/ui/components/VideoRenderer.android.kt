package com.example.detector.ui.components

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.viewinterop.AndroidView
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import com.shepeliev.webrtckmp.VideoTrack
import com.shepeliev.webrtckmp.WebRtc
import org.webrtc.RendererCommon
import org.webrtc.SurfaceViewRenderer

@Composable
actual fun VideoRenderer(track: VideoTrack?, modifier: Modifier) {
    if (track == null) return

    // We store the renderer in a state so the lifecycle observer can find it
    var renderer by remember { mutableStateOf<SurfaceViewRenderer?>(null) }

    val lifecycle = LocalLifecycleOwner.current.lifecycle

    DisposableEffect(lifecycle, track) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_RESUME -> {
                    renderer?.apply {
                        init(WebRtc.rootEglBase.eglBaseContext, null)
                        track.addSink(this)
                    }
                }
                Lifecycle.Event.ON_PAUSE -> {
                    renderer?.let {
                        track.removeSink(it)
                        it.release()
                    }
                }
                else -> {}
            }
        }
        lifecycle.addObserver(observer)
        onDispose {
            renderer?.let { track.removeSink(it) }
            lifecycle.removeObserver(observer)
        }
    }
    AndroidView(
        factory = { context ->
            SurfaceViewRenderer(context).apply {
                setScalingType(RendererCommon.ScalingType.SCALE_ASPECT_FIT)
                setEnableHardwareScaler(true)
                renderer = this
            }
        },
        modifier = modifier.fillMaxSize()
    )
}