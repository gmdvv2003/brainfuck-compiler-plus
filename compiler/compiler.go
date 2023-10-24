package compiler

import (
	"bytes"
	"fmt"

	"github.com/gmdvv2003/brainfuck-compiler-plus/parser"
)

const (
	programHeader = `
segment .bss
	tape 			resq 30000	; Define uninitialized memory for the tape array
	cell_pointer 	resq 1		; Assign one byte for the cell pointer
	print_buffer	resb 1		; 1 byte buffer for the print result
	read_buffer		resb 1		; 1 byte buffer for the read syscall output

section .data
	linefeed 		db 0x0A		; ASCII character 10, a line break

segment .text
global _start

; Used to print the cell that cell_pointer points to
print_cell:
	mov rbx, [cell_pointer]	; Load cell_pointer into rbx
	mov rax, [tape+rbx*8]	; Load the value of tape[cell_pointer] into rax
	
	sub rax, 48				; 48 is the code for "0", where it was starting at
	cmp rax, 10				; Checks if the entered character code is a linefeed
	
	je .is_linefeed
	jmp .is_not_linefeed

.is_linefeed
	mov al, [linefeed]		; Move the linefeed byte representation into al
	mov [print_buffer], al 	; Move al into the buffer

.is_not_linefeed
	add al, '0'				; Convert to ASCII
	mov [print_buffer], al	; Move the result to the buffer

.perform_print
	; Perform the syscall
	mov rax, 1
	mov rdi, 1
	mov rsi, print_buffer
	mov rdx, 1
	syscall

	ret

; Read the output and saves it to the cell that cell_pointer points to
input_cell:
	; Perform the syscall
	mov rax, 0
	mov rdi, 0
	mov rsi, read_buffer
	mov rdx, 1
	syscall

	; Save the entered character
	mov rbx, [cell_pointer]	; Load the cell pointer into rbx
	mov al, [read_buffer]	; Load the read result into al
	mov [tape+rbx*8], al	; Move the buffer into the index tape[cell_pointer]

	ret

_start:
`
)

// Internal function used to compile a sequence of nodes into assembly
func compile(ast *parser.AST, offset *int, instructions *bytes.Buffer) ([]byte, error) {
	for _, node := range ast.Nodes {
		// Increment the offset beforehand to avoid duplicated labels names due to recursion
		*offset += 1

		switch assertedNode := node.(type) {
		case *parser.NextCell:
			instructions.WriteString("    ; -- Next Cell -- ;\n")
			instructions.WriteString("    mov rbx, [cell_pointer]\n")                              // Load cell_pointer into rbx
			instructions.WriteString(fmt.Sprintf("    add rbx, %d\n", assertedNode.Node.Repeat+1)) // Increment rbx by repeat count
			instructions.WriteString("    mov [cell_pointer], rbx\n")                              // Set cell_pointer to rbx
		case *parser.PreviousCell:
			instructions.WriteString("    ; -- Previous Cell -- ;\n")
			instructions.WriteString("    mov rbx, [cell_pointer]\n")                              // Load cell_pointer into rbx
			instructions.WriteString(fmt.Sprintf("    sub rbx, %d\n", assertedNode.Node.Repeat+1)) // Decrement rbx by repeat count
			instructions.WriteString("    mov [cell_pointer], rbx\n")                              // Set cell_pointer to rbx
		case *parser.IncrementCell:
			instructions.WriteString("    ; -- Increment -- ;\n")
			instructions.WriteString("    mov rbx, [cell_pointer]\n")                              // Load the cell pointer into rbx
			instructions.WriteString("    mov rax, [tape+rbx*8]\n")                                // Load the value of tape[cell_pointer] into rax
			instructions.WriteString(fmt.Sprintf("    add rax, %d\n", assertedNode.Node.Repeat+1)) // Increment rax by repeat count
			instructions.WriteString("    mov [tape+rbx*8], rax\n")                                // Move rax into the index tape[cell_pointer]
		case *parser.DecrementCell:
			instructions.WriteString("    ; -- Decrement -- ;\n")
			instructions.WriteString("    mov rbx, [cell_pointer]\n")                              // Load the cell pointer into rbx
			instructions.WriteString("    mov rax, [tape+rbx*8]\n")                                // Load the value of tape[cell_pointer] into rax
			instructions.WriteString(fmt.Sprintf("    sub rax, %d\n", assertedNode.Node.Repeat+1)) // Decrement rax by repeat count
			instructions.WriteString("    mov [tape+rbx*8], rax\n")                                // Move rax into the index tape[cell_pointer]
		case *parser.OutputCell:
			instructions.WriteString("    ; -- Output -- ;\n")
			instructions.WriteString("    call print_cell\n")
		case *parser.InputCell:
			instructions.WriteString("    ; -- Input -- ;\n")
			instructions.WriteString("    call input_cell\n")
		case *parser.NodeWhile:
			var whileOffset = *offset

			instructions.WriteString("    ; -- While -- ;\n")
			instructions.WriteString(fmt.Sprintf("while_%d:\n", whileOffset))
			instructions.WriteString("    mov rbx, [cell_pointer]\n")                         // Load the cell pointer into rbx
			instructions.WriteString("    mov rax, [tape+rbx*8]\n")                           // Load the value of tape[cell_pointer] into rax
			instructions.WriteString("    cmp rax, 0\n")                                      // Compare rax to 0
			instructions.WriteString(fmt.Sprintf("    jnz while_%d.not_zero\n", whileOffset)) // Jump to .not_zero if rax is not zero
			instructions.WriteString(fmt.Sprintf("    jmp while_%d.done\n", whileOffset))     // Jump to .done if rax is zero
			instructions.WriteString(fmt.Sprintf("while_%d.not_zero\n", whileOffset))
			instructions.WriteString("    ; -- While Begin -- ;\n")

			// Recursively compile the while block
			if bytes, ok := compile(assertedNode.Tree, offset, new(bytes.Buffer)); ok == nil {
				for _, instruction := range bytes {
					instructions.WriteByte(instruction)
				}
			} else {
				return nil, fmt.Errorf("error while trying to compile while block: %s", ok.Error())
			}

			instructions.WriteString("    ; -- While End -- ;\n")
			instructions.WriteString(fmt.Sprintf("    jmp while_%d\n", whileOffset)) // Jump to the beginning of the while block
			instructions.WriteString(fmt.Sprintf("while_%d.done\n", whileOffset))
		default:
			// return nil, fmt.Errorf("invalid node type %T", node)
		}
	}

	return instructions.Bytes(), nil
}

// External function used to compile a sequence of nodes into assembly
func Compile(ast *parser.AST) ([]byte, error) {
	// Write the header
	var instructions bytes.Buffer
	instructions.WriteString(programHeader)

	if _, ok := compile(ast, new(int), &instructions); ok != nil {
		return nil, ok
	}

	instructions.WriteString("    ; -- New Line -- ;\n")
	instructions.WriteString("    mov rax, 1\n")
	instructions.WriteString("    mov rdi, 1\n")
	instructions.WriteString("    mov rsi, linefeed\n")
	instructions.WriteString("    mov rdx, 1\n")
	instructions.WriteString("    syscall\n")

	// Write the exit instructions
	instructions.WriteString("    ; -- Exit -- ;\n")
	instructions.WriteString("    mov rax, 60\n")
	instructions.WriteString("    xor rdi, rdi\n")
	instructions.WriteString("    syscall\n")

	return instructions.Bytes(), nil
}
