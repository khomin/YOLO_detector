package main

import (
	"net"
	"os"
	"yolo-detector-service/bootstrap"
	"yolo-detector-service/controller"
	pb "yolo-detector-service/grpc/generated"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) != 2 {
		logrus.Fatal("Failed, config path as argument is required")
	}
	app := bootstrap.App()
	env := app.Env

	s := grpc.NewServer()

	// lis, err := net.Listen("tcp", fmt.Sprintf("[::1]:%s", env.EVENT_SERVER_PORT))
	const port = ":8081"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logrus.Fatalf("Failed to listen: %v", err)
	}
	logrus.Printf("gRPC Server listening on %s", env.EVENT_SERVER_PORT)

	// gMain := gin.New()
	// gMain.Use(gin.Recovery())

	trackerServer := &controller.TrackerServer{
		Env: env,
	}

	// gMain.POST("/v1/test", trackerServer.TestMethod)

	// go gMain.Run(env.REST_IP + ":" + env.REST_PORT)
	// logrus.Printf("REST Server listening on %s", env.REST_PORT)

	pb.RegisterTrackerServiceServer(s, trackerServer)

	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("Failed to serve: %v", err)
	}

	defer app.Close()
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
