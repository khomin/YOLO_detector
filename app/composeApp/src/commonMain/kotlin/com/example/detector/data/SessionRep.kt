package com.example.detector.data

import kotlinx.coroutines.flow.MutableStateFlow

enum class SessionStatus {
    Active,
    Idle,
}

class SessionRep {
    val onStatus = MutableStateFlow(SessionStatus.Idle)
    val signalingClient = SignalingClient()

    fun connect() {

    }

    fun disconnect() {

    }
}