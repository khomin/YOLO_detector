package com.example.detector

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.shepeliev.webrtckmp.VideoTrack

@Composable
expect fun VideoRenderer(track: VideoTrack?, modifier: Modifier = Modifier)