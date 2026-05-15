import os
import time
from pathlib import Path
from typing import Any, Dict, List, Optional

import cv2
import numpy as np
import torch

os.environ.setdefault(
    "YOLO_CONFIG_DIR",
    str(Path(__file__).resolve().parent / ".ultralytics"),
)

from ultralytics import YOLO


DEFAULT_MODEL_PATH = Path("runs/detect/thermal_detector_v2/weights/best.pt")
MODEL_PATH_ENV = "THERMAL_MODEL_PATH"

LOW_LV1 = 60
LOW_LV2 = 90
HIGH_LV2 = 120
HIGH_LV1 = 200

LEVEL_COLORS = {
    "LOW Lv1": (255, 0, 0),
    "LOW Lv2": (255, 255, 0),
    "HIGH Lv1": (0, 0, 255),
    "HIGH Lv2": (0, 165, 255),
    "NORMAL": (0, 255, 0),
}


class ThermalDetector:
    def __init__(self, model_path: Optional[str] = None) -> None:
        configured_path = model_path or os.getenv(MODEL_PATH_ENV)
        self.model_path = Path(configured_path) if configured_path else DEFAULT_MODEL_PATH
        if not self.model_path.is_absolute():
            self.model_path = Path(__file__).resolve().parent / self.model_path

        if not self.model_path.exists():
            raise FileNotFoundError(f"Model file not found: {self.model_path}")

        self.model = YOLO(str(self.model_path))

    @property
    def is_loaded(self) -> bool:
        return self.model is not None

    def health(self) -> Dict[str, Any]:
        return {
            "model_loaded": self.is_loaded,
            "model_path": str(self.model_path),
            "device": self._device_info(),
        }

    def predict(
        self,
        image: np.ndarray,
        conf: float = 0.10,
        iou: float = 0.60,
        imgsz: int = 1024,
    ) -> Dict[str, Any]:
        started_at = time.perf_counter()
        height, width = image.shape[:2]
        gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        model_input = cv2.cvtColor(gray, cv2.COLOR_GRAY2BGR)

        results = self.model.predict(
            source=model_input,
            conf=conf,
            iou=iou,
            imgsz=imgsz,
            verbose=False,
        )

        detections = self._format_detections(results[0], gray, width, height)
        elapsed_ms = round((time.perf_counter() - started_at) * 1000, 2)

        return {
            "success": True,
            "image": {
                "width": width,
                "height": height,
            },
            "detections": detections,
            "elapsed_ms": elapsed_ms,
        }

    def save_annotated_image(
        self,
        image: np.ndarray,
        detections: List[Dict[str, Any]],
        output_path: Path,
    ) -> None:
        annotated = image.copy()

        for detection in detections:
            xyxy = detection["box"]["xyxy"]
            x1, y1, x2, y2 = [int(round(value)) for value in xyxy]
            x1, y1, x2, y2 = self._clip_xyxy(
                np.array([x1, y1, x2, y2]),
                annotated.shape[1],
                annotated.shape[0],
            )

            temperature = detection["temperature"]
            level = temperature["level"]
            mean_gray = temperature["mean_gray"]
            confidence = detection["confidence"]
            color = LEVEL_COLORS.get(level, (0, 255, 0))
            label = f"{level} conf:{confidence:.2f} gray:{mean_gray:.1f}"

            cv2.rectangle(annotated, (x1, y1), (x2, y2), color, 2)
            self._draw_label(annotated, label, x1, y1, color)

        output_path.parent.mkdir(parents=True, exist_ok=True)
        if not cv2.imwrite(str(output_path), annotated):
            raise RuntimeError(f"Failed to save annotated image: {output_path}")

    def _format_detections(
        self,
        result: Any,
        gray: np.ndarray,
        image_width: int,
        image_height: int,
    ) -> List[Dict[str, Any]]:
        boxes = result.boxes
        if boxes is None or len(boxes) == 0:
            return []

        xyxy_values = boxes.xyxy.cpu().numpy()
        xywh_values = boxes.xywh.cpu().numpy()
        conf_values = boxes.conf.cpu().numpy()

        detections: List[Dict[str, Any]] = []
        for xyxy, xywh, confidence in zip(xyxy_values, xywh_values, conf_values):
            x1, y1, x2, y2 = self._clip_xyxy(xyxy, image_width, image_height)
            roi = gray[y1:y2, x1:x2]
            mean_gray = float(np.mean(roi)) if roi.size else 0.0
            level = classify_gray_level(mean_gray)
            if level == "NORMAL":
                continue

            detections.append(
                {
                    "box": {
                        "xyxy": [round(float(v), 2) for v in xyxy],
                        "xywh": [round(float(v), 2) for v in xywh],
                    },
                    "confidence": round(float(confidence), 6),
                    "temperature": {
                        "mean_gray": round(mean_gray, 2),
                        "level": level,
                    },
                }
            )

        return detections

    def _device_info(self) -> Dict[str, Any]:
        device = "cuda" if torch.cuda.is_available() else "cpu"
        return {
            "type": device,
            "cuda_available": torch.cuda.is_available(),
            "cuda_device_name": torch.cuda.get_device_name(0)
            if torch.cuda.is_available()
            else None,
        }

    @staticmethod
    def _clip_xyxy(
        xyxy: np.ndarray,
        image_width: int,
        image_height: int,
    ) -> List[int]:
        x1, y1, x2, y2 = xyxy[:4]
        return [
            max(0, min(image_width, int(x1))),
            max(0, min(image_height, int(y1))),
            max(0, min(image_width, int(x2))),
            max(0, min(image_height, int(y2))),
        ]

    @staticmethod
    def _draw_label(
        image: np.ndarray,
        label: str,
        x: int,
        y: int,
        color: tuple,
    ) -> None:
        font = cv2.FONT_HERSHEY_SIMPLEX
        font_scale = 0.55
        thickness = 2
        text_size, baseline = cv2.getTextSize(label, font, font_scale, thickness)
        text_width, text_height = text_size

        top = max(0, y - text_height - baseline - 6)
        bottom = top + text_height + baseline + 6
        right = min(image.shape[1], x + text_width + 8)

        cv2.rectangle(image, (x, top), (right, bottom), color, -1)
        cv2.putText(
            image,
            label,
            (x + 4, bottom - baseline - 3),
            font,
            font_scale,
            (255, 255, 255),
            thickness,
            cv2.LINE_AA,
        )


def decode_image(image_bytes: bytes) -> np.ndarray:
    if not image_bytes:
        raise ValueError("Uploaded image is empty")

    image_array = np.frombuffer(image_bytes, dtype=np.uint8)
    image = cv2.imdecode(image_array, cv2.IMREAD_COLOR)
    if image is None:
        raise ValueError("Uploaded file is not a valid image")
    return image


def classify_gray_level(mean_gray: float) -> str:
    if mean_gray <= LOW_LV1:
        return "LOW Lv1"
    if mean_gray <= LOW_LV2:
        return "LOW Lv2"
    if mean_gray >= HIGH_LV1:
        return "HIGH Lv1"
    if mean_gray >= HIGH_LV2:
        return "HIGH Lv2"
    return "NORMAL"
