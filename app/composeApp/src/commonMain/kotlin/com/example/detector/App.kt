package com.example.detector

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.example.detector.data.SessionRep
import com.example.detector.data.ThemeMode
import com.example.detector.viewModels.MainViewModel
import com.example.detector.ui.screens.LoadScreen
import com.example.detector.ui.screens.MainScreen
import com.example.detector.ui.screens.SettingsScreen
import com.example.detector.viewModels.ThemeViewModel
import org.koin.compose.koinInject
import org.koin.compose.viewmodel.koinViewModel

//val viewModel = remember { MainViewModel(SignalingClient("http://192.168.1.6:8081")) }
//val sessions by viewModel.sessions.collectAsState()
//val remoteTrack by viewModel.remoteTrack.collectAsState()

@Composable
fun App() {
    val model = koinViewModel<MainViewModel>()
    val sessionRep: SessionRep = koinInject()
    val themeViewModel: ThemeViewModel = koinViewModel()

    val loading by model.loading.collectAsState()
    val status by sessionRep.onStatus.collectAsState()
    val themeMode by themeViewModel.themeMode.collectAsState()
    val systemInDarkTheme = isSystemInDarkTheme()

    // Determine actual theme
    val isDarkTheme = when (themeMode) {
        ThemeMode.LIGHT -> false
        ThemeMode.DARK -> true
        ThemeMode.SYSTEM -> systemInDarkTheme
    }

    LaunchedEffect(Unit) {
        model.init()
        sessionRep.connect()
        model.refreshSessions()
    }

    DisposableEffect(Unit) {
        onDispose {
            sessionRep.disconnect()
        }
    }

    MaterialTheme(
        colorScheme = if (isDarkTheme) darkColorScheme() else lightColorScheme()
    ) {
        if(loading) {
            LoadScreen()
        } else {
            NavigatorApp()
        }
    }
}

@Composable
fun NavigatorApp() {
    var selectedTab by rememberSaveable { mutableStateOf(0) }
    Scaffold(
        modifier = Modifier.fillMaxSize(),
        bottomBar = {
            Surface(
                modifier = Modifier
                    .fillMaxWidth()
                    .imePadding()
                    .navigationBarsPadding()
            ) {
                NavigationBar(
                    modifier = Modifier
                        .height(56.dp)
                        .fillMaxWidth(),
                    containerColor = MaterialTheme.colorScheme.surfaceContainer
                ) {
                    NavigationBarItem(
                        selected = selectedTab == 0,
                        onClick = { selectedTab = 0 },
                        icon = { Icon(Icons.Default.Home, null) },
                        label = { Text("Main") }
                    )
                    NavigationBarItem(
                        selected = selectedTab == 1,
                        onClick = { selectedTab = 1 },
                        icon = { Icon(Icons.Default.Settings, null) },
                        label = { Text("Settings") }
                    )
                }
            }
        }
    ) { innerPadding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding)
        ) {
            when (selectedTab) {
                0 -> MainScreen()
                1 -> SettingsScreen()
            }
        }
    }
}

//@Composable
//fun AppNavigation() {
//    val navController = rememberNavController()
//
//    NavHost(
//        navController = navController,
//        startDestination = "main"
//    ) {
//        composable("main") {
//            MainScreen(navController)
//        }
//        composable("settings") {
//            SettingsScreen(navController)
//        }
//        composable("gallery") {
//            GalleryScreen(navController)
//        }
//        composable("play/{sessionId}") { backStackEntry ->
//            val sessionId = backStackEntry.arguments?.getString("sessionId")
//            PlayScreen(navController, sessionId)
//        }
//    }
//}