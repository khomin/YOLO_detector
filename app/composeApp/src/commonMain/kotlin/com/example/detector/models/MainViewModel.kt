package com.example.detector.models

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.detector.SignalingClient
import com.example.detector.data.SessionInfo
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.launch
import com.shepeliev.webrtckmp.*
import kotlinx.coroutines.flow.asStateFlow

class MainViewModel(private val signalingClient: SignalingClient) : ViewModel() {

    private val _sessions = MutableStateFlow<List<SessionInfo>>(emptyList())
    val sessions = _sessions.asStateFlow()

    private val _remoteTrack = MutableStateFlow<VideoTrack?>(null)
    val remoteTrack = _remoteTrack.asStateFlow()

    private var pc: PeerConnection? = null

    init {
        refreshSessions()
    }

    fun refreshSessions() {
        viewModelScope.launch {
            try {
                _sessions.value = signalingClient.getSessions()
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    fun connectToSession(sessionId: String) {
        viewModelScope.launch {
            val pc = PeerConnection(RtcConfiguration(
                iceServers = listOf(IceServer(urls = listOf("stun:stun.l.google.com:19302")))
            ))

            // 1. Setup the track collector first
            launch {
                pc.onTrack.collect { event ->
                    if (event.track is VideoTrack) {
                        _remoteTrack.value = event.track as VideoTrack
                    }
                }
            }

            // 2. CREATE OFFER WITH OPTIONS
            // This is the alternative to addTransceiver
            val options = OfferAnswerOptions(
                offerToReceiveVideo = true,
                offerToReceiveAudio = false
            )
            val offer = pc.createOffer(options)
            pc.setLocalDescription(offer)

            // 3. WAIT FOR ICE GATHERING (Just like you did in Go!)
            // If you don't do this, the Offer might be sent without your local IP
//            pc.iceGatheringState.first { it == IceGatheringState.Complete }

            // Replace the .first line with this:
//            if (pc.iceGatheringState.ordinal != IceGatheringState.Complete.ordinal) {
//                pc.iceGatheringState.collect { state ->
//                    if (state == IceGatheringState.Complete) return@collect
//                }
//            }

            // 4. SIGNALING
            val answerSdp = signalingClient.sendOffer(sessionId, pc.localDescription!!.sdp)
            pc.setRemoteDescription(SessionDescription(SessionDescriptionType.Answer, answerSdp))
        }
    }
//        viewModelScope.launch {
//            // 1. Setup PeerConnection
//            val config = RtcConfiguration(
//                iceServers = listOf(IceServer(urls = listOf("stun:stun.l.google.com:19302")))
//            )
//            pc = PeerConnection(config)
//
////            pc.addTransceiver(
////                kind = MediaStreamTrackKind.Video,
////                init = RtpTransceiverInit(direction = RtpTransceiverDirection.RecvOnly)
////            )
//
//            // 2. Listen for the Video Track (The Flow!)
//            launch {
////                pc?.connectionState?. { state ->
////                    println("Connection State: $state")
////                }
////                pc?.iceConnectionState?.collect { state ->
////                    println("ICE State: $state")
////                }
//                pc?.onTrack?.collect { event ->
//                    val track = event.track
//                    if (track is VideoTrack) {
//                        _remoteTrack.value = track
//                    }
//                }
//            }
//
//            // 3. Signaling Handshake
//            val offer = pc?.createOffer(OfferAnswerOptions()) ?: return@launch
//            pc?.setLocalDescription(offer)
//
//            try {
//                // Send offer to Go, get Answer back
//                val answerSdp = signalingClient.sendOffer(sessionId, offer.sdp)
//
//                // Set the Remote Description to start the stream
//                val answer = SessionDescription(
//                    type = SessionDescriptionType.Answer,
//                    sdp = answerSdp
//                )
//                pc?.setRemoteDescription(answer)
//            } catch (e: Exception) {
//                e.printStackTrace()
//            }
//        }
//    }
}

//class MainViewModel : ViewModel() {
//    private val signalingClient = SignalingClient("http://192.168.1.6:8081")
//
//    val sessions = MutableStateFlow<List<SessionInfo>>(emptyList())
//
//    fun refreshSessions() {
//        viewModelScope.launch {
//            sessions.value = signalingClient.getSessions()
//        }
//    }
//
//    suspend fun init() {
//        val config = RtcConfiguration(
//            iceServers = listOf(IceServer(urls = listOf("stun:stun.l.google.com:19302")))
//        )
//        // In 0.12.0, PeerConnection is an interface you create via a factory or constructor
//        val pc = PeerConnection(config)
//
//        pc.onTrack.collect { event ->
//            val track = event.track
//            if (track?.kind == MediaStreamTrackKind.Video) {
//                // This is your YOLO video track!
//                remoteVideoTrack = track as VideoTrack
//
//                // If you're using Compose, update a State variable
//                _videoState.value = track
//            }
//        }
//
////        // To get the track
////        pc.onTrack
////        pc.onRemoveTrack
////        onVideoTrack.collect { track ->
////            // This is a Flow in the new version!
////            // This is much better for Compose.
////            remoteTrack = track
////        }
//    }
//}