package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/lexer"
	// "monkey/token"
	"monkey/parser"
)

// PROMPT is the prompt of the shell
const PROMPT = "λ>> "

// Start start a repl
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		prog := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParseError(out, p.Errors())
			continue
		}

		// token repl
		// for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		// 	fmt.Printf("%+v\n", tok)
		// }
		io.WriteString(out, prog.String())
		io.WriteString(out, "\n")
	}
}

func printParseError(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t" + msg + "\n")
	}
}
