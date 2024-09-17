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

import 'package:antlr4/antlr4.dart';
import 'PythonLexer.dart';

abstract class PythonLexerBase extends Lexer {
  // A stack that keeps track of the indentation lengths
  late List<int> indentLengthStack;
  // A list where tokens are waiting to be loaded into the token stream
  late List<Token> pendingTokens;

  // last pending token types
  late int previousPendingTokenType;
  late int lastPendingTokenTypeFromDefaultChannel;

  // The amount of opened parentheses, square brackets or curly braces
  late int opened;
  
  late bool wasSpaceIndentation;
  late bool wasTabIndentation;
  late bool wasIndentationMixedWithSpacesAndTabs;

  late Token? curToken; // current (under processing) token
  late Token? ffgToken; // following (look ahead) token

  static const int INVALID_LENGTH = -1;
  static const String ERR_TXT = " ERROR: ";

  PythonLexerBase(CharStream input) : super(input) {
    init();
  }

  @override
  Token nextToken() {
    // reading the input stream until a return EOF
    checkNextToken();
    // add the queued token to the token stream
    return pendingTokens.removeAt(0);
  }

  @override
  void reset([bool resetInput = false]) {
    init();
    super.reset(resetInput);
  }

  void init() {
    indentLengthStack = [];
    pendingTokens = [];
    previousPendingTokenType = 0;
    lastPendingTokenTypeFromDefaultChannel = 0;
    opened = 0;
    wasSpaceIndentation = false;
    wasTabIndentation = false;
    wasIndentationMixedWithSpacesAndTabs = false;
    curToken = null;
    ffgToken = null;
  }

  void checkNextToken() {
    if (previousPendingTokenType != Token.EOF) {
      setCurrentAndFollowingTokens();
      if (indentLengthStack.isEmpty) {
        // We're at the first token
        handleStartOfInput();
      }

      switch (curToken!.type) {
        case PythonLexer.TOKEN_LPAR:
        case PythonLexer.TOKEN_LSQB:
        case PythonLexer.TOKEN_LBRACE:
          opened++;
          addPendingToken(curToken!);
          break;
        case PythonLexer.TOKEN_RPAR:
        case PythonLexer.TOKEN_RSQB:
        case PythonLexer.TOKEN_RBRACE:
          opened--;
          addPendingToken(curToken!);
          break;
        case PythonLexer.TOKEN_NEWLINE:
          handleNEWLINEtoken();
          break;
        case PythonLexer.TOKEN_ERRORTOKEN:
          reportLexerError("token recognition error at: '${curToken!.text}'");
          addPendingToken(curToken!);
          break;
        case Token.EOF:
          handleEOFtoken();
          break;
        default:
          addPendingToken(curToken!);
      }
    }
  }

  void setCurrentAndFollowingTokens() {
    curToken = ffgToken == null ? super.nextToken() : ffgToken;
    ffgToken = curToken!.type == Token.EOF ? curToken : super.nextToken();
  }

  // initialize the indentLengthStack
  // hide the leading NEWLINE token(s)
  // if exists, find the first statement (not NEWLINE, not EOF token) that comes from the default channel
  // insert a leading INDENT token if necessary
  void handleStartOfInput() {
    // initialize the stack with a default 0 indentation length
    indentLengthStack.add(0) /* stack push */; // this will never be popped off
    while (curToken!.type != Token.EOF) {
      if (curToken!.channel == Token.DEFAULT_CHANNEL) {
        if (curToken!.type == PythonLexer.TOKEN_NEWLINE) {
          // all the NEWLINE tokens must be ignored before the first statement
          hideAndAddPendingToken(curToken!);
        } else {
          // We're at the first statement
          insertLeadingIndentToken();
          return; // continue the processing of the current token with checkNextToken()
        }
      } else {
        // it can be WS, EXPLICIT_LINE_JOINING or COMMENT token
        addPendingToken(curToken!);
      }
      setCurrentAndFollowingTokens();
    }
    // continue the processing of the EOF token with checkNextToken()
  }

  void insertLeadingIndentToken() {
    if (previousPendingTokenType == PythonLexer.TOKEN_WS) {
      Token prevToken = pendingTokens.last; // WS token
      if (getIndentationLength(prevToken.text!) != 0) {
        // there is an "indentation" before the first statement
        final String errMsg = "first statement indented";
        reportLexerError(errMsg);
        // insert an INDENT token before the first statement to raise an 'unexpected indent' error later by the parser
        createAndAddPendingToken(PythonLexer.TOKEN_INDENT,
            Token.DEFAULT_CHANNEL, ERR_TXT + errMsg, curToken!);
      }
    }
  }

  void handleNEWLINEtoken() {
    if (opened > 0) {
      // We're in an implicit line joining, ignore the current NEWLINE token
      hideAndAddPendingToken(curToken!);
    } else {
      final Token nlToken =
          CommonToken.copy(curToken!); // save the current NEWLINE token
      final bool isLookingAhead = ffgToken!.type == PythonLexer.TOKEN_WS;
      if (isLookingAhead) {
        setCurrentAndFollowingTokens(); // set the next two tokens
      }

      switch (ffgToken!.type) {
        case PythonLexer.TOKEN_NEWLINE: // We're before a blank line
        case PythonLexer.TOKEN_COMMENT: // We're before a comment
          hideAndAddPendingToken(nlToken);
          if (isLookingAhead) {
            addPendingToken(curToken!); // WS token
          }
          break;
        default:
          addPendingToken(nlToken);
          if (isLookingAhead) {
            // We're on whitespace(s) followed by a statement
            final int indentationLength = ffgToken!.type == Token.EOF
                ? 0
                : getIndentationLength(curToken!.text!);

            if (indentationLength != INVALID_LENGTH) {
              addPendingToken(curToken!); // WS token
              
              // may insert INDENT token or DEDENT token(s)
              insertIndentOrDedentToken(indentationLength);
            } else {
              reportError("inconsistent use of tabs and spaces in indentation");
            }
          } else {
            // We're at a newline followed by a statement (there is no whitespace before the statement)
            insertIndentOrDedentToken(0); // may insert DEDENT token(s)
          }
      }
    }
  }

  void insertIndentOrDedentToken(final int indentLength) {
    int prevIndentLength = indentLengthStack.last /* stack peek */;
    if (indentLength > prevIndentLength) {
      createAndAddPendingToken(
          PythonLexer.TOKEN_INDENT, Token.DEFAULT_CHANNEL, null, ffgToken!);
      indentLengthStack.add(indentLength) /* stack push */;
    } else {
      while (indentLength < prevIndentLength) {
        // more than 1 DEDENT token may be inserted to the token stream
        indentLengthStack.removeLast() /* stack pop */;
        prevIndentLength = indentLengthStack.last /* stack peek */;
        if (indentLength <= prevIndentLength) {
          createAndAddPendingToken(
              PythonLexer.TOKEN_DEDENT, Token.DEFAULT_CHANNEL, null, ffgToken!);
        } else {
          reportError("inconsistent dedent");
        }
      }
    }
  }

  void insertTrailingTokens() {
    switch (lastPendingTokenTypeFromDefaultChannel) {
      case PythonLexer.TOKEN_NEWLINE:
      case PythonLexer.TOKEN_DEDENT:
        break; // no trailing NEWLINE token is needed
      default:
        // insert an extra trailing NEWLINE token that serves as the end of the last statement
        createAndAddPendingToken(PythonLexer.TOKEN_NEWLINE,
            Token.DEFAULT_CHANNEL, null, ffgToken!); // ffgToken is EOF
    }
    // Now insert as much trailing DEDENT tokens as needed
    insertIndentOrDedentToken(0);
  }

  void handleEOFtoken() {
    if (lastPendingTokenTypeFromDefaultChannel > 0) {
      // there was statement in the input (leading NEWLINE tokens are hidden)
      insertTrailingTokens();
    }
    addPendingToken(curToken!);
  }

  void hideAndAddPendingToken(final Token tkn) {
    CommonToken ctkn = CommonToken.copy(tkn);
    ctkn.channel = Token.HIDDEN_CHANNEL;
    addPendingToken(ctkn);
  }

  void createAndAddPendingToken(final int ttype, final int channel,
      final String? text, final Token sampleToken) {
    CommonToken ctkn = CommonToken.copy(sampleToken);
    ctkn.type = ttype;
    ctkn.channel = channel;
    ctkn.stopIndex = sampleToken.startIndex - 1;
    ctkn.text = text == null
        ? "<" + vocabulary.getSymbolicName(ttype).toString() + ">"
        : text;

    addPendingToken(ctkn);
  }

  void addPendingToken(final Token tkn) {
    // save the last pending token type because the pendingTokens linked list can be empty by the nextToken()
    previousPendingTokenType = tkn.type;
    if (tkn.channel == Token.DEFAULT_CHANNEL) {
      lastPendingTokenTypeFromDefaultChannel = previousPendingTokenType;
    }
    pendingTokens.add(tkn);
  }

  int getIndentationLength(final String indentText) {
    // the indentText may contain spaces, tabs or form feeds

    // the standard number of spaces to replace a tab to spaces
    const int TAB_LENGTH = 8;
    int length = 0;
    for (var ch in indentText.split('')) {
      switch (ch) {
        case ' ':
          wasSpaceIndentation = true;
          length += 1;
          break;
        case '\t':
          wasTabIndentation = true;
          length += TAB_LENGTH - (length % TAB_LENGTH);
          break;
        case '\f': // form feed
          length = 0;
          break;
      }
    }

    if (wasTabIndentation && wasSpaceIndentation) {
      if (!(wasIndentationMixedWithSpacesAndTabs)) {
        wasIndentationMixedWithSpacesAndTabs = true;
        length = INVALID_LENGTH; // only for the first inconsistent indent
      }
    }
    return length;
  }

  void reportLexerError(String errMsg) {
    errorListenerDispatch.syntaxError(
      this,
      curToken,
      curToken!.line,
      curToken!.charPositionInLine,
      " LEXER" + ERR_TXT + errMsg,
      null,
    );
  }

  void reportError(String errMsg) {
    reportLexerError(errMsg);

    // the ERRORTOKEN will raise an error in the parser
    createAndAddPendingToken(PythonLexer.TOKEN_ERRORTOKEN,
        Token.DEFAULT_CHANNEL, ERR_TXT + errMsg, ffgToken!);
  }
}
