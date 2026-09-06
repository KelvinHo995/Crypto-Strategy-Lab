from dataclasses import dataclass
import re


@dataclass(frozen=True)
class SentimentResult:
    label: str
    score: float


class LexiconSentimentModel:
    """Deterministic MVP model with a stable, versioned contract.

    This avoids returning a fake placeholder while keeping the service light
    enough for local demos. It can later be replaced by FinBERT behind the same
    analyze method.
    """

    name = "crypto-lexicon"
    version = "v2"
    positive = frozenset({
        "adoption", "approve", "approved", "approval", "breakout", "bullish",
        "gain", "gains", "growth", "high", "inflow", "launch", "partnership",
        "rally", "rebound", "rebounds", "record", "recover", "recovered", "rise",
        "surge", "upgrade",
    })
    negative = frozenset({
        "attack", "ban", "bearish", "breach", "crackdown", "crash", "decline",
        "drop", "exploit", "exploited", "fall", "flaw", "fraud", "hack", "hacked",
        "hackers", "lawsuit", "loss", "losses", "outflow", "risk", "scam",
    })

    def analyze(self, text: str) -> SentimentResult:
        words = set(re.findall(r"[a-z0-9]+", text.lower()))
        positive_hits = len(words & self.positive)
        negative_hits = len(words & self.negative)
        total = positive_hits + negative_hits
        if total == 0 or positive_hits == negative_hits:
            return SentimentResult("NEUTRAL", 0.5)
        confidence = max(positive_hits, negative_hits) / total
        return SentimentResult("POSITIVE" if positive_hits > negative_hits else "NEGATIVE", confidence)


model = LexiconSentimentModel()
