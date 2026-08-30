from pydantic import BaseModel, Field

class AnalyzeRequest(BaseModel):
    newsId: str = Field(min_length=1, max_length=200)
    text: str = Field(min_length=1, max_length=20_000)

class ModelInfo(BaseModel):
    name: str
    version: str

class AnalyzeResponse(BaseModel):
    newsId: str
    sentiment: str
    score: float = Field(ge=0, le=1)
    model: ModelInfo
    createdAt: int
