from fastapi import FastAPI
from app.schemas import AnalyzeRequest, AnalyzeResponse

app = FastAPI()

@app.get("/health")
def health():
    return {"status": "ok"}

@app.post("/analyze", response_model=AnalyzeResponse)
def analyze(req: AnalyzeRequest):
    # placeholder until model.py is wired up
    return AnalyzeResponse(
        newsId=req.newsId,
        sentiment="NEUTRAL",
        score=0.0,
        model={"name": "placeholder", "version": "v0"},
        createdAt=0,
    )
