package com.example.detector.ui.screens

import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material3.Button
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.example.detector.ui.components.VideoRenderer

@Composable
fun MainScreen() {
    Text("main screen")
//    Text(if(remoteTrack == null) "Idle" else "Active")
//
//    if (remoteTrack != null) {
//        // Show the YOLO video
//        VideoRenderer(remoteTrack, Modifier.fillMaxWidth().height(400.dp))
//        Button(onClick = { /* stop logic */ }) { Text("Close Stream") }
//    } else {
//        // Show the list of sessions
//        LazyColumn {
//            items(sessions) { session ->
//                Button(onClick = { viewModel.connectToSession(session.id) }) {
//                    Text("Connect to Session: ${session.id} (${session.state})")
//                }
//            }
//        }
//        Button(onClick = {
//            viewModel.refreshSessions()
//        }) { Text("Refresh") }
//    }
}
