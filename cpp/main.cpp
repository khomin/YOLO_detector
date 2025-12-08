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

int frame_count = 0;
std::vector<cv::Rect> cached_boxes;
std::vector<int> cached_class_ids;
std::vector<float> cached_confidences;
const int INFERENCE_SKIP = 4;

// --- Global Variables ---
std::vector<std::string> class_names;

// --- Function Prototypes ---
void load_class_names(const std::string& path);
void process_predictions(cv::Mat& frame, cv::Mat& outs, const std::vector<cv::Scalar>& colors, int64& time_start,
                         std::vector<cv::Rect>& boxes,
                         std::vector<int>& class_ids,
                         std::vector<float>& confidences);

void draw_cached_boxes(cv::Mat& frame,
                       const std::vector<cv::Scalar>& colors,
                       int64& time_start,
                       std::vector<cv::Rect>& boxes,
                       std::vector<int>& class_ids,
                       std::vector<float>& confidences);

// --------------------------- MAIN FUNCTION -----------------------------
int main() {
    // 1. Load Class Names
    load_class_names("./resources/coco.names");
    if (class_names.empty()) {
        std::cerr << "ERROR: Could not load class names from coco.names!" << std::endl;
        return -1;
    }

    // 2. Load ONNX Model
    std::string module_path = "./resources/yolov5n.onnx";
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
    cap.set(cv::CAP_PROP_FRAME_WIDTH, 640);
    cap.set(cv::CAP_PROP_FRAME_HEIGHT, 480);
    if (!cap.isOpened()) {
        std::cerr << "ERROR: Could not open camera 0." << std::endl;
        return -1;
    }

    // 4. Detection Loop
    cv::Mat frame;

    std::vector<cv::Scalar> colors;
    colors.push_back(cv::Scalar(0, 255, 0));
    colors.push_back(cv::Scalar(0, 255, 255));
    colors.push_back(cv::Scalar(255, 255, 0));
    colors.push_back(cv::Scalar(255, 0, 0));
    colors.push_back(cv::Scalar(0, 0, 255));

    while (cap.read(frame) && cv::waitKey(1) < 0) {
        int64 time_start = cv::getTickCount();

        if (frame_count % INFERENCE_SKIP == 0) {
            std::vector<cv::Mat> outs;
            cached_boxes.clear();
            cached_class_ids.clear();
            cached_confidences.clear();

            // --- Pre-processing (Image to Blob) ---
            cv::Mat blob;
            cv::dnn::blobFromImage(frame, blob, 1/255.0, cv::Size(INPUT_WIDTH, INPUT_HEIGHT), cv::Scalar(), true, false);
            net.setInput(blob);

            // --- Inference (Forward Pass) ---
            net.forward(outs, net.getUnconnectedOutLayersNames());

            // --- Post-processing (NMS and Drawing) ---
            process_predictions(frame, outs[0], colors, time_start,
                                cached_boxes, cached_class_ids, cached_confidences);
            outs.clear();
        } else {
            draw_cached_boxes(frame, colors, time_start, cached_boxes, cached_class_ids, cached_confidences);
        }
        frame_count++; // Increment the counter

        // --- Display ---
        imshow("YOLOv5 C++ Detection (ThinkPad T14)", frame);
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
                         const std::vector<cv::Scalar>& colors,
                         int64& time_start,
                         std::vector<cv::Rect>& boxes,
                         std::vector<int>& class_ids,
                         std::vector<float>& confidences) {
    cv::Mat det_output(outs.size[1], outs.size[2], CV_32F, outs.ptr<float>());
    float confidence_threshold = 0.5;

    std::vector<cv::Rect> temp_boxes;
    std::vector<int> temp_class_ids;
    std::vector<float> temp_confidences;

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

            temp_boxes.push_back(box);
            temp_class_ids.push_back(class_id_point.x);
            temp_confidences.push_back(score);
        }
    }

    std::vector<int> indexes;
    cv::dnn::NMSBoxes(temp_boxes, temp_confidences, 0.25, 0.50, indexes);

    for (size_t i = 0; i < indexes.size(); i++) {
        int index = indexes[i];
        int idx = temp_class_ids[index];
        float confidence = temp_confidences[index];

        boxes.push_back(temp_boxes[index]);
        class_ids.push_back(temp_class_ids[index]);
        confidences.push_back(temp_confidences[index]);

        cv::rectangle(frame, temp_boxes[index], colors[idx % 5], 2, 8);

        std::string label = class_names[idx] + ": " + cv::format("%.2f", confidence);

        cv::rectangle(frame,
                      cv::Point(temp_boxes[index].tl().x, temp_boxes[index].tl().y - 40),
                      cv::Point(temp_boxes[index].br().x, temp_boxes[index].tl().y),
                      cv::Scalar(255, 255, 255), -1);

        cv::putText(frame, label,
                    cv::Point(temp_boxes[index].tl().x, temp_boxes[index].tl().y - 10),
                    cv::FONT_HERSHEY_SIMPLEX, 0.5, cv::Scalar());

        std::cout << "Detected: " << class_names[idx] << ", Confidence: " << confidence << std::endl;
    }

    float t = (cv::getTickCount() - time_start) / static_cast<float>(cv::getTickFrequency());
    cv::putText(frame, cv::format("FPS: %.2f", 1.0 / t), cv::Point(20, 40), cv::FONT_HERSHEY_PLAIN, 2.0, cv::Scalar(255, 0, 0), 2, 8);
}

void draw_cached_boxes(cv::Mat& frame,
                       const std::vector<cv::Scalar>& colors,
                       int64& time_start,
                       std::vector<cv::Rect>& boxes,
                       std::vector<int>& class_ids,
                       std::vector<float>& confidences) {
    int index = 0;
    for(auto box: boxes) {
        int idx = class_ids[index];
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
        index++;
    }
    float t = (cv::getTickCount() - time_start) / static_cast<float>(cv::getTickFrequency());
    cv::putText(frame, cv::format("FPS: %.2f", 1.0 / t), cv::Point(20, 40), cv::FONT_HERSHEY_PLAIN, 2.0, cv::Scalar(255, 0, 0), 2, 8);
}
