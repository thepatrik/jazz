#ifndef jazz_vm_h
#define jazz_vm_h

#include "object.h"
#include "table.h"
#include "value.h"

#define FRAMES_MAX 64
#define STACK_MAX (FRAMES_MAX * UINT8_COUNT)

#define GC_HEAP_GROW_FACTOR 2
#define GC_INITIAL_NEXT_GC  (1024 * 1024)  // 1 MiB

typedef struct {
    ObjClosure* closure;
    uint8_t* ip;
    Value* slots;  // window into vm.stack at this frame's base
} CallFrame;

typedef struct {
    CallFrame frames[FRAMES_MAX];
    int frameCount;
    Value stack[STACK_MAX];
    Value* stackTop;
    Table globals;
    Obj* objects;              // linked list of all heap objects
    ObjUpvalue* openUpvalues;  // head of the open-upvalue chain
    // GC state
    size_t bytesAllocated;
    size_t nextGC;
    int grayCount;
    int grayCapacity;
    Obj** grayStack;
} VM;

// Global VM instance — defined in vm.c.
extern VM vm;

typedef enum {
    INTERPRET_OK,
    INTERPRET_COMPILE_ERROR,
    INTERPRET_RUNTIME_ERROR,
} InterpretResult;

void initVM();
void freeVM();
InterpretResult interpret(const char* source);
InterpretResult interpretRepl(const char* source);
void push(Value value);
Value pop();

#endif
