#ifndef NETWORKCLIENT_H
#define NETWORKCLIENT_H

#include <string>
#include "protobuf/generated/tracker.pb.h"
#include <grpcpp/grpcpp.h>

class NetworkClient {
public:
    NetworkClient(std::string ip, int port);

    bool StartStreaming();
    bool sendUpdate(const tracker::FrameUpdate& update);
    void StopStreaming();

    int add(tracker::TrackEvent* event);

private:
    // gRPC objects
    std::shared_ptr<grpc::Channel> channel_;
    std::unique_ptr<tracker::TrackerService::Stub> stub_;

    // The context holds metadata, deadlines, etc., for the RPC.
    grpc::ClientContext context_;

    // The stream writer object, used to push messages to the server.
    // W is the message type (FrameUpdate) and R is the response type (StreamStatus).
    std::unique_ptr<grpc::ClientWriter<tracker::FrameUpdate>> client_writer_;

    // The object that will hold the final response from the Go server.
    tracker::StreamStatus status_response_;
};

#endif // NETWORKCLIENT_H
