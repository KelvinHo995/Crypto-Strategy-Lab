from fastapi import FastAPI
from time import time
from app.model import model
from app.schemas import AnalyzeRequest, AnalyzeResponse

app = FastAPI()

@app.get("/health")
def health():
    return {"status": "ok"}

@app.post("/analyze", response_model=AnalyzeResponse)
def analyze(req: AnalyzeRequest):
    result = model.analyze(req.text)
    return AnalyzeResponse(
        newsId=req.newsId,
        sentiment=result.label,
        score=result.score,
        model={"name": model.name, "version": model.version},
        createdAt=int(time()),
    )
