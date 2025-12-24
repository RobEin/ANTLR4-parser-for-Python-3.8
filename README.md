# ANTLR4 parser for Python 3.8.20 &nbsp; [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

### About files:
 - PythonParser.g4 is the ANTLR4 parser grammar that based on the last "traditional" [Python grammar](https://docs.python.org/3.8/reference/grammar.html) which not yet written in [PEG](https://peps.python.org/pep-0617/)<br/><br/>

 - PythonLexerBase:
   - handles the Python indentations
   - creates encoding token
   - and manage many other things<br/><br/>

 ### Recent changes:
- tokenizing BOM Unicode character at the start of the file so it is skipped in the token stream
- moved encoding detection from PythonLexerBase to a separate component (grun4py)

#### [Previous changes](https://github.com/RobEin/ANTLR4-parser-for-Python-3.8/blob/main/changes.md)<br/><br/>

### Related links:
[ANTLR 4](https://www.antlr.org/)

[ANTLR4-tools](https://github.com/antlr/antlr4/blob/master/doc/getting-started.md#getting-started-the-easy-way-using-antlr4-tools)

[ANTLR 4 Documentation](https://github.com/antlr/antlr4/tree/master/doc)

[ANTLR 4 Runtime API](https://www.antlr.org/api/Java/)

[Python 3.8 Lexical Analysis](https://docs.python.org/3.8/reference/lexical_analysis.html)

[ANTLR4 parser for Python 2.7.18](https://github.com/RobEin/ANTLR4-parser-for-Python-2.7.18)

[ANTLR4 Python parser based on the official PEG grammar](https://github.com/RobEin/ANTLR4-Python-parser-by-PEG)
