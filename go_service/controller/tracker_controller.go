package controller

import (
	"errors"
	"net/http"
	"sync"
	"yolo-detector-service/bootstrap"
	pb "yolo-detector-service/grpc/generated"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/peer"
)

type TrackerServer struct {
	Env            *bootstrap.Env
	Trackers       map[string]*TrackerSession
	lock           sync.Mutex
	sessionCounter int
	// Required to be embedded for forward compatibility
	pb.UnimplementedTrackerServiceServer
}

func (s *TrackerServer) StreamUpdates(stream pb.TrackerService_StreamUpdatesServer) error {
	p, ok := peer.FromContext(stream.Context())
	if !ok {
		return errors.New("peer information unavailable")
	}
	addr := p.Addr.String()
	logrus.Printf("New C++ client connected [%s]", addr)

	s.lock.Lock()
	session, ok := s.Trackers[addr]
	if ok {
		s.lock.Unlock()
		return nil
	}
	s.sessionCounter = s.sessionCounter + 1
	session = &TrackerSession{
		doneChan:  make(chan struct{}),
		sessionId: s.sessionCounter,
		env:       s.Env,
		trackerTime: TrackerTime{
			env: s.Env,
		},
	}
	s.Trackers[addr] = session
	s.lock.Unlock()
	defer func() {
		s.lock.Lock()
		delete(s.Trackers, addr)
		s.lock.Unlock()
		session.closeSession()
	}()
	session.startSession(addr, stream)
	return nil

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
