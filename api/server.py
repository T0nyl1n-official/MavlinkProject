from contextlib import asynccontextmanager
from pathlib import Path
from typing import Any, Dict, Optional
from uuid import uuid4

from fastapi import FastAPI, File, HTTPException, Query, UploadFile
from fastapi.staticfiles import StaticFiles

from thermal_detector import ThermalDetector, decode_image


TUNNEL_BASE_URL = "http://154265c9.r6.cpolar.cn"
PROJECT_ROOT = Path(__file__).resolve().parent.parent
ANNOTATED_DIR = PROJECT_ROOT / "outputs/annotated"
ANNOTATED_ROUTE = "/static/annotated"

detector: Optional[ThermalDetector] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global detector
    detector = ThermalDetector()
    try:
        yield
    finally:
        detector = None


app = FastAPI(
    title="YOLOv8s Thermal Detector API",
    version="1.0.0",
    lifespan=lifespan,
)

ANNOTATED_DIR.mkdir(parents=True, exist_ok=True)
app.mount(
    ANNOTATED_ROUTE,
    StaticFiles(directory=str(ANNOTATED_DIR)),
    name="annotated",
)


@app.get("/health")
def health() -> Dict[str, Any]:
    if detector is None:
        return {
            "model_loaded": False,
            "model_path": None,
            "device": None,
        }
    return detector.health()


@app.post("/api/v1/detect")
async def detect(
    image: UploadFile = File(...),
    conf: float = Query(0.10, ge=0.0, le=1.0),
    iou: float = Query(0.60, ge=0.0, le=1.0),
    imgsz: int = Query(1024, ge=32, le=4096),
) -> Dict[str, Any]:
    if detector is None:
        raise HTTPException(status_code=503, detail="Model is not loaded")

    try:
        image_bytes = await image.read()
        frame = decode_image(image_bytes)
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc

    result = detector.predict(frame, conf=conf, iou=iou, imgsz=imgsz)

    filename = f"{uuid4().hex}.jpg"
    output_path = ANNOTATED_DIR / filename
    try:
        detector.save_annotated_image(frame, result["detections"], output_path)
    except RuntimeError as exc:
        raise HTTPException(status_code=500, detail=str(exc)) from exc

    result["image"]["url"] = build_annotated_url(filename)
    return result


def build_annotated_url(filename: str) -> str:
    base_url = TUNNEL_BASE_URL.rstrip("/")
    return f"{base_url}{ANNOTATED_ROUTE}/{filename}"


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("api.server:app", host="0.0.0.0", port=8100, reload=False)
