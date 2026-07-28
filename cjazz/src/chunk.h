#ifndef jazz_chunk_h
#define jazz_chunk_h

#include "common.h"
#include "value.h"

typedef enum {
    // Literals
    OP_CONSTANT,
    OP_NIL,
    OP_TRUE,
    OP_FALSE,

    // Unary
    OP_NEGATE,
    OP_NOT,

    // Arithmetic
    OP_ADD,
    OP_SUBTRACT,
    OP_MULTIPLY,
    OP_DIVIDE,

    // Comparison (!=, >=, <= are derived via OP_NOT)
    OP_EQUAL,
    OP_GREATER,
    OP_LESS,

    // Statements
    OP_POP,
    OP_PRINT,

    // Local variables (slot index operand)
    OP_GET_LOCAL,
    OP_SET_LOCAL,

    // Global variables (constant index operand)
    OP_DEFINE_GLOBAL,
    OP_GET_GLOBAL,
    OP_SET_GLOBAL,

    // Control flow (16-bit offset operand)
    OP_JUMP,
    OP_JUMP_IF_FALSE,
    OP_LOOP,

    // Functions
    OP_CALL,           // operand: argument count
    OP_CLOSURE,        // operand: function constant index, then 2*upvalueCount bytes
    OP_GET_UPVALUE,    // operand: upvalue slot
    OP_SET_UPVALUE,    // operand: upvalue slot
    OP_CLOSE_UPVALUE,  // no operand; close top-of-stack into its upvalue
    OP_RETURN,

    // Arrays
    OP_ARRAY,      // operand: element count (uint8)
    OP_GET_INDEX,  // pops index + array, pushes element
    OP_SET_INDEX,  // pops value/index/array, pushes value
} OpCode;

typedef struct {
    int count;
    int capacity;
    uint8_t* code;
    int* lines;
    ValueArray constants;
} Chunk;

void initChunk(Chunk* chunk);
void freeChunk(Chunk* chunk);
void writeChunk(Chunk* chunk, uint8_t byte, int line);
int addConstant(Chunk* chunk, Value value);

#endif
