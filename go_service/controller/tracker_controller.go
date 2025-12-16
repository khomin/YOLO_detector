package controller

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
	"yolo-detector-service/bootstrap"
	pb "yolo-detector-service/grpc/generated"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/peer"
)

type TrackerState int

const (
	StateIdle TrackerState = iota
	StateRun
	StateCanceled
)

type TrackerSession struct {
	state    TrackerState
	timer    *time.Ticker
	doneChan chan struct{}
	gstCmd   *exec.Cmd
	gstIn    io.WriteCloser
	lock     sync.Mutex
	count    int
}

type TrackerServer struct {
	Env      *bootstrap.Env
	Trackers map[string]*TrackerSession
	lock     sync.Mutex
	// Required to be embedded for forward compatibility
	pb.UnimplementedTrackerServiceServer
}

// The duration the server waits after high-priority_active turns false
// const CooldownDuration = 10 * time.Second

// stream server
// - register streams
// - control its status
//		- idle, run, timeout
// - write temp file
// - put frames to temp file
// - add record to database when status - canceled
// - delete temp file

// 1. Check for High-Priority Activity (e.g., Dog detected)
// if update.GetHighPriorityActive() {
// 	if s.sessionState == "IDLE" {
// 		log.Println("-> STATE CHANGE: IDLE -> RECORDING (High priority event)")
// 		s.sessionState = "RECORDING"
// 		// Action: Start video archiving / web stream here
// 	}
// 	// Reset the cooldown timer whenever activity is detected
// 	cooldownTimer.Reset(CooldownDuration)
// }

func (s *TrackerServer) StreamUpdates(stream pb.TrackerService_StreamUpdatesServer) error {
	p, ok := peer.FromContext(stream.Context())
	if !ok {
		return errors.New("peer information unavailable")
	}

	addr := p.Addr.String()
	log.Printf("New C++ client connected [%s]", addr)

	s.lock.Lock()
	defer s.lock.Unlock()
	sessionPtr, ok := s.Trackers[addr]
	if !ok {
		sessionPtr = &TrackerSession{
			doneChan: make(chan struct{}),
		}
		s.Trackers[addr] = sessionPtr
	}

	sessionPtr.lock.Lock()
	defer sessionPtr.lock.Unlock()
	defer close(sessionPtr.doneChan)

	switch sessionPtr.state {
	case StateIdle:
		sessionPtr.state = StateRun
		sessionPtr.timer = time.NewTicker(time.Duration(1 * time.Second))
		if err := sessionPtr.startGStreamerPipeline(addr); err != nil {
			log.Printf("Failed to start GStreamer for %s: %v", addr, err)
			return err // Close the gRPC connection on failure
		}
		go func() {
			for {
				select {
				case <-sessionPtr.timer.C:
					logrus.Printf("[%s] Session timer ticked.", addr)
				case <-sessionPtr.doneChan: // Assuming you add a DoneChan to TrackerSession
					logrus.Printf("[%s] Session cleanup signal received. Stopping ticker.", addr)
					sessionPtr.timer.Stop()
					return
				}
			}
		}()
		for {
			update, err := stream.Recv()
			if err == io.EOF {
				log.Println("C++ client stream finished. Shutting down session.")
				sessionPtr.state = StateCanceled
				success := true
				return stream.SendAndClose(&pb.StreamStatus{Success: &success})
			}
			if err != nil {
				log.Printf("Error receiving frame update: %v", err)
				return err
			}
			sessionPtr.ProcessUpdate(update)
		}
	case StateRun:
		break
	case StateCanceled:
		break
	}
	return nil
}

func (cc *TrackerSession) ProcessUpdate(update *pb.FrameUpdate) {
	//
	// events := update.GetEvents()
	// frameNum := update.GetFrameNumber()
	frame := update.EncodedFrame
	if len(frame) > 0 {
		logrus.Printf("Received frame: %d", len(frame))

		// if cc.count >= 2 {
		// 	cc.gstIn.Close()
		// 	return
		// }
		// cc.count++
		n, err := cc.gstIn.Write(frame)
		if err != nil {
			logrus.Errorf("Error writing frame to GStreamer stdin: %v", err)
			// You may want to kill the GStreamer process and/or session here
			return
		}
		if n != len(frame) {
			logrus.Warnf("Incomplete write to GStreamer: Wrote %d of %d bytes", n, len(frame))
		}
	} else {
		logrus.Print("No Frame")
	}
}

func (s *TrackerSession) startGStreamerPipeline(addr string) error {
	// pipeline := "appsrc name=mysource is-live=true block=true format=3 ! image/jpeg ! jpegdec ! videoconvert ! autovideosink"
	// pipeline := "appsrc name=mysource ! image/jpeg ! jpegdec ! videoconvert ! x264enc tune=zerolatency bitrate=500 speed-preset=ultrafast ! mp4mux ! queue leaky=downstream max-size-buffers=1 ! filesink location=/home/khomin/Desktop/capture1.mp4"

	// pipeline := "appsrc name=mysource is-live=true format=3 ! image/jpeg ! jpegdec ! videoconvert ! fakesink"
	// pipeline := "appsrc name=mysource is-live=true format=3 do-timestamp=true ! image/jpeg ! multipartmux ! fakesink sync=false"

	// pipeline := "videotestsrc ! autovideosink"

	// args := []string{
	// 	"videotestsrc",
	// 	"!",
	// 	"autovideosink",
	// }

	args := []string{
		// "fdsrc",
		// "!",
		// "filesink", "location=/home/khomin/Desktop/capture1.jpeg", "buffer-mode=0",

		// "fdsrc", // "do-timestamp=true", // Tell GStreamer to time the frames as they arrive
		// "!",
		// "image/jpeg", //,framerate=30/1", // Force a framerate so the video has a "speed"
		// "!",
		// "jpegparse", // ASSEMBLER: Ensures the encoder gets 100% of the image
		// "!",
		// "jpegdec",
		// "!",
		// "videoconvert",
		// "!",
		// "x264enc", // "tune=zerolatency", "speed-preset=ultrafast",
		// "!",
		// "h264parse",
		// "!",
		// "mp4mux",
		// "!",
		// "filesink", "location=/home/khomin/Desktop/capture1.mp4", //, "sync=false",

		// "fdsrc",

		// // 2. Define Input Caps (CRITICAL)
		// // You must tell GStreamer this is JPEG and invent a framerate (e.g., 25 or 30 fps)
		// // otherwise x264enc will refuse to start.
		// "!", "image/jpeg", //,framerate=30/1",

		// // 3. Parse the Bytes
		// "!", "jpegparse", // Finds the start/end of each JPEG frame

		// // 4. Decode JPEG to Raw Video
		// "!", "jpegdec",

		// // 5. Convert Color Space
		// "!", "videoconvert", // Ensures compatibility with the encoder

		// // 6. Encode to H.264
		// // tune=zerolatency: Don't buffer frames; output immediately (prevents hangs)
		// // speed-preset=ultrafast: Sacrifice quality for speed (crucial for ARM CPUs)
		// "!", "x264enc", // "tune=zerolatency", "speed-preset=ultrafast",

		// // 7. Parse H.264 stream (Safety for the muxer)
		// "!", "h264parse",

		// // 8. Container Muxing
		// "!", "mp4mux",

		// // 9. Write to File
		// "!", "filesink", "location=/home/khomin/Desktop/capture1.mp4", //"sync=false",

		// #

		"fdsrc", "do-timestamp=true", // Reads from the pipe (your Go .Write calls)
		"!",
		"image/jpeg,framerate=30/1", // Tell GStreamer what the bytes are (CAPS ARE CRITICAL HERE)
		"!",
		"jpegparse", // robustly finds the start/end of JPEGs in the byte stream
		"!",
		"multipartmux", // wraps them into a playable MJPEG stream
		"!",
		"filesink", "location=/home/khomin/Desktop/capture_fixed.mjpeg", "sync=false",

		// "appsrc", "name=mysource", "is-live=true", "format=3", "emit-signals=false",
		// "!",
		// "image/jpeg",
		// "!",
		// "filesink", "location=/home/khomin/Desktop/capture1.jpeg", "buffer-mode=0",

		// 	"appsrc",
		// 	"name=mysource",
		// 	"!",
		// 	"image/jpeg",
		// 	"!",
		// 	"filesink", "location=/home/khomin/Desktop/capture1.jpeg",

		// "jpegdec",
		// "!",
		// "videoconvert",
		// "!",
		// "x264enc",
		// "tune=zerolatency",
		// "!",
		// "mp4mux",
		// "!",
		// "queue", "leaky=downstream", "max-size-buffers=100",
		// "!",
		// "filesink","location=/home/khomin/Desktop/capture1.mp4",
	}

	s.gstCmd = exec.Command("gst-launch-1.0", args...)

	// TODO:
	// s.gstIn.Close() // Sends EOF
	// s.gstCmd.Wait()

	// s.gstCmd = exec.Command("gst-launch-1.0", "-v", pipeline)

	// pipeline := "fdsrc fd=0 ! image/jpeg ! jpegdec ! videoconvert ! autovideosink"

	// s.gstCmd = exec.Command("/bin/sh", "-c", "gst-launch-1.0 -v "+pipeline)

	// s.gstCmd = exec.Command(
	// 	"gst-launch-1.0",
	// 	"-v",
	// 	"fdsrc fd=0 ! image/jpeg ! jpegdec ! videoconvert ! autovideosink",
	// )

	// stdin, _ := cmd.StdinPipe()
	// cmd.Start()

	// stdin.Write(jpegFrame)

	// var stderr bytes.Buffer // Buffer to capture error output

	// // 1. Get STDIN pipe (for sending data to GStreamer)
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
	go func() {
		if err := s.gstCmd.Wait(); err != nil {
			// This runs if the GStreamer process exits with a non-zero code (crashed)
			log.Printf("!!! GStreamer CRASHED or EXITED !!! Address: [%s]", addr)
			// log.Printf("!!! Crash Reason (stderr): %s", stderr.String())
			log.Printf("!!! Exit Error: %v", err)
			// You might want to close the gRPC connection (or send a signal) here
		}
	}()

	log.Printf("[%s] GStreamer pipeline started", addr)
	return nil
}

func (cc *TrackerServer) TestMethod(c *gin.Context) {
	response := map[string]interface{}{
		"success": true,
	}
	c.JSON(http.StatusOK, response)
}

func (cc *TrackerServer) OnTrackUpdate(c *gin.Context) {
	response := map[string]interface{}{
		"success": true,
	}
	c.JSON(http.StatusOK, response)
}

func (s *TrackerSession) DD() {

}
