package com.example.detector.data

import kotlinx.serialization.Serializable

@Serializable
data class SessionDescriptionRequest(
    val sdp: String,
    val type: String // Usually "offer" or "answer"
)

@Serializable
data class SessionDescriptionResponse(
    val sdp: String,
    val type: String
)

@Serializable
data class SessionInfo(
    val id: String,
    val state: String
)