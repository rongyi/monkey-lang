package token

import (
	"testing"
)

func TestToken(t *testing.T) {
	ident := "x"
	identType := LookupIdent(ident)
	if identType != IDENT {
		t.Fatalf("get ident fail\n")
	}
	ident = "fn"
	identType = LookupIdent(ident)
	if identType != FUNCTION {
		t.Fatalf("get function fail\n")
	}
}
