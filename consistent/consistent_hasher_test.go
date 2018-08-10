package consistent

import (
	"testing"
)

func TestGet(t *testing.T) {
	h := NewConsistentHasher([]Machine{Machine{Url: "http://localhost", Port: "8080"}})

	if _, err := h.GetServer(0); err != nil {
		t.Fatalf("Couldn't find a server when one should exist.")
	}

	h = NewConsistentHasher([]Machine{})

	if machine, err := h.GetServer(0); err == nil {
		t.Fatalf("Found a server: %v, while there should be none.", machine)
	}
}

func TestRemove(t *testing.T) {
	h := NewConsistentHasher([]Machine{Machine{Url: "http://localhost", Port: "8080"}})

	if _, err := h.GetServer(0); err != nil {
		t.Fatalf("Couldn't find a server when one should exist.")
	}

	h.RemoveServer(Machine{Url: "http://localhost", Port: "8080"})

	if machine, err := h.GetServer(0); err == nil {
		t.Fatalf("Found a server: %v, while there should be none.", machine)
	}
}
