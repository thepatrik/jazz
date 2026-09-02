MKDIR	:= mkdir
RMDIR	:= rm -rf /S /Q
CC		:= gcc
BIN		:= ./bin
OBJ		:= ./obj
INCLUDE	:= ./include
SRC		:= ./src
VENDOR	:= ./vendor/linenoise
SRCS	:= $(wildcard $(SRC)/*.c)
OBJS	:= $(patsubst $(SRC)/%.c,$(OBJ)/%.o,$(SRCS)) $(OBJ)/linenoise.o
EXE		:= $(BIN)/jazz
CFLAGS	:= -I$(INCLUDE) -I$(VENDOR) -O2
LDLIBS	:= -lm

.DEFAULT_GOAL := run

.PHONY: all run test format clean

all: $(EXE)

$(EXE): $(OBJS) | $(BIN)
	$(CC) $(LDFLAGS) $^ -o $@ $(LDLIBS)

$(OBJ)/%.o: $(SRC)/%.c | $(OBJ)
	$(CC) $(CFLAGS) -c $< -o $@

# Vendored linenoise (line-editing library). Built separately from $(SRC).
$(OBJ)/linenoise.o: $(VENDOR)/linenoise.c | $(OBJ)
	$(CC) $(CFLAGS) -c $< -o $@

$(BIN) $(OBJ):
	$(MKDIR) $@

run: $(EXE)
	$<

test: $(EXE)
	bash test.sh

format:
	clang-format -i $(SRC)/*.c $(SRC)/*.h

clean:
	$(RMDIR) $(OBJ) $(BIN)