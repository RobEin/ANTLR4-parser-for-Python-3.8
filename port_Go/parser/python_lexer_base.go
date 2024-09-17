/*
The MIT License (MIT)
Copyright (c) 2021 Robert Einhorn

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

/*
 *
 * Project      : Python Indent/Dedent handler for ANTLR4 grammars
 *
 * Developed by : Robert Einhorn, robert.einhorn.hu@gmail.com
 *
 */

package parser

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
)

type PythonLexerBase struct {
	*antlr.BaseLexer

	// A stack that keeps track of the indentation lengths
	indentLengthStack []int
	// A list where tokens are waiting to be loaded into the token stream
	pendingTokens []antlr.Token

	// last pending token types
	previousPendingTokenType               int
	lastPendingTokenTypeFromDefaultChannel int

	// The amount of opened parentheses, square brackets or curly braces
	opened int

	wasSpaceIndentation           bool
	wasTabIndentation             bool
	wasIndentationMixedWithSpaces bool

	curToken antlr.Token // current (under processing) token
	ffgToken antlr.Token // following (look ahead) token
}

const INVALID_LENGTH int = -1
const ERR_TXT string = " ERROR: "

func NewPythonLexerBase(input antlr.CharStream) *PythonLexerBase {
	plb := new(PythonLexerBase)
	plb.BaseLexer = antlr.NewBaseLexer(input)
	plb.init()
	return plb
}

func (p *PythonLexerBase) NextToken() antlr.Token { // reading the input stream until a return EOF
	p.checkNextToken()
	first := p.pendingTokens[0]
	p.pendingTokens = p.pendingTokens[1:] // remove the first element
	return first                          // add the queued token to the token stream
}

func (p *PythonLexerBase) Reset() {
	p.init()
	p.BaseLexer.Reset()
}

func (p *PythonLexerBase) init() {
	p.indentLengthStack = []int{}     // clear stack
	p.pendingTokens = []antlr.Token{} // clear stack
	p.previousPendingTokenType = 0
	p.lastPendingTokenTypeFromDefaultChannel = 0
	p.opened = 0
	p.wasSpaceIndentation = false
	p.wasTabIndentation = false
	p.wasIndentationMixedWithSpaces = false
	p.curToken = nil
	p.ffgToken = nil
}

func (p *PythonLexerBase) checkNextToken() {
	if p.previousPendingTokenType != antlr.TokenEOF {
		p.setCurrentAndFollowingTokens()
		if len(p.indentLengthStack) == 0 { // We're at the first token
			p.handleStartOfInput()
		}

		switch p.curToken.GetTokenType() {
		case PythonLexerLPAR, PythonLexerLSQB, PythonLexerLBRACE:
			p.opened++
			p.addPendingToken(p.curToken)
		case PythonLexerRPAR, PythonLexerRSQB, PythonLexerRBRACE:
			p.opened--
			p.addPendingToken(p.curToken)
		case PythonLexerNEWLINE:
			p.handleNEWLINEtoken()
		case PythonLexerERRORTOKEN:
			p.reportLexerError(fmt.Sprintf("token recognition error at: '%s'", p.curToken.GetText()))
			p.addPendingToken(p.curToken)
		case antlr.TokenEOF:
			p.handleEOFtoken()
		default:
			p.addPendingToken(p.curToken)
		}
	}
}

func (p *PythonLexerBase) setCurrentAndFollowingTokens() {
	if p.ffgToken == nil {
		p.curToken = p.BaseLexer.NextToken()
	} else {
		p.curToken = p.ffgToken
	}

	if p.curToken.GetTokenType() == antlr.TokenEOF {
		p.ffgToken = p.curToken
	} else {
		p.ffgToken = p.BaseLexer.NextToken()
	}
}

// initialize the indentLengthStack
// hide the leading NEWLINE token(s)
// if exists, find the first statement (not NEWLINE, not EOF token) that comes from the default channel
// insert a leading INDENT token if necessary
func (p *PythonLexerBase) handleStartOfInput() {
	// initialize the stack with a default 0 indentation length
	p.indentLengthStack = append(p.indentLengthStack, 0) /* stack push */ // this will never be popped off
	for p.curToken.GetTokenType() != antlr.TokenEOF {
		if p.curToken.GetChannel() == antlr.TokenDefaultChannel {
			if p.curToken.GetTokenType() == PythonLexerNEWLINE {
				// all the NEWLINE tokens must be ignored before the first statement
				p.hideAndAddPendingToken(p.curToken)
			} else { // We're at the first statement
				p.insertLeadingIndentToken()
				return // continue the processing of the current token with checkNextToken()
			}
		} else {
			p.addPendingToken(p.curToken) // it can be WS, EXPLICIT_LINE_JOINING or COMMENT token
		}
		p.setCurrentAndFollowingTokens()
		// continue the processing of the EOF token with checkNextToken()
	}
}

func (p *PythonLexerBase) insertLeadingIndentToken() {
	if p.previousPendingTokenType == PythonLexerWS {
		prevToken := p.pendingTokens[len(p.pendingTokens)-1]  /* .peekLast() */ // WS token
		if p.getIndentationLength(prevToken.GetText()) != 0 { // there is an "indentation" before the first statement
			errMsg := "first statement indented"
			p.reportLexerError(errMsg)
			// insert an INDENT token before the first statement to raise an 'unexpected indent' error later by the parser
			p.createAndAddPendingToken(PythonLexerINDENT, antlr.TokenDefaultChannel, ERR_TXT+errMsg, p.curToken)
		}
	}
}

func (p *PythonLexerBase) handleNEWLINEtoken() {
	if p.opened > 0 { // We're in an implicit line joining, ignore the current NEWLINE token
		p.hideAndAddPendingToken(p.curToken)
	} else {
		var nlToken antlr.Token = antlr.NewCommonTokenFromToken(p.curToken) // save the current NEWLINE token
		isLookingAhead := p.ffgToken.GetTokenType() == PythonLexerWS
		if isLookingAhead {
			p.setCurrentAndFollowingTokens() // set the next two tokens
		}

		switch p.ffgToken.GetTokenType() {
		case PythonLexerNEWLINE, // We're before a blank line
			PythonLexerCOMMENT: // We're before a comment
			p.hideAndAddPendingToken(nlToken)
			if isLookingAhead {
				p.addPendingToken(p.curToken) // WS token
			}
		default:
			p.addPendingToken(nlToken)
			if isLookingAhead { // We're on whitespace(s) followed by a statement
				var indentationLength int
				if p.ffgToken.GetTokenType() == antlr.TokenEOF {
					indentationLength = 0
				} else {
					indentationLength = p.getIndentationLength(p.curToken.GetText())
				}

				if indentationLength != INVALID_LENGTH {
					p.addPendingToken(p.curToken)                  // WS token
					p.insertIndentOrDedentToken(indentationLength) // may insert INDENT token or DEDENT token(s)
				} else {
					p.reportError("inconsistent use of tabs and spaces in indentation")
				}
			} else { // We're at a newline followed by a statement (there is no whitespace before the statement)
				p.insertIndentOrDedentToken(0) // may insert DEDENT token(s)
			}
		}
	}
}

func (p *PythonLexerBase) insertIndentOrDedentToken(indentLength int) {
	prevIndentLength := p.indentLengthStack[len(p.indentLengthStack)-1] // stack peek
	if indentLength > prevIndentLength {
		p.createAndAddPendingToken(PythonLexerINDENT, antlr.TokenDefaultChannel, "", p.ffgToken)
		p.indentLengthStack = append(p.indentLengthStack, indentLength) // stack push
	} else {
		for indentLength < prevIndentLength { // more than 1 DEDENT token may be inserted to the token stream
			p.indentLengthStack = p.indentLengthStack[:len(p.indentLengthStack)-1] // stack pop
			prevIndentLength = p.indentLengthStack[len(p.indentLengthStack)-1]     // stack peek
			if indentLength <= prevIndentLength {
				p.createAndAddPendingToken(PythonLexerDEDENT, antlr.TokenDefaultChannel, "", p.ffgToken)
			} else {
				p.reportError("inconsistent dedent")
			}
		}
	}
}

func (p *PythonLexerBase) insertTrailingTokens() {
	switch p.lastPendingTokenTypeFromDefaultChannel {
	case PythonLexerNEWLINE,
		PythonLexerDEDENT:

		// no trailing NEWLINE token is needed
	default:
		// insert an extra trailing NEWLINE token that serves as the end of the last statement
		p.createAndAddPendingToken(PythonLexerNEWLINE, antlr.TokenDefaultChannel, "", p.ffgToken) // ffgToken is EOF
	}
	p.insertIndentOrDedentToken(0) // Now insert as much trailing DEDENT tokens as needed
}

func (p *PythonLexerBase) handleEOFtoken() {
	if p.lastPendingTokenTypeFromDefaultChannel > 0 {
		// there was statement in the input (leading NEWLINE tokens are hidden)
		p.insertTrailingTokens()
	}
	p.addPendingToken(p.curToken)
}

func (p *PythonLexerBase) hideAndAddPendingToken(tkn antlr.Token) {
	ctkn := antlr.NewCommonTokenFromToken(tkn)
	ctkn.SetChannel(antlr.TokenHiddenChannel)
	p.addPendingToken(ctkn)
}

func (p *PythonLexerBase) createAndAddPendingToken(ttype int, channel int, text string, sampleToken antlr.Token) {
	ctkn := antlr.NewCommonTokenFromToken(sampleToken)
	ctkn.SetTokenType(ttype)
	ctkn.SetChannel(channel)
	ctkn.SetStop(sampleToken.GetStart() - 1)
	if text == "" {
		ctkn.SetText(fmt.Sprintf("<%s>", p.GetSymbolicNames()[ttype]))
	} else {
		ctkn.SetText(text)
	}
	p.addPendingToken(ctkn)
}

func (p *PythonLexerBase) addPendingToken(token antlr.Token) {
	// save the last pending token type because the pendingTokens linked list can be empty by the nextToken()
	p.previousPendingTokenType = token.GetTokenType()
	if token.GetChannel() == antlr.TokenDefaultChannel {
		p.lastPendingTokenTypeFromDefaultChannel = p.previousPendingTokenType
	}
	p.pendingTokens = append(p.pendingTokens, token) /* .addLast(token) */
}

func (p *PythonLexerBase) getIndentationLength(indentText string) int { // the indentText may contain spaces, tabs or form feeds
	TAB_LENGTH := 8 // the standard number of spaces to replace a tab to spaces
	length := 0
	for _, ch := range indentText {
		switch ch {
		case ' ':
			p.wasSpaceIndentation = true
			length += 1
		case '\t':
			p.wasTabIndentation = true
			length += TAB_LENGTH - (length % TAB_LENGTH)
		case '\f': // form feed
			length = 0
		}
	}
	if p.wasTabIndentation && p.wasSpaceIndentation {
		if !p.wasIndentationMixedWithSpaces {
			p.wasIndentationMixedWithSpaces = true
			length = INVALID_LENGTH // only for the first inconsistent indent
		}
	}
	return length
}

func (p *PythonLexerBase) reportLexerError(errMsg string) {
	p.GetErrorListenerDispatch().SyntaxError(p, p.curToken, p.curToken.GetLine(), p.curToken.GetColumn(), " LEXER"+ERR_TXT+errMsg, nil)
}

func (p *PythonLexerBase) reportError(errMsg string) {
	p.reportLexerError(errMsg)

	// the ERRORTOKEN will raise an error in the parser
	p.createAndAddPendingToken(PythonLexerERRORTOKEN, antlr.TokenDefaultChannel, ERR_TXT+errMsg, p.ffgToken)
}
