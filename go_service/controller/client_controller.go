package controller

import (
	"io"
	"log"
	"sync"
	"time"
	pb "yolo-detector-service/grpc/generated"

	"google.golang.org/grpc"
)

// type ClientControlller struct {
// 	Env *bootstrap.Env
// }

type TrackerServer struct {
	// Env      *bootstrap.Env
	// trackers map[string]interface{}
	// Required to be embedded for forward compatibility
	pb.UnimplementedTrackerServiceServer
	// pb.TrackerServiceServer
	lock         sync.Mutex
	SessionState string // "IDLE", "RECORDING"
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

func NewTrackerServer(s *grpc.Server) *TrackerServer {
	server := &TrackerServer{
		UnimplementedTrackerServiceServer: pb.UnimplementedTrackerServiceServer{},
		SessionState:                      "IDLE",
	}
	pb.RegisterTrackerServiceServer(s, server)
	return server
}

func (s *TrackerServer) StreamUpdates(stream pb.TrackerService_StreamUpdatesServer) error {
	log.Println("New C++ client connected. Starting session watcher...")

	// --- Session State Management ---
	// Use a non-blocking channel for the cooldown timer
	cooldownTimer := time.NewTimer(0)
	cooldownTimer.Stop() // Stop immediately; will be reset on first activity

	// Main loop to continuously read messages from the C++ client
	for {
		update, err := stream.Recv()

		// Handle stream closure (Client calls WritesDone())
		if err == io.EOF {
			log.Println("C++ client stream finished. Shutting down session.")
			// Send the final status back to the C++ client
			success := true
			return stream.SendAndClose(&pb.StreamStatus{Success: &success})
		}
		if err != nil {
			log.Printf("Error receiving frame update: %v", err)
			return err
		}

		s.lock.Lock()

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

		// Optional: Log the state and current event count for debugging
		log.Printf("Session: %s | Events: %d | Frame: %d",
			s.SessionState, len(update.GetEvents()), update.GetFrameNumber())

		s.lock.Unlock()
	}
}

// func (cc *TrackerServer) TestMethod(c *gin.Context) {
// 	response := map[string]interface{}{
// 		"success": true,
// 	}
// 	c.JSON(http.StatusOK, response)
// }

// func (cc *TrackerServer) OnTrackUpdate(c *gin.Context) {
// 	response := map[string]interface{}{
// 		"success": true,
// 	}
// 	c.JSON(http.StatusOK, response)
// }
