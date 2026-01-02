// push constant 7
@7
D=A
@SP
A=M
M=D
@SP
M=M+1
// push constant 8
@8
D=A
// push
@SP
A=M
M=D
@SP
M=M+1
// add
// pop to R13
@SP
A=M-1
D=M
@R13
M=D
@SP
M=M-1
// pop to R14
@SP
A=M-1
D=M
@R14
M=D
@SP
M=M-1
// do the add and push to stack
@R13
D=M
@R14
D=D+M
@SP
A=M
M=D
@SP
M=M+1
