#include <vector>

#include "config-cxx/Config.h"
#include "network_client.h"
#include "detector.h"

std::vector<std::string> load_coco_names(std::string names_string) {
    std::vector<std::string> coco_names;
    std::stringstream ss(names_string);
    std::string segment;

    while (std::getline(ss, segment, ',')) {
        segment.erase(0, segment.find_first_not_of(" \n\r\t"));
        segment.erase(segment.find_last_not_of(" \n\r\t") + 1);
        if (!segment.empty()) {
            coco_names.push_back(segment);
        }
    }
    return coco_names;
}

int main() {
    config::Config config;

    NetworkClient signal_client(
        config.get<std::string>("Networking.signal_ip"),
        config.get<int>("Networking.signal_port")
    );
    signal_client.startStreaming();

    Detector detector(
        load_coco_names(config.get<std::string>("CocoNames")),
        config.get<std::string>("YOLO.model_path")
    );

    detector.onFrameReady = [&](tracker::FrameUpdate& event) {
        bool success = signal_client.sendUpdate(event);
        if(!success) {
//            signal_client.stopStreaming();
            signal_client.startStreaming();
        }
    };
    detector.run();

    signal_client.stopStreaming();

    return 0;
}
