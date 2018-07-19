package manager

import (
	"testing"
)

func TestGet(t *testing.T) {
	m := HashManager{
		alias: []string{"A", "B", "", "", "", "", "", "", "", "", "C"},
		urls: map[string]bool{
			"A": true,
			"B": true,
			"C": true,
		},
	}

	var url string
	var err error
	url, err = m.Get("1")
	if err != nil {
		t.Fatalf("1 not found when it should \n error: %v", err)
	}

	url, err = m.Get("a")
	if err != nil {
		t.Fatalf("11 not found when it should \n error: %v", err)
	}

	url, err = m.Get("*")
	if err == nil {
		t.Fatalf("* found while it is not a valid hash\nResponse: %v", url)
	}
}

func TestDelete(t *testing.T) {
	m := HashManager{
		alias: []string{"A"},
		urls: map[string]bool{
			"A": true,
		},
	}

	var err error
	err = m.Delete("2")
	if err == nil {
		t.Fatalf("2 entry deleted while it was not known")
	}

	err = m.Delete("0")
	if err != nil {
		t.Fatalf("0 entry wasn't deleted while it was present")
	}

	var url string
	url, err = m.Get("0")
	if err == nil {
		t.Fatalf("1 entry was found (mapped to \"%s\") after being deleted.", url)
	}

	err = m.Delete("0")
	if err == nil {
		t.Fatalf("1 entry was deleted twice, was it deleted the first time ?")
	}
}

func TestAdd(t *testing.T) {
	m := HashManager{
		alias: []string{"A"},
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
		t.Fatalf("\"B\" was not added even though it was not already known.")
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

func TestAddMany(t *testing.T) {
	m := HashManager{
		alias: []string{},
		urls:  map[string]bool{},
	}
	s := []byte{}
	hash, err := m.Add("")
	if err == nil {
		t.Fatalf("Empty string was added, it should be forbiden. The returned hash was: %v", hash)
	}
	for i := 0; i < 100; i++ {
		s = append(s, 'A')
		hash, err = m.Add(string(s))
		if err != nil {
			t.Fatalf("%s could be added but it wasn't added before.", string(s))
		}
	}

	if hash != "1B" {
		t.Fatalf("The 100th created hash is %s, it should be \"1B\"", hash)
	}

	var url string
	url, err = m.Get(hash)
	if err != nil {
		t.Fatalf("Can't get hash %s even though it was returned by Add", hash)
	}
	if url != string(s) {
		t.Fatalf("Returned url is %s, but should be %s", url, string(s))
	}
}
