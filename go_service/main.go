// package main
package main

import (
	"io"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	// IMPORTANT: Update this import path to where your generated tracker.pb.go lives
	// pb "grpc/generated/your/project/go_service/grpc/generated/tracker_pb"
	pb "grpc/generated/your/project/go_service/grpc/generated/tracker_pb"
)

// The duration the server waits after high-priority_active turns false
const CooldownDuration = 10 * time.Second

// TrackerServer implements the generated gRPC service interface
type TrackerServer struct {
	// Required to be embedded for forward compatibility
	pb.UnimplementedTrackerServiceServer

	mu           sync.Mutex // Mutex to protect shared state access
	sessionState string     // "IDLE", "RECORDING"
}

func NewTrackerServer() *TrackerServer {
	return &TrackerServer{
		sessionState: "IDLE",
	}
}

// -------------------------------------------------------------
// Core Business Logic: The Client Streaming Handler
// -------------------------------------------------------------

// StreamUpdates implements the Client Streaming RPC method.
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
			return stream.SendAndClose(&pb.StreamStatus{Success: true})
		}
		if err != nil {
			log.Printf("Error receiving frame update: %v", err)
			return err
		}

		s.mu.Lock()

		// 1. Check for High-Priority Activity (e.g., Dog detected)
		if update.GetHighPriorityActive() {
			if s.sessionState == "IDLE" {
				log.Println("-> STATE CHANGE: IDLE -> RECORDING (High priority event)")
				s.sessionState = "RECORDING"
				// Action: Start video archiving / web stream here
			}
			// Reset the cooldown timer whenever activity is detected
			cooldownTimer.Reset(CooldownDuration)
		}

		// 2. Check for Cooldown Expiration
		select {
		case <-cooldownTimer.C:
			// Timer fired: No activity detected for CooldownDuration
			if s.sessionState == "RECORDING" {
				log.Println("-> STATE CHANGE: RECORDING -> IDLE (Cooldown expired)")
				s.sessionState = "IDLE"
				// Action: Stop video archiving / web stream here
			}
		default:
			// Timer has not fired (or was just reset). Do nothing.
		}

		// Optional: Log the state and current event count for debugging
		log.Printf("Session: %s | Events: %d | Frame: %d",
			s.sessionState, len(update.GetDetections()), update.GetFrameId())

		s.mu.Unlock()
	}
}

// -------------------------------------------------------------
// Main Function
// -------------------------------------------------------------

func main() {
	// 1. Setup Listener
	const port = ":8081"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 2. Create gRPC Server
	s := grpc.NewServer()

	// 3. Register our service implementation
	trackerServer := NewTrackerServer()
	pb.RegisterTrackerServiceServer(s, trackerServer)

	log.Printf("gRPC Server listening on %s", port)
	// Run the server (this is a blocking call)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// import (
// 	"net"

// 	"example.com/app/bootstrap"
// )

// func main() {
// 	app := bootstrap.App()
// 	// env := app.Env

// 	lis, err := net.Listen("tcp", ":8081") // Or the specified address
// 	// ...
// 	s.Serve(lis)

// 	// gMain := gin.New()
// 	// gMain.Use(cors.AllowAll())
// 	// gMain.Use(gin.Recovery())

// 	// promet := ginprometheus.NewPrometheus("gin")
// 	// promet.Use(gMain)

// 	// route.Setup(env, app.Db, gMain)

// 	// channel_ = grpc::CreateChannel(target_address, grpc::InsecureChannelCredentials());

// 	// go gMain.Run(env.DASHBOARD_ADDR + ":" + env.DASHBOARD_API_PORT)

// 	// defer app.FreeResources()
// }
