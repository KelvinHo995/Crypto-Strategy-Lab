import unittest

from app.model import LexiconSentimentModel


class LexiconSentimentModelTest(unittest.TestCase):
    def setUp(self):
        self.model = LexiconSentimentModel()

    def test_positive_negative_and_neutral(self):
        self.assertEqual(self.model.analyze("ETF inflow sparks bullish rally").label, "POSITIVE")
        self.assertEqual(self.model.analyze("Market crash after exchange hack").label, "NEGATIVE")
        self.assertEqual(self.model.analyze("Bitcoin trades sideways today").label, "NEUTRAL")

    def test_realistic_headlines_and_version(self):
        self.assertEqual(self.model.version, "v2")
        self.assertEqual(self.model.analyze("Crypto funding rebounds as adoption gains").label, "POSITIVE")
        self.assertEqual(self.model.analyze("Hackers exploited a critical wallet flaw").label, "NEGATIVE")


if __name__ == "__main__":
    unittest.main()
