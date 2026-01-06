#ifndef ASYNCNETWORKCLIENT_H
#define ASYNCNETWORKCLIENT_H

#include <string>
#include <thread>
#include "safe_queue.h"
#include "protobuf/generated/tracker.pb.h"
#include "protobuf/generated/tracker.grpc.pb.h"
#include <grpcpp/grpcpp.h>

class AsyncNetworkClient {
public:
    AsyncNetworkClient(std::string target_address) : target_(target_address), running_(true) {
        grpc::ChannelArguments args;
        // High-performance settings: fix the "slow reconnect"
        args.SetInt(GRPC_ARG_KEEPALIVE_TIME_MS, 5000);
        args.SetInt(GRPC_ARG_KEEPALIVE_TIMEOUT_MS, 2000);
        args.SetInt(GRPC_ARG_MIN_RECONNECT_BACKOFF_MS, 1000);
        args.SetInt(GRPC_ARG_MAX_RECONNECT_BACKOFF_MS, 3000);

        channel_ = grpc::CreateCustomChannel(target_, grpc::InsecureChannelCredentials(), args);
        stub_ = tracker::TrackerService::NewStub(channel_);

        // Start the background sender thread
        worker_thread_ = std::thread(&AsyncNetworkClient::sendLoop, this);
    }

    ~AsyncNetworkClient() {
        running_ = false;
        queue_.request_shutdown();
        if (worker_thread_.joinable()) {
            worker_thread_.join();
        }
    }

    void queueUpdate(tracker::FrameUpdate update) {
        queue_.push(std::move(update));
    }

private:
    void sendLoop() {
        while (running_) {
            auto context = std::make_unique<grpc::ClientContext>();
            tracker::StreamStatus response;
            auto writer = stub_->StreamUpdates(context.get(), &response);

            if (!writer) {
                std::this_thread::sleep_for(std::chrono::seconds(1));
                continue; // Retry connection
            }

            while (running_) {
                auto update = queue_.pop();
                if(update.has_value()) {
                    if (!writer->Write(update.value())) {
                        std::cerr << "Stream broke. Attempting reconnect..." << std::endl;
                        break; // Exit inner loop to trigger reconnect
                    }
                }
            }
            writer->WritesDone();
            writer->Finish();
        }
    }

    std::string target_;
    bool running_;
    SafeQueue<tracker::FrameUpdate> queue_;
    std::thread worker_thread_;
    std::shared_ptr<grpc::Channel> channel_;
    std::unique_ptr<tracker::TrackerService::Stub> stub_;
};
#endif // ASYNCNETWORKCLIENT_H
