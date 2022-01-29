# ANTLR4 parser for Python 3.8.12 &nbsp; [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

### About files:
 - PythonParser.g4
   is the ANTLR4 parser grammar that based on the official [Python grammar](https://docs.python.org/3.8/reference/grammar.html) that was the last traditional Python grammar not yet written in [PEG](https://en.wikipedia.org/wiki/Parsing_expression_grammar).

 - PythonLexerBase.java
   handles the Python indentations.
   This class also can be used with older ANTLR4 Python grammars.
   See the instructions in the comments.


### A simple usage example in command line:
```bash
antlr4 PythonLexer.g4
antlr4 PythonParser.g4
javac *.java
grun Python file_input -tokens test.py
```


### Related links:
[ANTLR 4](https://www.antlr.org/)

[ANTLR 4 Documentation](https://github.com/antlr/antlr4/tree/master/doc)

[ANTLR 4 Runtime API](https://www.antlr.org/api/Java/)

[Python 3 Lexical Analysis](https://docs.python.org/3/reference/lexical_analysis.html)

