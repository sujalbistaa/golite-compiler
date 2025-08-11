# GoLite MVP Compiler (Stage 1)

This is the first stage of the GoLite compiler, implementing a basic lexer and a recursive-descent parser with Pratt parsing for expressions.

## Language Features (Stage 1)

- `let` bindings: `let x = 5;`
- `print` statements: `print x + 10;`
- Integer literals
- Basic arithmetic operators: `+`, `-`, `*`, `/`
- Parenthesized expressions for grouping
- Simple function definitions and calls (without return values yet)

## How to Build and Run

You must have Go (version 1.21 or newer) installed.

### Build

To build the `golite` command-line tool, run the following from the root directory of the project:

```sh
go build ./...