package com.example.detector

import android.app.Application
import com.example.detector.data.SessionRep
import com.example.detector.data.SettingsRep
import com.example.detector.data.SignalingClient
import com.example.detector.viewModels.MainViewModel
import com.example.detector.viewModels.ThemeViewModel
import com.shepeliev.webrtckmp.WebRtc
import org.koin.android.ext.koin.androidContext
import org.koin.android.ext.koin.androidLogger
import org.koin.core.context.startKoin
import org.koin.core.module.dsl.viewModel
import org.koin.dsl.module

val appModule = module {
    single { SettingsRep() }
    single { SessionRep() }
    single { ThemeViewModel(get()) }
    single {
        SignalingClient()
    }
    viewModel { MainViewModel(get(), get(), get()) }
}

class MainApplication : Application() {
    override fun onCreate() {
        super.onCreate()

        startKoin {
            // Log Koin into Android logger
            androidLogger()
            // Reference Android context
            androidContext(this@MainApplication)
            // Load modules
            modules(appModule)
        }

        val initializationOptionsBuilder = WebRtc.createInitializationOptionsBuilder()
        val peerConnectionFactoryBuilder = WebRtc.createPeerConnectionFactoryBuilder(initializationOptionsBuilder = initializationOptionsBuilder)
        WebRtc.configure(peerConnectionFactoryBuilder = peerConnectionFactoryBuilder)
    }
}