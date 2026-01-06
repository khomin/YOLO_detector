package controller

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
	"yolo-detector-service/bootstrap"
	pb "yolo-detector-service/grpc/generated"

	"github.com/pion/webrtc/v4"
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
	s.startWebRtcPipeline()
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
				switch cc.state {
				case StateIdle:
					cc.lock.Lock()
					if cc.trackerTime.hasTargetFor(cc.env.TARGET_THRESHOLD_DURATION) {
						cc.trackerTime.clear()
						cc.startEventPipeline()
						cc.state = StateRun
					}
					cc.lock.Unlock()
				case StateRun:
					cc.lock.Lock()
					if cc.trackerTime.noTargetFor(cc.env.TARGET_THRESHOLD_DURATION) {
						cc.trackerTime.clear()
						cc.stopEventPipeline()
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
	cc.stopEventPipeline()
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
	if cc.gstIn == nil {
		return fmt.Errorf("errror wriring, stdin is nil")
	}
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
