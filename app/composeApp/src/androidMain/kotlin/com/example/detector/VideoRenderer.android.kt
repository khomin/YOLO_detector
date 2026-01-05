package com.example.detector

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalLifecycleOwner
import androidx.compose.ui.viewinterop.AndroidView
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import com.shepeliev.webrtckmp.AudioTrack
import com.shepeliev.webrtckmp.VideoTrack
import com.shepeliev.webrtckmp.WebRtc
import org.webrtc.RendererCommon
import org.webrtc.SurfaceViewRenderer
import org.webrtc.VideoSink

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
                        // CRITICAL: Must use the library's rootEglBase
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

//@Composable
//actual fun VideoRenderer(track: VideoTrack?, modifier: Modifier) {
//    val context = LocalContext.current
//
//    // 1. Create and remember the view so it's NOT recreated
//    val renderer = remember {
//        SurfaceViewRenderer(context).apply {
//            init(WebRtc.rootEglBase.eglBaseContext, null)
//            setScalingType(RendererCommon.ScalingType.SCALE_ASPECT_FIT)
//            setEnableHardwareScaler(true)
//            setZOrderMediaOverlay(true)
//        }
//    }
//
//    // 2. Handle the Track attachment/detachment
//    DisposableEffect(track) {
//        track?.addSink(renderer)
//        onDispose {
//            track?.removeSink(renderer)
//        }
//    }
//
//    // 3. Place the view in the UI
//    AndroidView(
//        factory = { renderer },
//        modifier = modifier.fillMaxSize()
//    )
//}

//@Composable
//actual fun VideoRenderer(track: VideoTrack?, modifier: Modifier) {
//    if (track == null) return
//
//    // FORCE a size and a border so you can see where the view actually is
//    Box(modifier = modifier.size(300.dp).border(2.dp, Color.Red)) {
//        AndroidView(
//            modifier = modifier,
//            factory = { context ->
//                SurfaceViewRenderer(context).apply {
//                    init(rootEglBase.eglBaseContext, null)
//                    setZOrderOnTop(true) // Try this instead of MediaOverlay
//                    setScalingType(RendererCommon.ScalingType.SCALE_ASPECT_FILL)
//                }
//            },
//            update = { view ->
//                track.addSink(view)
//            }
//        )
//    }
//}

//@Composable
//actual fun VideoRenderer(track: VideoTrack?, modifier: Modifier) {
//    if (track == null) return
//    AndroidView(
//        modifier = modifier,
//        factory = { context ->
//            SurfaceViewRenderer(context).apply {
//                // Use the shared global context
//                init(globalEglBase.eglBaseContext, null)
//                setScalingType(RendererCommon.ScalingType.SCALE_ASPECT_FIT)
//                setEnableHardwareScaler(true)
//                setZOrderMediaOverlay(true)
//            }
//        },
//        update = { view ->
//            track.addSink(view)
//        },
//        onRelease = { view ->
//            // CRITICAL: Clean up
//            track.removeSink(view)
//            view.release()
//        }
//    )
//}

//@Composable
//actual fun VideoRenderer(track: VideoTrack?, modifier: Modifier) {
//    if (track == null) return
//    AndroidView(
//        factory = { context ->
//            SurfaceViewRenderer(context).apply {
//                val eglContext = EglBase.create().eglBaseContext
//                init(eglContext, null)
//                setEnableHardwareScaler(true)
//                setZOrderMediaOverlay(true)
//            }
//        },
//        update = { view ->
//            track.addSink(view)
//        }
//    )
////    AndroidView(
////        factory = { context ->
////            SurfaceViewRenderer(context).apply {
////                // Use the EglBase from the webrtc-kmp library if available,
////                // or create a new one for the surface.
////                val eglContext = EglBase.create().eglBaseContext
////                init(eglContext, null)
////                setScalingType(RendererCommon.ScalingType.SCALE_ASPECT_FIT)
////                setEnableHardwareScaler(true)
////
////                // This connects the WebRTC track to the Android Surface
////                track.addSink(this)
////            }
////        },
////        modifier = modifier
////    )
//}

//@Composable
//actual fun VideoRenderer(track: VideoTrack?, modifier: Modifier) {
//    if (track == null) return
//    val renderer = remember {
//        SurfaceViewRenderer(globalEglBase.eglBaseContext).apply {
//            init(globalEglBase.eglBaseContext, null)
//            setScalingType(RendererCommon.ScalingType.SCALE_ASPECT_FIT)
//            setEnableHardwareScaler(true)
//            setZOrderMediaOverlay(true)
//        }
//    }
//
//    AndroidView(
//        factory = { renderer },
//        modifier = modifier
//    )
//
//    DisposableEffect(track) {
//        track.addSink(renderer)
//        onDispose {
//            track.removeSink(renderer)
//            renderer.release()
//        }
//    }
//}
