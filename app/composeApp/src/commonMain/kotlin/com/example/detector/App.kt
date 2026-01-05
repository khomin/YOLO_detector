package com.example.detector

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.safeContentPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.example.detector.models.MainViewModel

@Composable
fun App() {
    val viewModel = remember { MainViewModel(SignalingClient("http://192.168.1.6:8081")) }
    val sessions by viewModel.sessions.collectAsState()
    val remoteTrack by viewModel.remoteTrack.collectAsState()

    MaterialTheme {
        Column(
            modifier = Modifier
                .background(MaterialTheme.colorScheme.primaryContainer)
                .safeContentPadding()
                .fillMaxSize(),
        ) {
            Text(if(remoteTrack == null) "Idle" else "Active")

            if (remoteTrack != null) {
                // Show the YOLO video
                VideoRenderer(remoteTrack, Modifier.fillMaxWidth().height(400.dp))
                Button(onClick = { /* stop logic */ }) { Text("Close Stream") }
            } else {
                // Show the list of sessions
                LazyColumn {
                    items(sessions) { session ->
                        Button(onClick = { viewModel.connectToSession(session.id.toString()) }) {
                            Text("Connect to Session: ${session.id} (${session.state})")
                        }
                    }
                }
                Button(onClick = {
                    viewModel.refreshSessions()
                }) { Text("Refresh") }
            }
        }
    }
}

//@Composable
//@Preview
//fun App() {
//    val viewModel = remember { MainViewModel() }
//    val sessions by viewModel.sessions.collectAsState()
//
//    MaterialTheme {
////        var showContent by remember { mutableStateOf(false) }
//        Column(
//            modifier = Modifier
//                .background(MaterialTheme.colorScheme.primaryContainer)
//                .safeContentPadding()
//                .fillMaxSize(),
//            horizontalAlignment = Alignment.CenterHorizontally,
//        ) {
////            Button(onClick = { showContent = !showContent }) {
////                Text("Click me! $showContent")
////            }
//            Button(onClick = {
//                viewModel.refreshSessions()
//            }) {
//                Text("Refresh")
//            }
//
//            LazyColumn(modifier = Modifier.fillMaxSize()) {
//                items(sessions.size) { index ->
//                    val session = sessions[index]
//                    Text("session: id=${session.id}, state=${session.state}")
//                }
//            }
//        }
//    }
//}