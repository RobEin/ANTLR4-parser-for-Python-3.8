# Dec. 24, 2025
- tokenizing BOM Unicode character at the start of the file so it is skipped in the token stream
- moved encoding detection from PythonLexerBase to a separate component (grun4py)

# Jan. 07, 2025
- added ENCODING token

# Sept. 05, 2024
- Type comment tokens are no longer generated.  
  Type comments will now be tokenized as plain comment tokens.<br/><br/>

- Line continuation for string literals (backslash followed by a newline) is no longer resolved.  
  (backslash+newline is no longer removed from string literals)
