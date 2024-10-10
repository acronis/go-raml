if ! command -v antlr4 &> /dev/null
then
    echo "ANTLR4 is missing. If you have Python installed, you can install it by running 'pip install antlr4-tools'."
    exit 1
fi

antlr4 -Dlanguage=Go ./rdtLexer.g4
antlr4 -Dlanguage=Go -visitor -no-listener ./rdtParser.g4