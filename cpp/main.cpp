#include <fstream>
#include <sstream>
#include <iostream>

#include <opencv2/dnn.hpp>
#include <opencv2/imgproc.hpp>
#include <opencv2/highgui.hpp>

// --- Configuration Constants ---
const float INPUT_WIDTH = 640.0;
const float INPUT_HEIGHT = 640.0;
const float CONF_THRESHOLD = 0.50; // Minimum confidence to keep a box
const float NMS_THRESHOLD = 0.50;  // IoU threshold for Non-Maximum Suppression

// --- Global Variables ---
std::vector<std::string> class_names;

// --- Function Prototypes ---
void load_class_names(const std::string& path);
void process_predictions(cv::Mat& frame, cv::Mat& outs, const std::vector<cv::Scalar>& colors, int64& time_start);

// --------------------------- MAIN FUNCTION -----------------------------
int main() {
    // 1. Load Class Names
    load_class_names("../../resources/coco.names");
    if (class_names.empty()) {
        std::cerr << "ERROR: Could not load class names from coco.names!" << std::endl;
        return -1;
    }

    // 2. Load ONNX Model
    std::string module_path = "../../resources/yolov5n.onnx";
    cv::dnn::Net net = cv::dnn::readNetFromONNX(module_path);
    if (net.empty()) {
        std::cerr << "ERROR: Failed to load ONNX model!" << std::endl;
        return -1;
    }

    // Set backend and target for performance optimization (Crucial for TBB/CPU)
    net.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
    net.setPreferableTarget(cv::dnn::DNN_TARGET_CPU);

    // 3. Initialize Camera (0 for default webcam)
    cv::VideoCapture cap(0);
    if (!cap.isOpened()) {
        std::cerr << "ERROR: Could not open camera 0." << std::endl;
        return -1;
    }

    // 4. Detection Loop
    cv::Mat frame;
    std::vector<cv::Mat> outs;

    std::vector<cv::Scalar> colors;
    colors.push_back(cv::Scalar(0, 255, 0));
    colors.push_back(cv::Scalar(0, 255, 255));
    colors.push_back(cv::Scalar(255, 255, 0));
    colors.push_back(cv::Scalar(255, 0, 0));
    colors.push_back(cv::Scalar(0, 0, 255));

    while (cap.read(frame) && cv::waitKey(1) < 0) {
        int64 time_start = cv::getTickCount();
        // --- Pre-processing (Image to Blob) ---
        cv::Mat blob;
        cv::dnn::blobFromImage(frame, blob, 1/255.0, cv::Size(INPUT_WIDTH, INPUT_HEIGHT), cv::Scalar(), true, false);
        net.setInput(blob);

        // --- Inference (Forward Pass) ---
        net.forward(outs, net.getUnconnectedOutLayersNames());

        // --- Post-processing (NMS and Drawing) ---
        process_predictions(frame, outs[0], colors, time_start);

        // --- Display ---
        imshow("YOLOv5 C++ Detection (ThinkPad T14)", frame);
        outs.clear(); // Clear outputs for the next frame
    }

    cap.release();
    cv::destroyAllWindows();
    return 0;
}

// --------------------------- HELPER FUNCTIONS -----------------------------

void load_class_names(const std::string& path) {
    std::ifstream ifs(path.c_str());
    std::string line;
    while (getline(ifs, line)) {
        class_names.push_back(line);
    }
}

void process_predictions(cv::Mat& frame, cv::Mat& outs,
                         const std::vector<cv::Scalar>& colors, int64& time_start) {
    cv::Mat det_output(outs.size[1], outs.size[2], CV_32F, outs.ptr<float>());

    float confidence_threshold = 0.5;
    std::vector<cv::Rect> boxes;
    std::vector<int> classIds;
    std::vector<float> confidences;

    for (int i = 0; i < det_output.rows; i++) {
        float confidence = det_output.at<float>(i, 4);
        if (confidence < 0.25) {
            continue;
        }
        cv::Mat classes_scores = det_output.row(i).colRange(5, outs.size[2]);
        cv::Point class_id_point;
        double score;
        minMaxLoc(classes_scores, 0, &score, 0, &class_id_point);

        if (score > 0.25) {
            float x_factor = frame.cols / 640.0f;
            float y_factor = frame.rows / 640.0f;
            float cx = det_output.at<float>(i, 0);
            float cy = det_output.at<float>(i, 1);
            float ow = det_output.at<float>(i, 2);
            float oh = det_output.at<float>(i, 3);
            int x = static_cast<int>((cx - 0.5 * ow) * x_factor);
            int y = static_cast<int>((cy - 0.5 * oh) * y_factor);
            int width = static_cast<int>(ow * x_factor);
            int height = static_cast<int>(oh * y_factor);
            cv::Rect box;
            box.x = x;
            box.y = y;
            box.width = width;
            box.height = height;

            boxes.push_back(box);
            classIds.push_back(class_id_point.x);
            confidences.push_back(score);
        }
    }

    std::vector<int> indexes;
    cv::dnn::NMSBoxes(boxes, confidences, 0.25, 0.50, indexes);

    for (size_t i = 0; i < indexes.size(); i++) {
        int index = indexes[i];
        int idx = classIds[index];
        float confidence = confidences[index];

        cv::rectangle(frame, boxes[index], colors[idx % 5], 2, 8);

        std::string label = class_names[idx] + ": " + cv::format("%.2f", confidence);

        cv::rectangle(frame,
                      cv::Point(boxes[index].tl().x, boxes[index].tl().y - 40),
                      cv::Point(boxes[index].br().x, boxes[index].tl().y),
                      cv::Scalar(255, 255, 255), -1);

        cv::putText(frame, label,
                    cv::Point(boxes[index].tl().x, boxes[index].tl().y - 10),
                    cv::FONT_HERSHEY_SIMPLEX, 0.5, cv::Scalar());

        std::cout << "Detected: " << class_names[idx] << ", Confidence: " << confidence << std::endl;
    }

    float t = (cv::getTickCount() - time_start) / static_cast<float>(cv::getTickFrequency());

    cv::putText(frame, cv::format("FPS: %.2f", 1.0 / t), cv::Point(20, 40), cv::FONT_HERSHEY_PLAIN, 2.0, cv::Scalar(255, 0, 0), 2, 8);
}
