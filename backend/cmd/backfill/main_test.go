package main

import (
	"reflect"
	"testing"
)

func TestBackfillSymbols(t *testing.T) {
	got, err := backfillSymbols(" btcusdt, ETHUSDT,btcusdt ", "SOLUSDT")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"BTCUSDT", "ETHUSDT"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("symbols = %v, want %v", got, want)
	}
	fallback, err := backfillSymbols("", "solusdt")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(fallback, []string{"SOLUSDT"}) {
		t.Fatalf("legacy symbol = %v", fallback)
	}
	if _, err := backfillSymbols("BTCUSDT,not-real", ""); err == nil {
		t.Fatal("invalid symbol should fail instead of silently backfilling another market")
	}
}
