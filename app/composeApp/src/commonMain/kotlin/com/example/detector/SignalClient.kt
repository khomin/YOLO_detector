package com.example.detector

import com.example.detector.data.SessionDescriptionRequest
import com.example.detector.data.SessionDescriptionResponse
import com.example.detector.data.SessionInfo
import io.ktor.client.*
import io.ktor.client.plugins.contentnegotiation.*
import io.ktor.serialization.kotlinx.json.*
import io.ktor.client.request.*
import io.ktor.client.call.*
import io.ktor.http.*
import kotlinx.serialization.json.Json

class SignalingClient(private val baseUrl: String) {
    private val client = HttpClient {
        install(ContentNegotiation) {
            json(Json {
                ignoreUnknownKeys = true
                isLenient = true
            })
        }
    }

    suspend fun sendOffer(sessionId: String, offerSdp: String): String {
        val response: SessionDescriptionResponse = client.post("$baseUrl/signal/$sessionId") {
            contentType(ContentType.Application.Json)
            // Wraps the SDP string into the JSON object Go expects
            setBody(SessionDescriptionRequest(sdp = offerSdp, type = "offer"))
        }.body()
        return response.sdp // This is the 'Answer' from your Go service
    }

    suspend fun getSessions(): List<SessionInfo> {
        // Replace with your ARM board's IP
        val resp = client.get("$baseUrl/sessions").body<List<SessionInfo>>()
        return resp
    }
}