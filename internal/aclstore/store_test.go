package aclstore

import (
	"testing"
)

func TestAddListDelete(t *testing.T) {
	tmp := t.TempDir()
	s := New(tmp)

	if err := s.Add("porn", TypeURLs, "sexy.com"); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if err := s.Add("porn", TypeDomains, "example.org"); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	rows, err := s.List("porn", "")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(rows["porn/urls"]) != 1 || rows["porn/urls"][0] != "sexy.com" {
		t.Fatalf("unexpected url rows: %#v", rows["porn/urls"])
	}

	if err := s.Delete("porn", TypeURLs, "sexy.com"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if err := s.Delete("porn", TypeURLs, "sexy.com"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	rows, err = s.List("porn", TypeURLs)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if _, ok := rows["porn/urls"]; ok {
		t.Fatalf("expected urls to be empty after delete, got %#v", rows["porn/urls"])
	}
}

func TestQueryEngineMatch(t *testing.T) {
	tmp := t.TempDir()
	s := New(tmp)
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(s.Add("mylist", TypeDomains, "example.com"))
	must(s.Add("mylist", TypeURLs, "/blocked/"))
	must(s.Add("mylist", TypeExpressions, `https?://bad\.example/.*`))

	q := NewQueryEngine(s)
	if _, ok, err := q.Match("mylist", "sub.example.com", "http://good.example/"); err != nil || !ok {
		t.Fatalf("expected domain match, ok=%v err=%v", ok, err)
	}
	if kind, ok, err := q.Match("mylist", "other.net", "http://site/blocked/path"); err != nil || !ok || kind != "url" {
		t.Fatalf("expected url match, kind=%s ok=%v err=%v", kind, ok, err)
	}
	if kind, ok, err := q.Match("mylist", "other.net", "https://bad.example/path"); err != nil || !ok || kind != "expression" {
		t.Fatalf("expected expression match, kind=%s ok=%v err=%v", kind, ok, err)
	}
	if _, ok, err := q.Match("mylist", "other.net", "https://safe.example/"); err != nil || ok {
		t.Fatalf("expected no match, ok=%v err=%v", ok, err)
	}
}
