package main

import (
	"log"
	"net"
	"yolo-detector-service/controller"

	"google.golang.org/grpc"
)

// type TrackerServer struct {
// 	// Env      *bootstrap.Env
// 	// trackers map[string]interface{}
// 	// Required to be embedded for forward compatibility
// 	pb.UnimplementedTrackerServiceServer
// 	// pb.TrackerServiceServer
// 	lock         sync.Mutex
// 	sessionState string // "IDLE", "RECORDING"
// }

// func NewTrackerServer() *TrackerServer {
// 	return &TrackerServer{
// 		sessionState: "IDLE",
// 	}
// }

// func NewTrackerServer() *controller.TrackerServer {
// 	return &controller.TrackerServer{
// 		UnimplementedTrackerServiceServer: pb.UnimplementedTrackerServiceServer{},
// 		SessionState:                      "IDLE",
// 	}
// }

// func (s *TrackerServer) StreamUpdates(stream pb.TrackerService_StreamUpdatesServer) error {
// 	log.Println("New C++ client connected. Starting session watcher...")

// 	// --- Session State Management ---
// 	// Use a non-blocking channel for the cooldown timer
// 	cooldownTimer := time.NewTimer(0)
// 	cooldownTimer.Stop() // Stop immediately; will be reset on first activity

// 	// Main loop to continuously read messages from the C++ client
// 	for {
// 		update, err := stream.Recv()

// 		// Handle stream closure (Client calls WritesDone())
// 		if err == io.EOF {
// 			log.Println("C++ client stream finished. Shutting down session.")
// 			// Send the final status back to the C++ client
// 			success := true
// 			return stream.SendAndClose(&pb.StreamStatus{Success: &success})
// 		}
// 		if err != nil {
// 			log.Printf("Error receiving frame update: %v", err)
// 			return err
// 		}

// 		s.lock.Lock()

// 		// 1. Check for High-Priority Activity (e.g., Dog detected)
// 		// if update.GetHighPriorityActive() {
// 		// 	if s.sessionState == "IDLE" {
// 		// 		log.Println("-> STATE CHANGE: IDLE -> RECORDING (High priority event)")
// 		// 		s.sessionState = "RECORDING"
// 		// 		// Action: Start video archiving / web stream here
// 		// 	}
// 		// 	// Reset the cooldown timer whenever activity is detected
// 		// 	cooldownTimer.Reset(CooldownDuration)
// 		// }

// 		// 2. Check for Cooldown Expiration
// 		select {
// 		case <-cooldownTimer.C:
// 			// Timer fired: No activity detected for CooldownDuration
// 			if s.sessionState == "RECORDING" {
// 				log.Println("-> STATE CHANGE: RECORDING -> IDLE (Cooldown expired)")
// 				s.sessionState = "IDLE"
// 				// Action: Stop video archiving / web stream here
// 			}
// 		default:
// 			// Timer has not fired (or was just reset). Do nothing.
// 		}

// 		// Optional: Log the state and current event count for debugging
// 		log.Printf("Session: %s | Events: %d | Frame: %d",
// 			s.sessionState, len(update.GetEvents()), update.GetFrameNumber())

// 		s.lock.Unlock()
// 	}
// }

func main() {
	// if len(os.Args) != 2 {
	// 	logrus.Fatal("Failed, config path as argument is required")
	// }
	// app := bootstrap.App()
	// env := app.Env

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()

	controller.NewTrackerServer(s)
	// trackerServer := &controller.TrackerServer{}
	// pb.RegisterTrackerServiceServer(s, trackerServer)

	log.Printf("gRPC Server listening on %s", ":8081")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	// if len(os.Args) != 2 {
	// 	logrus.Fatal("Failed, config path as argument is required")
	// }
	// app := bootstrap.App()
	// env := app.Env

	// s := grpc.NewServer()

	// // lis, err := net.Listen("tcp", fmt.Sprintf("[::1]:%s", env.EVENT_SERVER_PORT))
	// const port = ":8081"
	// lis, err := net.Listen("tcp", port)
	// if err != nil {
	// 	logrus.Fatalf("Failed to listen: %v", err)
	// }
	// logrus.Printf("gRPC Server listening on %s", env.EVENT_SERVER_PORT)

	// // gMain := gin.New()
	// // gMain.Use(gin.Recovery())

	// trackerServer := &controller.TrackerServer{
	// 	Env: env,
	// }

	// // gMain.POST("/v1/test", trackerServer.TestMethod)

	// // go gMain.Run(env.REST_IP + ":" + env.REST_PORT)
	// // logrus.Printf("REST Server listening on %s", env.REST_PORT)

	// pb.RegisterTrackerServiceServer(s, trackerServer)

	// if err := s.Serve(lis); err != nil {
	// 	logrus.Fatalf("Failed to serve: %v", err)
	// }

	// defer app.Close()
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
