# ANTLR4 parser for Python 3.8.12 to 3.8.17 &nbsp; [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

### About files:
 - PythonParser.g4
   is the ANTLR4 parser grammar that based on the last "traditional" [Python grammar](https://docs.python.org/3.8/reference/grammar.html) which not yet written in [PEG](https://en.wikipedia.org/wiki/Parsing_expression_grammar).
The official grammar of the Python 3.8.12 to 3.8.17 are the same.

 - PythonLexerBase.java
   handles the Python indentations.
   This class also can be used with older ANTLR4 Python grammars.
   See the instructions in the comments.
   
 - Test files: [Python 3.8 Standard Lib](http://www.jorkka.net:8082/ruuvi/Python-3.8.0/Lib/)


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

[Python 3.8 Lexical Analysis](https://docs.python.org/3.8/reference/lexical_analysis.html)
