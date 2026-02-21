.DEFAULT_GOAL := compile

.PHONY:compile

compile:
	cp ./Array.jack test/ArrayTest
	./tools/JackCompiler.sh ./test/ArrayTest
	cp ./Math.jack test/MathTest
	./tools/JackCompiler.sh ./test/MathTest
	cp ./Memory.jack test/MemoryTest
	./tools/JackCompiler.sh ./test/MemoryTest
