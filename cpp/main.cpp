#include <vector>

#include "config-cxx/Config.h"

#include "network_client.h"
#include "detector.h"

int main() {
    config::Config config;

    NetworkClient signal_client(
        config.get<std::string>("Networking.signal_ip"),
        config.get<int>("Networking.signal_port")
    );
    signal_client.startStreaming();

    Detector detector(
        config.get<std::vector<std::string>>("CocoNames"),
        config.get<std::string>("YOLO.model_path")
    );

    detector.onFrameReady = [&](tracker::FrameUpdate& event) {
        signal_client.sendUpdate(event);
    };
    detector.run();

    signal_client.stopStreaming();

    return 0;
}
