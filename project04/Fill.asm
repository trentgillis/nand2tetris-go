// Runs an infinite loop that listens to the keyboard input.
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel. When no key is pressed,
// the screen should be cleared.

(MAIN_LOOP)
    @KBD
    D=M
    @COLOR_WHITE
    D;JEQ
    @COLOR_BLACK
    0;JMP

(COLOR_BLACK)
    // Sets @color to -1 before beginning the main fill loop
    @color
    M=-1
    @FILL_INIT
    0;JMP

(COLOR_WHITE)
    // Sets @color to 0 before beginning the main fill loop
    @color
    M=0
    @FILL_INIT
    0;JMP

(FILL_INIT)
    // Init @screen_ptr to @SCREEN + 8192 to iteratively fill all screen registers
    @SCREEN
    D=A-1
    @8192
    D=D+A
    @screen_ptr // used to store the screen address to be filled as we loop
    M=D
(FILL_LOOP)
    // Break out of FILL_LOOP if @screen_ptr == @SCREEN
    @SCREEN
    D=A
    @screen_ptr
    D=D-A
    @MAIN_LOOP
    D;JEQ
    // Set address at @screen_ptr to the current color
    @color
    D=M
    @screen_ptr
    A=M
    M=D
    // Decrement @screen_ptr
    @screen_ptr
    M=M-1
    @FILL_LOOP
    0;JMP
