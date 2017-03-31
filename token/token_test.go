package token

import (
	"testing"
)

func TestToken(t *testing.T) {
	ident := []string{"x", "fn", "return", "false", "if", "true"}
	expectedType :=[]TokenType{IDENT, FUNCTION, RETURN, FALSE, IF, TRUE}
	for i, curIdent := range ident {
		curExpectType := expectedType[i]
		typep := LookupIdent(curIdent)
		if typep != curExpectType {
			t.Fatalf("test fail, expect: %v, got %v\n", curExpectType, typep)
		}
	}
}
