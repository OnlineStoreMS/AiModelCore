package secret

import "testing"

func TestRoundTrip(t *testing.T) {
	enc, err := Encrypt("k", "sk-test-1234")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decrypt("k", enc)
	if err != nil {
		t.Fatal(err)
	}
	if got != "sk-test-1234" {
		t.Fatalf("got %q", got)
	}
	if Hint("sk-test-1234") != "····1234" {
		t.Fatalf("hint %s", Hint("sk-test-1234"))
	}
}
