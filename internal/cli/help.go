package cli

import (
	"fmt"
)

func Help() {
	fmt.Println(`beam - command-line file transfer utility for local network 

usage: beam [-v | --version] [-h | --help]
            <command> <arg>

available commands:
    emit	send a file to other machine
    absorb	receive a file from other machine

emit <path>
    send a file to other machine. requires single <path> argument,
    which is a relative or absolute path to a file.

    usage examples:
	beam emit foo.bar
	beam emit ~/Documents/books/book1.pdf

absorb <code>
    receive a file from other machine. requires single <code> argument,
    which contains sender address information and a randomly generated
    code used for veryfication.

    usage examples:
	beam absorb 61f0
	beam absorb 010e85
	`)
}

func Usage() {
	fmt.Println(` usage: beam [-v | --version] [-h | --help]
            <command> <arg>

available commands:
    emit	send a file to other machine
    absorb	receive a file from other machine
	`)
}
