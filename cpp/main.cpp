#include <vector>

#include "config-cxx/Config.h"
//#include "protobuf/generated/tracking_events.pb.h"

#include "network_client.h"
#include "detector.h"

int main() {
    config::Config config;

    NetworkClient signal_client(
        config.get<std::string>("Networking.signal_ip"),
        config.get<int>("Networking.signal_port")
        );


    Detector detector(
        config.get<std::vector<std::string>>("CocoNames"),
        config.get<std::string>("YOLO.model_path")
        );

    detector.onFrameReady = [&](tracker::TrackEvent* event) {
        signal_client.add(event);
    };
    detector.run();

    return 0;
}
