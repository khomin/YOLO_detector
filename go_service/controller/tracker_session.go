package controller

import (
	"io"
	"os/exec"
	"sync"
	"time"
	"yolo-detector-service/bootstrap"
	pb "yolo-detector-service/grpc/generated"

	"github.com/sirupsen/logrus"
)

type TrackerState int

const (
	StateIdle TrackerState = iota
	StateConfirmation
	StateRun
	StateCanceled
)

type TrackerSession struct {
	state        TrackerState
	timer        *time.Ticker
	creationTime time.Time
	doneChan     chan struct{}
	gstCmd       *exec.Cmd
	gstIn        io.WriteCloser
	lock         sync.Mutex
	env          *bootstrap.Env
}

func (cc *TrackerSession) startSession(addr string, stream pb.TrackerService_StreamUpdatesServer) error {
	cc.state = StateConfirmation
	cc.creationTime = time.Now()
	cc.timer = time.NewTicker(time.Duration(cc.env.SESSION_TIMER_SEC) * time.Second)
	go func() {
		for {
			select {
			case <-cc.timer.C:
				logrus.Printf("[%s] Session timer ticked.", addr)
				switch cc.state {
				case StateConfirmation:
					startDelay := cc.creationTime.Add(time.Duration(cc.env.SESSION_START_DELAY_SEC) * time.Second)
					if time.Now().After(startDelay) {
						cc.state = StateRun
						if err := cc.startGStreamerPipeline(addr); err != nil {
							logrus.Printf("Failed to start GStreamer for %s: %v", addr, err)
							return
						}
					}
				case StateRun:
					// TODO: collect data and sent pushes
					break
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

func (cc *TrackerSession) processUpdate(update *pb.FrameUpdate) {
	frame := update.EncodedFrame
	for _, event := range update.Events {
		logrus.Printf("Event: classId=%d, className=%s", *event.ClassId, *event.ClassName)
	}
	if len(frame) > 0 {
		if cc.state == StateRun {
			logrus.Printf("Received frame: %d, num=%d", len(frame), *update.FrameNumber)
			n, err := cc.gstIn.Write(frame)
			if err != nil {
				logrus.Errorf("Error writing frame to GStreamer stdin: %v", err)
				// You may want to kill the GStreamer process and/or session here
				return
			}
			if n != len(frame) {
				logrus.Warnf("Incomplete write to GStreamer: Wrote %d of %d bytes", n, len(frame))
			}
		}
	} else {
		logrus.Print("No Frame")
	}
}

func (cc *TrackerSession) closeSession() {
	logrus.Println("Stopping GStreamer...")
	cc.gstIn.Close()
	// Wait for the process to exit naturally.
	err := cc.gstCmd.Wait()
	if err != nil {
		logrus.Printf("GStreamer exited with error: %v", err)
	}
	logrus.Println("GStreamer finished and file saved.")
}
