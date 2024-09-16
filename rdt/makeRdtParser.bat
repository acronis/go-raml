@echo off
WHERE /q antlr4
IF ERRORLEVEL 1 (
    ECHO "ANTLR4 is missing. If you have Python installed, you can install it by running 'pip install anltr4-tools'."
    PAUSE
) ELSE (
    antlr4 -Dlanguage=Go .\rdtLexer.g4
    antlr4 -Dlanguage=Go -visitor -no-listener .\rdtParser.g4
)