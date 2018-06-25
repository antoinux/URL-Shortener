package main

import (
	"testing"
)

func TestGet(t *testing.T) {
	m := HashManager{
		alias: map[uint64]string{
			1:  "A",
			10: "B",
		},
		urls: map[string]bool{
			"A": true,
			"B": true,
		},
	}

	var url string
	var err error
	url, err = m.Get("1")
	if err != nil {
		t.Fatalf("1 not found when it should \n error: %v", err)
	}

	url, err = m.Get("A")
	if err != nil {
		t.Fatalf("11 not found when it should \n error: %v", err)
	}

	url, err = m.Get("H")
	if err == nil {
		t.Fatalf("H found while it is not a valid hexstring\nResponse: %v", url)
	}
}

func TestAddEntry(t *testing.T) {
	m := HashManager{
		alias: map[uint64]string{
			1: "A",
		},
		urls: map[string]bool{
			"A": true,
		},
	}

	var err error
	err = m.AddEntry("1", "B")
	if err == nil {
		t.Fatalf("(1, \"B\") entry added while 1 hash already exists")
	}

	err = m.AddEntry("2", "A")
	if err == nil {
		t.Fatalf("(2, \"A\") entry added while \"A\" URL already exists")
	}

	err = m.AddEntry("2", "B")
	if err != nil {
		t.Fatalf("Couldn't add entry (2, \"B\") where it should fit")
	}

	var url string
	url, err = m.Get("2")
	if err != nil {
		t.Fatalf("Couldn't get inserted url (should be \"B\").")
	}
	if url != "B" {
		t.Fatalf("Returned URL is \"%s\" where it should be \"B\"", url)
	}

	err = m.AddEntry("2", "B")
	if err == nil {
		t.Fatalf("Added (2, \"B\") twice, was it actually inserted ?")
	}
}

func TestDelete(t *testing.T) {
	m := HashManager{
		alias: map[uint64]string{
			1: "A",
		},
		urls: map[string]bool{
			"A": true,
		},
	}

	var err error
	err = m.Delete("2")
	if err == nil {
		t.Fatalf("2 entry deleted while it was not known")
	}

	err = m.Delete("1")
	if err != nil {
		t.Fatalf("1 entry wasn't deleted while it was present")
	}

	var url string
	url, err = m.Get("1")
	if err == nil {
		t.Fatalf("1 entry was found (mapped to \"%s\") after being deleted.", url)
	}

	err = m.Delete("1")
	if err == nil {
		t.Fatalf("1 entry was deleted twice, was it deleted the first time ?")
	}
}

func TestAdd(t *testing.T) {
	m := HashManager{
		alias: map[uint64]string{
			0: "A",
		},
		urls: map[string]bool{
			"A": true,
		},
	}
	var err error
	var hash string
	hash, err = m.Add("A")
	if err == nil {
		t.Fatalf("\"A\" URL was added but it was already known.")
	}

	hash, err = m.Add("B")
	if err != nil {
		t.Fatalf("\"B\" was not added while it was now already known.")
	}

	var url string
	url, err = m.Get(hash)
	if err != nil {
		t.Fatalf("hash %s was not found after being created.", hash)
	}
	if url != "B" {
		t.Fatalf("hash %s is mapped to \"%s\" while it should be mapped to \"B\"", hash, url)
	}

	hash, err = m.Add("B")
	if err == nil {
		t.Fatalf("\"B\" was added twice. Was it actually added the first time ?")
	}
}
