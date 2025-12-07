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
void process_predictions(cv::Mat& frame, const std::vector<cv::Mat>& outs);

// --------------------------- MAIN FUNCTION -----------------------------
int main() {
    // 1. Load Class Names
    load_class_names("coco.names");
    if (class_names.empty()) {
        std::cerr << "ERROR: Could not load class names from coco.names!" << std::endl;
        return -1;
    }

    // 2. Load ONNX Model
    std::string module_path = "/home/khomin/Documents/PROJECTS/YOLO_detector/models/yolov5n.onnx";
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

    while (cap.read(frame) && cv::waitKey(1) < 0) {
        // --- Pre-processing (Image to Blob) ---
        cv::Mat blob;
        cv::dnn::blobFromImage(frame, blob, 1/255.0, cv::Size(INPUT_WIDTH, INPUT_HEIGHT), cv::Scalar(), true, false);
        net.setInput(blob);

        // --- Inference (Forward Pass) ---
        net.forward(outs, net.getUnconnectedOutLayersNames());

        // --- Post-processing (NMS and Drawing) ---
        process_predictions(frame, outs);

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

void process_predictions(cv::Mat& frame, const std::vector<cv::Mat>& outs) {
    cv::Mat det_output(preds.size[1], preds.size[2], CV_32F, preds.ptr<float>());

    float confidence_threshold = 0.5;
    std::vector<cv::Rect> boxes;
    std::vector<int> classIds;
    std::vector<float> confidences;

    for (int i = 0; i < det_output.rows; i++) {
        float confidence = det_output.at<float>(i, 4);
        if (confidence < 0.25) {
            continue;
        }
        cv::Mat classes_scores = det_output.row(i).colRange(5, preds.size[2]);
        cv::Point class_id_point;
        double score;
        minMaxLoc(classes_scores, 0, &score, 0, &class_id_point);

        if (score > 0.25) {
            float x_factor = image.cols / 640.0f;
            float y_factor = image.rows / 640.0f;
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

    float t = (cv::getTickCount() - start) / static_cast<float>(cv::getTickFrequency());

    cv::putText(frame, cv::format("FPS: %.2f", 1.0 / t),
                cv::Point(20, 40), cv::FONT_HERSHEY_PLAIN, 2.0, cv::Scalar(255, 0, 0), 2, 8);

//    std::vector<int> classIds;
//    std::vector<float> confidences;
//    std::vector<cv::Rect> boxes;

//    // YOLOv5 output is a single tensor: [1, 25200, 85] -> Reshaped to rows x cols
//    const int rows = outs[0].size[1]; // Typically 25200 (number of candidate boxes)
//    const int cols = outs[0].size[2]; // 85 (5 box attributes + 80 classes)

//    // The entire output array is a single continuous memory block
//    float* data = (float*)outs[0].data;

//    // Scaling factors to map 640x640 prediction back to original image size
//    float x_factor = frame.cols / INPUT_WIDTH;
//    float y_factor = frame.rows / INPUT_HEIGHT;

//    const int num_classes = class_names.size(); // 80
//    const int row_size = 5 + num_classes; // 85

//    // Loop through all candidate prediction rows
//    for (int r = 0; r < outs[0].size[1]; ++r) {
//        // --- 1. Get Global Confidence ---
//        float box_confidence = data[r * row_size + 4]; // Confidence is at index 4

//        if (box_confidence >= CONF_THRESHOLD) {

//            // --- 2. Extract Class Scores (The Fix!) ---
//            // Create a Mat header pointing *only* to the class score data (80 values)
//            cv::Mat scores_mat(1, num_classes, CV_32F, data + (r * row_size + 5));

//            cv::Point classIdPoint;
//            double confidence;

//            // This now works because scores_mat is explicitly 1 row, 80 columns (dims=2 or less)
//            minMaxLoc(scores_mat, 0, &confidence, 0, &classIdPoint);

//            // 3. Class Score Threshold & Bounding Box Calculation
//            if (confidence >= 0.0) { // Can use a class-specific threshold here if desired
//                // ... (rest of your logic for bounding box and storing results) ...

//                // Example for getting bounding box data:
//                float cx = data[r * row_size + 0];
//                float cy = data[r * row_size + 1];
//                float w = data[r * row_size + 2];
//                float h = data[r * row_size + 3];

//                // Add this DEBUG CODE inside your loop before the final NMS:
//                std::cout << "Frame Size: " << frame.cols << "x" << frame.rows
//                     << " | Factors: " << x_factor << ", " << y_factor
//                     << " | Raw CX: " << cx << " | Scaled Left: " << std::left << std::endl;

//                // Convert to top-left (x,y) and (w,h)
//                float x_factor = frame.cols / INPUT_WIDTH;
//                float y_factor = frame.rows / INPUT_HEIGHT;
//                int left = (int)((cx - 0.5 * w) * x_factor);
//                int top = (int)((cy - 0.5 * h) * y_factor);
//                int width = (int)(w * x_factor);
//                int height = (int)(h * y_factor);
//                boxes.push_back(cv::Rect(left, top, width, height));


//                // The existing logic for classIds, confidences, and boxes goes here
//                classIds.push_back(classIdPoint.x);
//                confidences.push_back((float)confidence * box_confidence); // Use final confidence
//                boxes.push_back(cv::Rect(left, top, width, height));
//            }
//        }
//    }

//    // 2. Non-Maximum Suppression (NMS)
//    // NMSBoxes is an OpenCV function that handles the suppression logic efficiently.
//    std::vector<int> indices;
//    cv::dnn::NMSBoxes(boxes, confidences, CONF_THRESHOLD, NMS_THRESHOLD, indices);

//    // 3. Draw final detections
//    for (size_t i = 0; i < indices.size(); ++i) {
//        int idx = indices[i];
//        cv::Rect box = boxes[idx];

//        // Draw the box
//        rectangle(frame, box, cv::Scalar(0, 255, 0), 2);

//        // Draw the label
//        std::string label = class_names[classIds[idx]] + cv::format(":%.2f", confidences[idx]);
//        putText(frame, label, cv::Point(box.x, box.y - 5), cv::FONT_HERSHEY_SIMPLEX, 0.6, cv::Scalar(0, 255, 0), 2);
//    }
}
