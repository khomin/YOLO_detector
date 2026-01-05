package controller

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
	"yolo-detector-service/bootstrap"
	pb "yolo-detector-service/grpc/generated"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/sirupsen/logrus"
)

type TrackerState int

const (
	StateIdle TrackerState = iota
	StateRun
	StateCanceled
)

type TrackerSession struct {
	sessionId     int
	state         TrackerState
	timer         *time.Ticker
	streamStarted time.Time
	trackerTime   TrackerTime
	videoTrack    *webrtc.TrackLocalStaticSample
	recordCount   int
	// gstreamer webRTC pipe
	gstWebRtcCmd *exec.Cmd
	gstWebRtcIn  io.WriteCloser
	gstWebRtcOut io.ReadCloser
	// gstreamer file recording
	gstCmd   *exec.Cmd
	gstIn    io.WriteCloser
	doneChan chan struct{}
	env      *bootstrap.Env
	lock     sync.Mutex
}

type TrackerTime struct {
	firstEvent    *pb.TrackEvent
	lastEvent     *pb.TrackEvent
	env           *bootstrap.Env
	preRecordBuff [][]byte
}

type SessionInfo struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

func InitSession(id int, env *bootstrap.Env) *TrackerSession {
	s := TrackerSession{
		doneChan:  make(chan struct{}),
		sessionId: id,
		env:       env,
		trackerTime: TrackerTime{
			env: env,
		},
	}
	s.startWebRTC()
	return &s
}

func (cc *TrackerSession) startSession(addr string, stream pb.TrackerService_StreamUpdatesServer) error {
	cc.state = StateIdle
	cc.streamStarted = time.Now()
	cc.timer = time.NewTicker(cc.env.SESSION_TASK_TIMER)
	go func() {
		for {
			select {
			case <-cc.timer.C:
				// logrus.Printf("[%s] Session timer ticked.", addr)
				switch cc.state {
				case StateIdle:
					cc.lock.Lock()
					if cc.trackerTime.hasTargetFor(cc.env.TARGET_THRESHOLD_DURATION) {
						cc.trackerTime.clear()
						cc.startPipeline()
						cc.state = StateRun
					}
					cc.lock.Unlock()
				case StateRun:
					cc.lock.Lock()
					if cc.trackerTime.noTargetFor(cc.env.TARGET_THRESHOLD_DURATION) {
						cc.trackerTime.clear()
						cc.stopPipeline()
						cc.state = StateIdle
					}
					cc.lock.Unlock()
				case StateCanceled:
					break
				}
			case <-cc.doneChan:
				logrus.Printf("[%s] Session cleanup signal received. Stopping ticker.", addr)
				return
			}
		}
	}()
	for {
		update, err := stream.Recv()
		if err == io.EOF {
			logrus.Println("C++ client stream finished. Shutting down session.")
			cc.state = StateCanceled
			success := true
			return stream.SendAndClose(&pb.StreamStatus{Success: &success})
		}
		if err != nil {
			logrus.Printf("Error receiving frame update: %v", err)
			return err
		}
		cc.processUpdate(update)
	}
}

func (cc *TrackerSession) closeSession() {
	logrus.Println("Stopping GStreamer...")
	close(cc.doneChan)
	cc.stopPipeline()
}

func (c *TrackerTime) updateTime(events []*pb.TrackEvent) {
	hasTarget := c.containsAllowedClass(events)
	if hasTarget {
		if c.firstEvent == nil {
			c.firstEvent = events[len(events)-1]
		}
		c.lastEvent = events[len(events)-1]
	}
}

func (c *TrackerTime) clear() {
	c.firstEvent = nil
	c.lastEvent = nil
}

func (cc *TrackerTime) containsAllowedClass(events []*pb.TrackEvent) bool {
	classes := cc.env.SESSION_ALLOWED_CLASSES
	for _, event := range events {
		for _, name := range classes {
			if name == *event.ClassName {
				return true
			}
		}
	}
	return false
}

func (cc *TrackerTime) hasTargetFor(duration time.Duration) bool {
	if cc.firstEvent == nil {
		return false
	}
	if time.Since(time.UnixMilli(*cc.firstEvent.TimestampMs)) > duration {
		return true
	}
	return false
}

func (cc *TrackerTime) noTargetFor(duration time.Duration) bool {
	if cc.lastEvent == nil {
		return true
	}
	if time.Since(time.UnixMilli(*cc.lastEvent.TimestampMs)) > duration {
		return true
	}
	return false
}

func (cc *TrackerSession) processUpdate(update *pb.FrameUpdate) {
	cc.trackerTime.updateTime(update.Events)

	if len(update.EncodedFrame) > 0 {
		switch cc.state {
		case StateIdle:
			maxPreRoll := 150
			cc.trackerTime.preRecordBuff = append(cc.trackerTime.preRecordBuff, update.EncodedFrame)
			if len(cc.trackerTime.preRecordBuff) > maxPreRoll {
				cc.trackerTime.preRecordBuff = cc.trackerTime.preRecordBuff[1:]
			}
		case StateRun:
			if len(cc.trackerTime.preRecordBuff) > 0 {
				for _, i := range cc.trackerTime.preRecordBuff {
					cc.writeFrame(i)
				}
				cc.trackerTime.preRecordBuff = [][]byte{}
			}
			cc.writeFrame(update.EncodedFrame)
		}
	}
}

func (cc *TrackerSession) writeFrame(frame []byte) error {
	n, err := cc.gstIn.Write(frame)
	if err != nil {
		logrus.Errorf("Error writing frame to GStreamer stdin: %v", err)
		return err
	}
	if n != len(frame) {
		logrus.Warnf("Incomplete write to GStreamer: Wrote %d of %d bytes", n, len(frame))
	}
	_, err = cc.gstWebRtcIn.Write(frame)
	if err != nil {
		fmt.Println("Error writing to track:", err)
	}
	return nil
}

func (cc *TrackerSession) startWebRTC() error {
	// 1. Setup the command. REMOVE stdout/stderr assignment here.
	// cc.gstWebRtcCmd = exec.Command("gst-launch-1.0", []string{
	// 	"fdsrc", "do-timestamp=true",
	// 	"!", "image/jpeg",
	// 	"!", "jpegparse",
	// 	"!", "jpegdec",
	// 	"!", "videoconvert",
	// 	"!", "video/x-raw,format=I420", // Explicitly set the format Android loves
	// 	"!", "x264enc", "bitrate=2000", "tune=zerolatency", "speed-preset=ultrafast", "key-int-max=15", "byte-stream=true",
	// 	"!", "video/x-h264,profile=baseline",
	// 	"!", "h264parse", "config-interval=1",
	// 	"!", "video/x-h264,stream-format=byte-stream",
	// 	// "!", "filesink", "location=/home/khomin/Desktop/test.h264",
	// 	"!", "fdsink", "fd=1", "sync=false",
	// }...)

	cc.gstWebRtcCmd = exec.Command("gst-launch-1.0", []string{
		"fdsrc", "do-timestamp=true",
		"!", "image/jpeg",
		"!", "jpegparse",
		"!", "jpegdec",
		"!", "videoconvert",
		"!", "video/x-raw,format=I420", // Explicitly set the format Android loves
		"!", "x264enc", "bitrate=2000", "tune=zerolatency", "speed-preset=ultrafast", "sliced-threads=false", "key-int-max=15",
		"!", "video/x-h264,profile=baseline,stream-format=byte-stream",
		"!", "h264parse", "config-interval=-1",
		"!", "video/x-h264,stream-format=byte-stream,alignment=au", // 'au' means Access Unit (Full Frame)
		"!", "fdsink", "fd=1", "sync=false",
	}...)

	// 2. Setup Pipes
	cc.gstWebRtcIn, _ = cc.gstWebRtcCmd.StdinPipe()
	cc.gstWebRtcOut, _ = cc.gstWebRtcCmd.StdoutPipe()
	cc.gstWebRtcCmd.Stderr = os.Stderr // Only pipe stderr to console

	// 3. Create the Track
	cc.videoTrack, _ = webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeH264,
			SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
		},
		"video",
		"pion",
	)
	if err := cc.gstWebRtcCmd.Start(); err != nil {
		return err
	}
	// 4. The Reader Loop
	// go func() {
	// 	// Large buffer for full frames
	// 	reader := bufio.NewReaderSize(cc.gstWebRtcOut, 256*1024)
	// 	for {
	// 		// Look for the next Start Code (00 00 00 01)
	// 		// This is a simple way: Read until we find the next start of a frame
	// 		data, err := reader.ReadBytes(0x01)
	// 		if err != nil {
	// 			return
	// 		}
	// 		// Logic: Collect data until you have a full NAL unit.
	// 		// For now, let's ensure we aren't dropping data.
	// 		if len(data) > 100 {
	// 			// Re-add the start code prefix that ReadBytes consumed
	// 			nal := append([]byte{0x00, 0x00, 0x00, 0x01}, data...)

	// 			cc.videoTrack.WriteSample(media.Sample{
	// 				Data:     nal,
	// 				Duration: time.Millisecond * 33,
	// 			})
	// 		}
	// 	}
	// }()
	go func() {
		scanner := bufio.NewScanner(cc.gstWebRtcOut)
		buf := make([]byte, 0, 1024*1024)
		scanner.Buffer(buf, 1024*1024)
		scanner.Split(splitAnnexB)

		var headerStack []byte // To store SPS/PPS until a real frame arrives

		for scanner.Scan() {
			data := scanner.Bytes()
			if len(data) == 0 {
				continue
			}

			// 1. If it's a small packet (SPS/PPS/Metadata), save it.
			if len(data) < 100 {
				headerStack = append(headerStack, []byte{0x00, 0x00, 0x00, 0x01}...)
				headerStack = append(headerStack, data...)
				continue
			}

			// 2. If it's a big packet (Video Frame), attach any saved headers and send.
			finalPacket := append([]byte{0x00, 0x00, 0x00, 0x01}, data...)
			if len(headerStack) > 0 {
				finalPacket = append(headerStack, finalPacket...)
				headerStack = nil // Clear the stack
			}

			cc.videoTrack.WriteSample(media.Sample{
				Data:     finalPacket,
				Duration: time.Millisecond * 33,
			})
		}
	}()
	// go func() {
	// 	// 1MB buffer to hold the incoming stream
	// 	reader := bufio.NewReaderSize(cc.gstWebRtcOut, 1024*1024)

	// 	// We look for the start code: 00 00 00 01
	// 	// A simple trick is to Read until 0x01, then check if the previous 3 bytes were 0x00
	// 	for {
	// 		data, err := reader.ReadBytes(0x01)
	// 		if err != nil {
	// 			logrus.Errorf("GStreamer pipe closed: %v", err)
	// 			return
	// 		}

	// 		// We found a '01'. Now we check if we have enough data to be a NAL unit
	// 		if len(data) > 5 {
	// 			// We strip the trailing 01 and the preceding 00s to get the raw NAL
	// 			// Then we wrap it in a clean 4-byte start code
	// 			cleanData := append([]byte{0x00, 0x00, 0x00, 0x01}, data[:len(data)-1]...)

	// 			err = cc.videoTrack.WriteSample(media.Sample{
	// 				Data:     cleanData,
	// 				Duration: time.Millisecond * 33,
	// 			})

	// 			if err == nil {
	// 				// You should see lengths like 5000-20000 here
	// 				// If you still see len=2, the logic below is skipping them
	// 				if len(cleanData) > 100 {
	// 					logrus.Debugf("Sent Frame: %d bytes", len(cleanData))
	// 				}
	// 			}
	// 		}
	// 	}
	// }()
	// go func() {
	// 	// Use a H.264 Annex-B splitter to find the 'Start Codes'
	// 	// This ensures Pion gets whole frames/NALs
	// 	scanner := bufio.NewScanner(cc.gstWebRtcOut)
	// 	scanner.Split(splitAnnexB)
	// 	for scanner.Scan() {
	// 		data := scanner.Bytes()
	// 		// Ignore junk
	// 		if len(data) < 5 {
	// 			continue
	// 		}
	// 		// Prefix the data with the start code that the scanner stripped
	// 		nal := append([]byte{0x00, 0x00, 0x00, 0x01}, data...)

	// 		err := cc.videoTrack.WriteSample(media.Sample{
	// 			Data: nal,
	// 			// Data:     data,
	// 			Duration: time.Millisecond * 33,
	// 		})
	// 		if err == nil {
	// 			logrus.Printf("write sample: len=%d", len(data))
	// 		} else {
	// 			logrus.Printf("write sample: len=%d, error=%w", len(data), err)
	// 		}
	// 	}
	// }()
	return nil
}

// func (cc *TrackerSession) startWebRTC() error {
// 	cc.gstWebRtcCmd = exec.Command("gst-launch-1.0", []string{
// 		"fdsrc", "do-timestamp=true",
// 		"!", "image/jpeg",
// 		"!", "jpegparse",
// 		"!", "jpegdec",
// 		"!", "videoconvert",
// 		"!", "x264enc", "bitrate=2000", "tune=zerolatency", "speed-preset=ultrafast", "key-int-max=30",
// 		"!", "h264parse", "config-interval=1", // config-interval=1 is MAGIC for fixing black screens
// 		"!", "video/x-h264,stream-format=byte-stream", // Ensures format is Annex-B
// 		"!", "fdsink", "fd=1", "sync=false", // Output to STDOUT
// 	}...)
// 	stdIn, err := cc.gstWebRtcCmd.StdinPipe()
// 	if err != nil {
// 		return fmt.Errorf("failed to get stdin pipe: %w", err)
// 	}
// 	stdOut, err := cc.gstWebRtcCmd.StdoutPipe()
// 	if err != nil {
// 		return fmt.Errorf("failed to get stdin pipe: %w", err)
// 	}
// 	cc.gstWebRtcIn = stdIn
// 	cc.gstWebRtcOut = stdOut
// 	cc.gstWebRtcCmd.Stderr = os.Stderr
// 	// cc.gstWebRtcCmd.Stdout = os.Stdout

// 	if err = cc.gstWebRtcCmd.Start(); err != nil {
// 		return fmt.Errorf("failed to start gst-launch: %w", err)
// 	}
// 	cc.videoTrack, _ = webrtc.NewTrackLocalStaticSample(
// 		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
// 		"video",
// 		"pion",
// 	)
// 	go func() {
// 		// A simple buffer to read the H.264 stream
// 		// In a production app, you'd use a more robust NAL unit splitter,
// 		// but for a start, reading chunks works with many decoders.
// 		buffer := make([]byte, 4096)
// 		for {
// 			n, err := cc.gstWebRtcOut.Read(buffer)
// 			if err != nil {
// 				break
// 			}
// 			if n > 0 {
// 				// Push the encoded H.264 bytes to Android
// 				cc.videoTrack.WriteSample(media.Sample{
// 					Data:     buffer[:n],
// 					Duration: time.Millisecond * 33, // Assume 30fps
// 				})
// 			}
// 		}
// 	}()
// 	return nil
// }

func splitAnnexB(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{0x00, 0x00, 0x00, 0x01}); i >= 0 {
		if i == 0 {
			// Skip the first start code
			advance, token, err = splitAnnexB(data[4:], atEOF)
			return advance + 4, token, err
		}
		return i, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func (s TrackerState) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateRun:
		return "Run"
	case StateCanceled:
		return "Canceled"
	default:
		return "Unknown"
	}
}
