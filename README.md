# Brainfuck Compiler in Golang

Brainfuck compiler made in Golang. Brainfuck is a minimalistic esoteric programming language created in 1993 by Urban Müller.

### Intro

I decided to start working on this due to my compilers and interpreters classes being way too underwhelming (I also have always wanted to try this out). This isn't meant to be small nor fast, but just so that I could learn some more about how compilers and interpreters work.

The idea of this project, as the name suggests, is to create a Brainfuck compiler and then expand it upon its functionalities. My final target on this is to have something that resembles [Brain](https://github.com/brain-labs/brain) which is yet another Brainfuck compiler.

I tried my best to keep the code as straight as possible in case anyone ever ends up taking a look at this repository for resources.

### Execution Flow
  1. Reads through the entered file, lexing it and producing a sequence of tokens.
  2. Parsers the just lexed file tokens, constructing an AST.
  3. Reads off the AST nodes, generating corresponding instructions for them.
  4. Passes the just generated .asm output file into ``nasm``, and then link it using GNU linker ``ld``, generating an exectuable.
  5. Ta-da! You now have a functional executable Brainfuck file.

### Prerequisites
 - Have Go installed in your machine [Go website](https://go.dev/).
 - Have NASM installed in your machine [NASM website](https://www.nasm.us/).

### Usage

You can get the compiler via:
```sh
$ go get github.com/gmdvv2003/brainfuck-compiler-plus
```

To be able to use it, make sure you're inside the source folder:
```sh
$ cd /path/to/source
```

...And then you simply need to specify the name of your ``.bf`` file that's in the source directory. After doing so, it'll then generate an executable that you can run.
```sh
$ brainfuck-compiler-plus -file=example.bf
$ ./output
```

By default the generated files will be named "output", if you wish to change, you pass in ``-name`` to specify a different one.
```sh
$ brainfuck-compiler-plus -file=example.bf -name=brainfucker
$ ./brainfucker
```

You can also pass in a ``-debug=true`` flag; however, as of right now all it does it print the lexed tokens.

### Resources
[How to write a lexer in Go](https://www.aaronraff.dev/blog/how-to-write-a-lexer-in-go)

[Lexical Scanning in Go](https://www.youtube.com/watch?v=HxaD_trXwRE)
