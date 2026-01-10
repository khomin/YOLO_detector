package com.example.detector.data

import com.russhwolf.settings.Settings
import kotlinx.coroutines.flow.Flow

enum class ThemeMode {
    LIGHT,
    DARK,
    SYSTEM
}

class SettingsRep {
    private val settings: Settings = Settings()

    companion object {
        private const val THEME_KEY = "theme_mode"
    }

    fun saveThemeMode(mode: ThemeMode) {
        settings.putString(THEME_KEY, mode.name)
    }

    fun getThemeMode(): ThemeMode {
        val saved = settings.getStringOrNull(THEME_KEY)
        return saved?.let { ThemeMode.valueOf(it) } ?: ThemeMode.SYSTEM
    }

//    fun getThemeModeFlow(): Flow<ThemeMode> = flow {
//        // Emit initial value
//        emit(getThemeMode())
//        // In a real app, you'd want to listen for changes
//        // For now, this is a simple flow
//    }

    fun saveServerUrl(url: String) {
        settings.putString("server_url", url)
    }

    fun getServerUrl(): String {
        return settings.getString("server_url", "http://192.168.1.6:8081")
    }

//    fun getServerUrl(): String {
//        return "http://192.168.1.6:8081"
//    }
}