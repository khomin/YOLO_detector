package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"yolo-detector-service/bootstrap"
	"yolo-detector-service/controller"

	pb "yolo-detector-service/grpc/generated"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) != 2 {
		logrus.Fatal("Failed, config path as argument is required")
	}
	app := bootstrap.App()
	env := app.Env
	defer app.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ---- gRPC ----
	trackerListen, err := net.Listen("tcp", env.EVENT_SERVER_IP+":"+env.EVENT_SERVER_PORT)
	if err != nil {
		logrus.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	tracker := &controller.TrackerServer{
		UnimplementedTrackerServiceServer: pb.UnimplementedTrackerServiceServer{},
		Env:                               env,
		Trackers:                          make(map[string]*controller.TrackerSession),
	}
	pb.RegisterTrackerServiceServer(grpcServer, tracker)

	go func() {
		logrus.Printf("gRPC Server listening on %s", env.EVENT_SERVER_PORT)
		err := grpcServer.Serve(trackerListen)
		if err != nil {
			logrus.Fatalf("Failed to serve: %v", err)
		}
	}()

	// ---- REST ----
	router := gin.New()
	router.Use(gin.Recovery())

	router.POST("/v1/test", tracker.TestMethod)

	logrus.Printf("REST Server listening on %s", env.REST_PORT)

	err = router.Run(env.REST_IP + ":" + env.REST_PORT)
	if err != nil {
		logrus.Fatalf("Failed to serve: %v", err)
	}

	<-ctx.Done()
	logrus.Info("Shutting down...")
	grpcServer.GracefulStop()
}
