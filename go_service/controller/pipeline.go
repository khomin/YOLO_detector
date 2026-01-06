package controller

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/sirupsen/logrus"
)

func (s *TrackerSession) startEventPipeline() error {
	fileName := fmt.Sprintf("session_%04d_%03d.mp4", s.sessionId, s.recordCount)
	path := path.Join(s.env.RECORDINGS_TMP_DIR, fileName)
	os.MkdirAll(s.env.RECORDINGS_TMP_DIR, os.ModePerm)
	s.recordCount += 1

	switch s.env.VIDEO_CODEC {
	case "H264":
		s.gstCmd = exec.Command("gst-launch-1.0", []string{
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
		}...)
	case "H265":
		s.gstCmd = exec.Command("gst-launch-1.0", []string{
			"fdsrc", "do-timestamp=true",
			"!", "image/jpeg",
			"!", "jpegparse",
			"!", "jpegdec",
			"!", "videoconvert",
			"!", "videorate",
			"!", "video/x-raw,framerate=30/1",
			"!", "x265enc", "speed-preset=ultrafast", "bitrate=2000",
			"!", "h265parse",
			"!", "mp4mux", "fragment-duration=2000",
			"!", "filesink",
			"location=" + path, "sync=false",
		}...)
	default:
		return fmt.Errorf("unsupported video codec: %s", s.env.VIDEO_CODEC)
	}
	// get STDIN pipe (for sending data to gstreamer)
	var err error
	s.gstIn, err = s.gstCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	s.gstCmd.Stderr = os.Stderr
	s.gstCmd.Stdout = os.Stdout // Optional: keep stdout visible

	// start the gstreamer pipeline process
	if err := s.gstCmd.Start(); err != nil {
		return fmt.Errorf("failed to start gst-launch: %w", err)
	}
	logrus.Printf("gstreamer pipeline started")
	return nil
}

func (s *TrackerSession) stopEventPipeline() error {
	if s.gstIn != nil {
		s.gstIn.Close()
	}
	if s.gstCmd != nil {
		err := s.gstCmd.Wait()
		if err != nil {
			logrus.Printf("gstreamer exited with error: %v", err)
		}
	}
	logrus.Println("gstreamer finished")
	return nil
}

func (cc *TrackerSession) startWebRtcPipeline() error {
	webrtcMimeType := ""
	switch cc.env.VIDEO_CODEC {
	case "H264":
		cc.gstWebRtcCmd = exec.Command("gst-launch-1.0", []string{
			"fdsrc", "do-timestamp=true",
			"!", "image/jpeg",
			"!", "jpegparse",
			"!", "jpegdec",
			"!", "videoconvert",
			"!", "video/x-raw,format=I420",
			"!", "x264enc", "bitrate=2000", "tune=zerolatency", "speed-preset=ultrafast", "sliced-threads=false", "key-int-max=15",
			"!", "video/x-h264,profile=baseline,stream-format=byte-stream",
			"!", "h264parse", "config-interval=-1",
			"!", "video/x-h264,stream-format=byte-stream,alignment=au", // 'au' means Access Unit (Full Frame)
			"!", "fdsink", "fd=1", "sync=false",
		}...)
		webrtcMimeType = webrtc.MimeTypeH264
	case "H265":
		cc.gstWebRtcCmd = exec.Command("gst-launch-1.0", []string{
			"fdsrc", "do-timestamp=true",
			"!", "image/jpeg",
			"!", "jpegparse",
			"!", "jpegdec",
			"!", "videoconvert",
			"!", "video/x-raw,format=I420",
			"!", "x265enc", "speed-preset=ultrafast", "bitrate=2000", "key-int-max=15",
			"!", "video/x-h265,profile=baseline,stream-format=byte-stream",
			"!", "h265parse", "config-interval=-1",
			"!", "video/x-h265,stream-format=byte-stream,alignment=au", // 'au' means Access Unit (Full Frame)
			"!", "fdsink", "fd=1", "sync=false",
		}...)
		webrtcMimeType = webrtc.MimeTypeH265
	default:
		return fmt.Errorf("unsupported video codec: %s", cc.env.VIDEO_CODEC)
	}
	cc.gstWebRtcIn, _ = cc.gstWebRtcCmd.StdinPipe()
	cc.gstWebRtcOut, _ = cc.gstWebRtcCmd.StdoutPipe()
	cc.gstWebRtcCmd.Stderr = os.Stderr

	cc.videoTrack, _ = webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{
			MimeType:    webrtcMimeType,
			SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
		},
		"video",
		"pion",
	)
	if err := cc.gstWebRtcCmd.Start(); err != nil {
		return err
	}
	// reader Loop
	go func() {
		scanner := bufio.NewScanner(cc.gstWebRtcOut)
		buf := make([]byte, 0, 1024*1024)
		scanner.Buffer(buf, 1024*1024)
		scanner.Split(splitAnnexB)
		var headerStack []byte // to store SPS/PPS until a real frame arrives

		for scanner.Scan() {
			data := scanner.Bytes()
			if len(data) == 0 {
				continue
			}
			// if it's a small packet (SPS/PPS/Metadata), save it.
			if len(data) < 100 {
				headerStack = append(headerStack, []byte{0x00, 0x00, 0x00, 0x01}...)
				headerStack = append(headerStack, data...)
				continue
			}
			// if it's a big packet (Video Frame), attach any saved headers and send.
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
		logrus.Info("scanner goroutine exited")
	}()
	return nil
}

func (s *TrackerSession) stopWebRtcPipeline() error {
	if s.gstWebRtcIn != nil {
		s.gstWebRtcIn.Close()
	}
	if s.gstWebRtcCmd != nil {
		err := s.gstWebRtcCmd.Wait()
		if err != nil {
			logrus.Printf("gstreamer exited with error: %v", err)
		}
	}
	logrus.Println("gstreamer finished")
	return nil
}

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
