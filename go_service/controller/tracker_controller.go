package controller

import (
	"errors"
	"io"
	"log"
	"net/http"
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
	time     *time.Timer
	timer    *time.Ticker
	doneChan chan struct{}
	lock     sync.Mutex
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
	} else {
		logrus.Print("No Frame")
	}
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
