import unittest

from app.model import LexiconSentimentModel


class LexiconSentimentModelTest(unittest.TestCase):
    def setUp(self):
        self.model = LexiconSentimentModel()

    def test_positive_negative_and_neutral(self):
        self.assertEqual(self.model.analyze("ETF inflow sparks bullish rally").label, "POSITIVE")
        self.assertEqual(self.model.analyze("Market crash after exchange hack").label, "NEGATIVE")
        self.assertEqual(self.model.analyze("Bitcoin trades sideways today").label, "NEUTRAL")


if __name__ == "__main__":
    unittest.main()
