from pydantic import BaseModel

class AnalyzeRequest(BaseModel):
    newsId: str
    text: str

class ModelInfo(BaseModel):
    name: str
    version: str

class AnalyzeResponse(BaseModel):
    newsId: str
    sentiment: str
    score: float
    model: ModelInfo
    createdAt: int
