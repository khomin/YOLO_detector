package com.example.detector.viewModels

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.detector.data.SignalingClient
import com.example.detector.data.SessionInfo
import com.example.detector.data.SessionRep
import com.example.detector.data.SettingsRep
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.launch
import com.shepeliev.webrtckmp.*
import kotlinx.coroutines.flow.asStateFlow

class MainViewModel(
    private val settingsRep: SettingsRep,
    private val signalingClient: SignalingClient,
    private val sessionRep: SessionRep
) : ViewModel() {

    private var _loading = MutableStateFlow(true)
    val loading = _loading.asStateFlow()

    private val _sessions = MutableStateFlow<List<SessionInfo>>(emptyList())
    val sessions = _sessions.asStateFlow()

    private val _remoteTrack = MutableStateFlow<VideoTrack?>(null)
    val remoteTrack = _remoteTrack.asStateFlow()

    // url in settings pointing to SignalClient
    // if empty do nothing -> can be set in SettingsScreen
    // if not empty -> use cached data and start SignalClient then get available sessions in list

    suspend fun init() {
        // read settings
        val uri = settingsRep.getServerUrl()

        signalingClient.init(uri)

        _loading.value = false
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
}