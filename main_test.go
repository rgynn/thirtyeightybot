package main

import (
	"io/ioutil"
	"testing"
)

func TestFoundButButton(t *testing.T) {

	body, err := ioutil.ReadFile("test/test_body")
	if err != nil {
		t.Fatal(err)
	}

	found, err := foundBuyButton(body)
	if err != nil {
		t.Fatal(err)
	}

	if !found {
		t.Fatalf("expected buy button, got: false")
	}
}
