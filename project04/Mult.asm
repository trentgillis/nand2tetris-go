// Multiplies R0 and R1 and stores the result in R2.
// (R0, R1, R2 refer to RAM[0], RAM[1], and RAM[2], respectively.)
// The algorithm is based on repetitive addition.

// Initialize R2 (our result) to 0
@R2
M=0

// Return 0 if R0 or R1 are 0
@R0
D=M
@END
D;JEQ
@R1
D=M
@END
D;JEQ

(LOOP)
    // Loop until R1 is 0
    @R0
    M=M-1
    D=M
    @END
    D;JLT
    // Update multiplication result
    @R1
    D=M
    @R2
    M=D+M
    @LOOP
    0;JMP
(END)
    @END
    0;JMP
