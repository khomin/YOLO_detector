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
	state TrackerState
	lock  sync.Mutex
}

type TrackerServer struct {
	Env      *bootstrap.Env
	Trackers map[string]TrackerSession
	lock     sync.Mutex
	// Required to be embedded for forward compatibility
	pb.UnimplementedTrackerServiceServer
}

// The duration the server waits after high-priority_active turns false
const CooldownDuration = 10 * time.Second

// stream server
// - register streams
// - control its status
//		- idle, run, timeout
// - write temp file
// - put frames to temp file
// - add record to database when status - canceled
// - delete temp file

func (s *TrackerServer) StreamUpdates(stream pb.TrackerService_StreamUpdatesServer) error {
	log.Println("New C++ client connected. Starting session watcher...")

	// --- Session State Management ---
	// Use a non-blocking channel for the cooldown timer
	cooldownTimer := time.NewTimer(0)
	cooldownTimer.Stop() // Stop immediately; will be reset on first activity

	p, ok := peer.FromContext(stream.Context())
	if !ok {
		return errors.New("peer information unavailable")
	}

	addr := p.Addr.String()

	s.lock.Lock()
	defer s.lock.Unlock()
	session, ok := s.Trackers[addr]
	if !ok {
		session = TrackerSession{}
		s.Trackers[addr] = session
	}

	// Main loop to continuously read messages from the C++ client
	for {
		update, err := stream.Recv()

		if err == io.EOF {
			log.Println("C++ client stream finished. Shutting down session.")
			success := true
			return stream.SendAndClose(&pb.StreamStatus{Success: &success})
		}
		if err != nil {
			log.Printf("Error receiving frame update: %v", err)
			return err
		}

		s.lock.Lock()
		defer s.lock.Unlock()

		// 2. Check for Cooldown Expiration
		select {
		case <-cooldownTimer.C:
			// Timer fired: No activity detected for CooldownDuration
			if s.SessionState == "RECORDING" {
				log.Println("-> STATE CHANGE: RECORDING -> IDLE (Cooldown expired)")
				s.SessionState = "IDLE"
				// Action: Stop video archiving / web stream here
			}
		default:
			// Timer has not fired (or was just reset). Do nothing.
		}

		events := update.GetEvents()
		logrus.Print(events)

		// events[0].
		// update.

		update.GetFrameNumber()

		// Optional: Log the state and current event count for debugging
		// log.Printf("Session: %s | Events: %d | Frame: %d",
		// 	s.SessionState, len(update.GetEvents()), update.GetFrameNumber())

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
