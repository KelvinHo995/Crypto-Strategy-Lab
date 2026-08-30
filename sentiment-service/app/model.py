from dataclasses import dataclass


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
    version = "v1"
    positive = frozenset({"gain", "gains", "bullish", "surge", "rally", "rise", "approval", "inflow", "record", "upgrade"})
    negative = frozenset({"loss", "losses", "bearish", "drop", "crash", "fall", "hack", "ban", "outflow", "fraud"})

    def analyze(self, text: str) -> SentimentResult:
        words = {word.strip(".,:;!?()[]{}\"'").lower() for word in text.split()}
        positive_hits = len(words & self.positive)
        negative_hits = len(words & self.negative)
        total = positive_hits + negative_hits
        if total == 0 or positive_hits == negative_hits:
            return SentimentResult("NEUTRAL", 0.5)
        confidence = max(positive_hits, negative_hits) / total
        return SentimentResult("POSITIVE" if positive_hits > negative_hits else "NEGATIVE", confidence)


model = LexiconSentimentModel()
