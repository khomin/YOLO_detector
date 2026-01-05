package controller

import (
	"errors"
	"net/http"
	"sync"
	"yolo-detector-service/bootstrap"
	pb "yolo-detector-service/grpc/generated"

	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v4"
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
	session = InitSession(s.sessionCounter, s.Env)
	s.Trackers[addr] = session
	s.sessionCounter = s.sessionCounter + 1
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

func (s *TrackerServer) HandleSignaling(c *gin.Context) {
	// 1. Get the ID from the URL
	sessionID := c.Param("id")

	// 2. Lock the server map to find the session safely
	s.lock.Lock()
	session, exists := s.Trackers[sessionID]
	s.lock.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// 3. Parse the SDP Offer from Android
	var offer webrtc.SessionDescription
	if err := c.ShouldBindJSON(&offer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SDP"})
		return
	}

	// Create a MediaEngine to support H.264 (common for YOLO/FFmpeg streams)
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		panic(err)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	// pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 5. Add the track from THIS specific session
	_, err = pc.AddTrack(session.videoTrack)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not add track"})
		return
	}

	// 6. Set Remote, Create Answer, Set Local
	pc.SetRemoteDescription(offer)
	answer, _ := pc.CreateAnswer(nil)

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if state == webrtc.PeerConnectionStateDisconnected ||
			state == webrtc.PeerConnectionStateFailed ||
			state == webrtc.PeerConnectionStateClosed {
			pc.Close()
			logrus.Infof("PeerConnection for session %s closed", sessionID)
		}
	})

	// Use the GatheringCompletePromise to ensure the answer has all ICE info
	gatherComplete := webrtc.GatheringCompletePromise(pc)
	pc.SetLocalDescription(answer)
	<-gatherComplete

	// 7. Return the final Answer
	c.JSON(http.StatusOK, pc.LocalDescription())
}

func (s *TrackerServer) GetSessions(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()

	var activeSessions []SessionInfo

	// Iterate through your map and build the response list
	for id, session := range s.Trackers {
		activeSessions = append(activeSessions, SessionInfo{
			ID:    id,
			State: session.state.String(),
		})
	}

	// Return the list as a JSON array
	c.JSON(http.StatusOK, activeSessions)
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
