#include "vm.h"

#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "common.h"
#include "compiler.h"
#include "debug.h"
#include "object.h"

VM vm;

static void resetStack() {
    vm.stackTop   = vm.stack;
    vm.frameCount = 0;
}

void initVM() {
    resetStack();
    vm.objects = NULL;
    initTable(&vm.globals);
}

void freeVM() {
    freeTable(&vm.globals);
    Obj* obj = vm.objects;
    while (obj) {
        Obj* next = obj->next;
        freeObject(obj);
        obj = next;
    }
}

void push(Value value) {
    *vm.stackTop++ = value;
}
Value pop() {
    return *--vm.stackTop;
}

static Value peek(int distance) {
    return vm.stackTop[-1 - distance];
}

static bool isFalsy(Value value) {
    return IS_NIL(value) || (IS_BOOL(value) && !AS_BOOL(value));
}

static bool valuesEqual(Value a, Value b) {
    if (a.type != b.type)
        return false;
    switch (a.type) {
        case VAL_BOOL:
            return AS_BOOL(a) == AS_BOOL(b);
        case VAL_NIL:
            return true;
        case VAL_NUMBER:
            return AS_NUMBER(a) == AS_NUMBER(b);
        case VAL_OBJ: {
            if (OBJ_TYPE(a) != OBJ_TYPE(b))
                return false;
            if (IS_STRING(a)) {
                ObjString* sa = AS_STRING(a);
                ObjString* sb = AS_STRING(b);
                return sa->length == sb->length && memcmp(sa->chars, sb->chars, sa->length) == 0;
            }
            return AS_OBJ(a) == AS_OBJ(b);
        }
        default:
            return false;
    }
}

static void runtimeError(const char* format, ...) {
    fprintf(stderr, "Runtime error: ");
    va_list args;
    va_start(args, format);
    vfprintf(stderr, format, args);
    va_end(args);
    fputs("\n", stderr);

    // Print stack trace (innermost first).
    for (int i = vm.frameCount - 1; i >= 0; i--) {
        CallFrame* frame = &vm.frames[i];
        ObjFunction* fn  = frame->function;
        int line         = fn->chunk.lines[(int)(frame->ip - fn->chunk.code - 1)];
        fprintf(stderr, "[line %d] in ", line);
        if (fn->name == NULL) {
            fprintf(stderr, "script\n");
        } else {
            fprintf(stderr, "%s()\n", fn->name);
        }
    }

    resetStack();
}

static InterpretResult run() {
    CallFrame* frame = &vm.frames[vm.frameCount - 1];

#define READ_BYTE() (*frame->ip++)
#define READ_SHORT() (frame->ip += 2, (uint16_t)((frame->ip[-2] << 8) | frame->ip[-1]))
#define READ_CONSTANT() (frame->function->chunk.constants.values[READ_BYTE()])
#define BINARY_OP(valueType, op)                          \
    do {                                                  \
        if (!IS_NUMBER(peek(0)) || !IS_NUMBER(peek(1))) { \
            runtimeError("Operands must be numbers.");    \
            return INTERPRET_RUNTIME_ERROR;               \
        }                                                 \
        double b = AS_NUMBER(pop());                      \
        double a = AS_NUMBER(pop());                      \
        push(valueType(a op b));                          \
    } while (false)

    for (;;) {
#ifdef DEBUG_TRACE_EXECUTION
        printf("          ");
        for (Value* slot = vm.stack; slot < vm.stackTop; slot++) {
            printf("[ ");
            printValue(*slot);
            printf(" ]");
        }
        printf("\n");
        disassembleInstruction(&frame->function->chunk,
                               (int)(frame->ip - frame->function->chunk.code));
#endif
        uint8_t instruction;
        switch (instruction = READ_BYTE()) {
            // --- Literals ---
            case OP_CONSTANT:
                push(READ_CONSTANT());
                break;
            case OP_NIL:
                push(NIL_VAL);
                break;
            case OP_TRUE:
                push(BOOL_VAL(true));
                break;
            case OP_FALSE:
                push(BOOL_VAL(false));
                break;

            // --- Unary ---
            case OP_NEGATE:
                if (!IS_NUMBER(peek(0))) {
                    runtimeError("Operand must be a number.");
                    return INTERPRET_RUNTIME_ERROR;
                }
                push(NUMBER_VAL(-AS_NUMBER(pop())));
                break;
            case OP_NOT:
                push(BOOL_VAL(isFalsy(pop())));
                break;

            // --- Arithmetic ---
            case OP_ADD: {
                if (IS_STRING(peek(0)) && IS_STRING(peek(1))) {
                    ObjString* b = AS_STRING(pop());
                    ObjString* a = AS_STRING(pop());
                    int len      = a->length + b->length;
                    char* buf    = (char*)malloc(len + 1);
                    memcpy(buf, a->chars, a->length);
                    memcpy(buf + a->length, b->chars, b->length);
                    buf[len] = '\0';
                    push(OBJ_VAL(newString(buf, len)));
                    free(buf);
                } else if (IS_NUMBER(peek(0)) && IS_NUMBER(peek(1))) {
                    double b = AS_NUMBER(pop());
                    double a = AS_NUMBER(pop());
                    push(NUMBER_VAL(a + b));
                } else {
                    runtimeError("Operands must be two numbers or two strings.");
                    return INTERPRET_RUNTIME_ERROR;
                }
                break;
            }
            case OP_SUBTRACT:
                BINARY_OP(NUMBER_VAL, -);
                break;
            case OP_MULTIPLY:
                BINARY_OP(NUMBER_VAL, *);
                break;
            case OP_DIVIDE: {
                if (!IS_NUMBER(peek(0)) || !IS_NUMBER(peek(1))) {
                    runtimeError("Operands must be numbers.");
                    return INTERPRET_RUNTIME_ERROR;
                }
                double b = AS_NUMBER(pop());
                double a = AS_NUMBER(pop());
                if (b == 0) {
                    runtimeError("Division by zero.");
                    return INTERPRET_RUNTIME_ERROR;
                }
                push(NUMBER_VAL(a / b));
                break;
            }

            // --- Comparison ---
            case OP_EQUAL: {
                Value b = pop(), a = pop();
                push(BOOL_VAL(valuesEqual(a, b)));
                break;
            }
            case OP_GREATER:
                BINARY_OP(BOOL_VAL, >);
                break;
            case OP_LESS:
                BINARY_OP(BOOL_VAL, <);
                break;

            // --- Control flow ---
            case OP_JUMP: {
                uint16_t offset = READ_SHORT();
                frame->ip += offset;
                break;
            }
            case OP_JUMP_IF_FALSE: {
                uint16_t offset = READ_SHORT();
                if (isFalsy(peek(0)))
                    frame->ip += offset;
                break;
            }
            case OP_LOOP: {
                uint16_t offset = READ_SHORT();
                frame->ip -= offset;
                break;
            }

            // --- Local variables ---
            case OP_GET_LOCAL: {
                uint8_t slot = READ_BYTE();
                push(frame->slots[slot]);
                break;
            }
            case OP_SET_LOCAL: {
                uint8_t slot       = READ_BYTE();
                frame->slots[slot] = peek(0);  // assign but don't pop
                break;
            }

            // --- Statements ---
            case OP_POP:
                pop();
                break;
            case OP_PRINT:
                printValue(pop());
                printf("\n");
                break;

            // --- Global variables ---
            case OP_DEFINE_GLOBAL: {
                const char* name = AS_IDENT(READ_CONSTANT());
                tableSet(&vm.globals, name, peek(0));
                pop();
                break;
            }
            case OP_GET_GLOBAL: {
                const char* name = AS_IDENT(READ_CONSTANT());
                Value value;
                if (!tableGet(&vm.globals, name, &value)) {
                    runtimeError("Undefined variable '%s'.", name);
                    return INTERPRET_RUNTIME_ERROR;
                }
                push(value);
                break;
            }
            case OP_SET_GLOBAL: {
                const char* name = AS_IDENT(READ_CONSTANT());
                if (tableSet(&vm.globals, name, peek(0))) {
                    tableDelete(&vm.globals, name);
                    runtimeError("Undefined variable '%s'.", name);
                    return INTERPRET_RUNTIME_ERROR;
                }
                break;
            }

            // --- Functions ---
            case OP_CALL: {
                int argCount = READ_BYTE();
                Value callee = peek(argCount);
                if (!IS_FUNCTION(callee)) {
                    runtimeError("Can only call functions.");
                    return INTERPRET_RUNTIME_ERROR;
                }
                ObjFunction* fn = AS_FUNCTION(callee);
                if (argCount != fn->arity) {
                    runtimeError("Expected %d arguments but got %d.", fn->arity, argCount);
                    return INTERPRET_RUNTIME_ERROR;
                }
                if (vm.frameCount == FRAMES_MAX) {
                    runtimeError("Stack overflow.");
                    return INTERPRET_RUNTIME_ERROR;
                }
                CallFrame* newFrame = &vm.frames[vm.frameCount++];
                newFrame->function  = fn;
                newFrame->ip        = fn->chunk.code;
                newFrame->slots     = vm.stackTop - argCount - 1;
                frame               = &vm.frames[vm.frameCount - 1];
                break;
            }
            case OP_RETURN: {
                Value result = pop();
                vm.frameCount--;
                if (vm.frameCount == 0) {
                    pop();  // pop the script function
                    return INTERPRET_OK;
                }
                vm.stackTop = frame->slots;
                push(result);
                frame = &vm.frames[vm.frameCount - 1];
                break;
            }
        }
    }

#undef READ_BYTE
#undef READ_SHORT
#undef READ_CONSTANT
#undef BINARY_OP
}

InterpretResult interpret(const char* source) {
    ObjFunction* fn = compile(source);
    if (fn == NULL)
        return INTERPRET_COMPILE_ERROR;

    push(OBJ_VAL(fn));
    CallFrame* frame = &vm.frames[vm.frameCount++];
    frame->function  = fn;
    frame->ip        = fn->chunk.code;
    frame->slots     = vm.stack;

    return run();
}
