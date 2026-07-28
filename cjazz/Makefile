MKDIR	:= mkdir
RMDIR	:= rm -rf /S /Q
CC		:= gcc
BIN		:= ./bin
OBJ		:= ./obj
INCLUDE	:= ./include
SRC		:= ./src
SRCS	:= $(wildcard $(SRC)/*.c)
OBJS	:= $(patsubst $(SRC)/%.c,$(OBJ)/%.o,$(SRCS))
EXE		:= $(BIN)/main.exe
CFLAGS	:= -I$(INCLUDE) -O2
LDLIBS	:= -lm

.PHONY: all run test format clean

all: $(EXE)

$(EXE): $(OBJS) | $(BIN)
	$(CC) $(LDFLAGS) $^ -o $@ $(LDLIBS)

$(OBJ)/%.o: $(SRC)/%.c | $(OBJ)
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