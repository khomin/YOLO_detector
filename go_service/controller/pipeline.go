package controller

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/sirupsen/logrus"
)

func (s *TrackerSession) startPipeline() error {
	s.recordCount += 1
	fileName := fmt.Sprintf("session_%04d_%03d.mp4", s.sessionId, s.recordCount)
	path := path.Join(s.env.RECORDINGS_TMP_DIR, fileName)
	args := []string{
		"fdsrc", "do-timestamp=true",
		"!", "image/jpeg",
		"!", "jpegparse",
		"!", "jpegdec",
		"!", "videoconvert",
		"!", "videorate",
		"!", "video/x-raw,framerate=30/1",
		"!", "x264enc", "tune=zerolatency", "speed-preset=ultrafast",
		"!", "h264parse",
		"!", "mp4mux", "fragment-duration=2000",
		"!", "filesink",
		"location=" + path, "sync=false",
	}

	s.gstCmd = exec.Command("gst-launch-1.0", args...)

	// 1. Get STDIN pipe (for sending data to GStreamer)
	var err error
	s.gstIn, err = s.gstCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// 2. Set STDERR (for capturing crash reasons)
	// s.gstCmd.Stderr = &stderr
	s.gstCmd.Stderr = os.Stderr
	s.gstCmd.Stdout = os.Stdout // Optional: keep stdout visible

	// 3. Start the GStreamer pipeline process
	if err := s.gstCmd.Start(); err != nil {
		return fmt.Errorf("failed to start gst-launch: %w", err)
		// return fmt.Errorf("failed to start gst-launch: %w (stderr: %s)", err, stderr.String())
	}

	// Now, launch a goroutine to wait for the GStreamer process to finish
	// This allows us to log the crash reason immediately.
	// go func() {
	// 	if err := s.gstCmd.Wait(); err != nil {
	// 		// This runs if the GStreamer process exits with a non-zero code (crashed)
	// 		logrus.Printf("!!! GStreamer CRASHED or EXITED")
	// 		// logrus.Printf("!!! Crash Reason (stderr): %s", stderr.String())
	// 		logrus.Printf("!!! Exit Error: %v", err)
	// 		// You might want to close the gRPC connection (or send a signal) here
	// 	}
	// }()

	logrus.Printf("GStreamer pipeline started")
	return nil
}

func (s *TrackerSession) stopPipeline() error {
	if s.gstIn != nil {
		s.gstIn.Close()
	}
	if s.gstCmd != nil {
		err := s.gstCmd.Wait()
		if err != nil {
			logrus.Printf("GStreamer exited with error: %v", err)
		}
	}
	logrus.Println("GStreamer finished")
	return nil
}

// args := []string{
// 	// "fdsrc",
// 	// "!",
// 	// "filesink", "location=/home/khomin/Desktop/capture1.jpeg", "buffer-mode=0",

// 	// "fdsrc", // "do-timestamp=true", // Tell GStreamer to time the frames as they arrive
// 	// "!",
// 	// "image/jpeg", //,framerate=30/1", // Force a framerate so the video has a "speed"
// 	// "!",
// 	// "jpegparse", // ASSEMBLER: Ensures the encoder gets 100% of the image
// 	// "!",
// 	// "jpegdec",
// 	// "!",
// 	// "videoconvert",
// 	// "!",
// 	// "x264enc", // "tune=zerolatency", "speed-preset=ultrafast",
// 	// "!",
// 	// "h264parse",
// 	// "!",
// 	// "mp4mux",
// 	// "!",
// 	// "filesink", "location=/home/khomin/Desktop/capture1.mp4", //, "sync=false",

// 	// "fdsrc",
// 	// // 2. Define Input Caps (CRITICAL)
// 	// // You must tell GStreamer this is JPEG and invent a framerate (e.g., 25 or 30 fps)
// 	// // otherwise x264enc will refuse to start.
// 	// "!", "image/jpeg", //,framerate=30/1",

// 	// // 3. Parse the Bytes
// 	// "!", "jpegparse", // Finds the start/end of each JPEG frame

// 	// // 4. Decode JPEG to Raw Video
// 	// "!", "jpegdec",

// 	// // 5. Convert Color Space
// 	// "!", "videoconvert", // Ensures compatibility with the encoder

// 	// // 6. Encode to H.264
// 	// // tune=zerolatency: Don't buffer frames; output immediately (prevents hangs)
// 	// // speed-preset=ultrafast: Sacrifice quality for speed (crucial for ARM CPUs)
// 	// "!", "x264enc", // "tune=zerolatency", "speed-preset=ultrafast",

// 	// // 7. Parse H.264 stream (Safety for the muxer)
// 	// "!", "h264parse",

// 	// // 8. Container Muxing
// 	// "!", "mp4mux",

// 	// // 9. Write to File
// 	// "!", "filesink", "location=/home/khomin/Desktop/capture1.mp4", //"sync=false",

// 	// // #
// 	// "fdsrc", "do-timestamp=true", // Reads from the pipe (your Go .Write calls)
// 	// "!",
// 	// "image/jpeg,framerate=30/1", // Tell GStreamer what the bytes are (CAPS ARE CRITICAL HERE)
// 	// "!",
// 	// "jpegparse", // robustly finds the start/end of JPEGs in the byte stream
// 	// "!",
// 	// "multipartmux", // wraps them into a playable MJPEG stream
// 	// "!",
// 	// "filesink", "location=/home/khomin/Desktop/capture_fixed.mjpeg", "sync=false",

// 	// // # gemini pro already better
// 	// "fdsrc", "do-timestamp=true",
// 	// "!",
// 	// "image/jpeg,framerate=5/1", // 1. Assume input is roughly 10 fps
// 	// "!",
// 	// "jpegparse",
// 	// "!",
// 	// "jpegdec", // 2. Decode so we can fix the timing
// 	// "!",
// 	// "videorate", // 3. SMOOTHING MAGIC: Fills gaps to make it steady
// 	// "!",
// 	// "video/x-raw,framerate=25/1", // 4. Output a standard 30fps stream (repeating frames if needed)
// 	// "!",
// 	// "jpegenc", // 5. Re-encode to JPEG (fast)
// 	// "!",
// 	// "multipartmux",
// 	// "!",
// 	// "filesink", "location=/home/khomin/Desktop/capture_fixed.mjpeg", "sync=false",

// 	// "fdsrc", "do-timestamp=true",
// 	// "!", "image/jpeg",
// 	// "!", "jpegparse", // Fixes the green corruption (frame assembly)

// 	// // 2. DECODE: We must decode to raw video to fix the timing
// 	// "!", "jpegdec",

// 	// // 3. THE GEARBOX: Fixes the "Time Lapse" / "Fast Forward" issue
// 	// "!", "videorate",

// 	// // 4. THE TARGET: Force the stream to become rigid 30fps
// 	// // GStreamer will now duplicate your 7 frames into 30 frames per second
// 	// "!", "video/x-raw,framerate=30/1",

// 	// // 5. ENCODE: Now we have a perfect stream for x264
// 	// "!", "videoconvert",
// 	// "!", "x264enc", "tune=zerolatency", "speed-preset=ultrafast",
// 	// "!", "h264parse",
// 	// "!", "mp4mux",
// 	// "!", "filesink", "location=/home/khomin/Desktop/capture_fixed.mp4", "sync=false",

// 	"fdsrc", "do-timestamp=true",
// 	"!", "image/jpeg",
// 	"!", "jpegparse",
// 	"!", "jpegdec",
// 	"!", "videoconvert",
// 	"!", "videorate",
// 	"!", "video/x-raw,framerate=30/1",
// 	"!", "x264enc", "tune=zerolatency", "speed-preset=ultrafast",
// 	"!", "h264parse",
// 	"!", "mp4mux", "fragment-duration=2000",
// 	"!", "filesink", "location=/home/khomin/Desktop/capture_fixed.mp4", "sync=false",
// }
