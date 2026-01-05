package com.example.detector

interface Platform {
    val name: String
}

expect fun getPlatform(): Platform