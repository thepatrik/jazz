#ifndef jazz_object_h
#define jazz_object_h

#include "chunk.h"
#include "common.h"
#include "value.h"

typedef enum {
    OBJ_CLOSURE,
    OBJ_FUNCTION,
    OBJ_NATIVE,
    OBJ_STRING,
    OBJ_UPVALUE,
} ObjType;

typedef struct Obj {
    ObjType type;
    struct Obj* next;
} Obj;

// ---- ObjFunction -----------------------------------------------------------

typedef struct {
    Obj obj;
    int arity;
    int upvalueCount;
    Chunk chunk;
    char* name;
} ObjFunction;

// ---- ObjUpvalue ------------------------------------------------------------
// While the captured variable lives on the stack, location points there.
// Once the enclosing scope exits, the value is copied into 'closed' and
// location is redirected to point at the field inside this struct.

typedef struct ObjUpvalue {
    Obj obj;
    Value* location;
    Value closed;
    struct ObjUpvalue* next;
} ObjUpvalue;

// ---- ObjClosure ------------------------------------------------------------

typedef struct {
    Obj obj;
    ObjFunction* function;
    ObjUpvalue** upvalues;
    int upvalueCount;
} ObjClosure;

// ---- ObjNative -------------------------------------------------------------

typedef Value (*NativeFn)(int argCount, Value* args);

typedef struct {
    Obj obj;
    NativeFn function;
} ObjNative;

// ---- ObjString -------------------------------------------------------------

typedef struct {
    Obj obj;
    int length;
    char* chars;
} ObjString;

// ---- Type-check macros -----------------------------------------------------

#define OBJ_TYPE(value) (AS_OBJ(value)->type)
#define IS_CLOSURE(value) (isObjType(value, OBJ_CLOSURE))
#define IS_FUNCTION(value) (isObjType(value, OBJ_FUNCTION))
#define IS_NATIVE(value) (isObjType(value, OBJ_NATIVE))
#define IS_STRING(value) (isObjType(value, OBJ_STRING))
#define IS_UPVALUE(value) (isObjType(value, OBJ_UPVALUE))
#define AS_CLOSURE(value) ((ObjClosure*)AS_OBJ(value))
#define AS_FUNCTION(value) ((ObjFunction*)AS_OBJ(value))
#define AS_NATIVE(value) ((ObjNative*)AS_OBJ(value))
#define AS_STRING(value) ((ObjString*)AS_OBJ(value))
#define AS_CSTRING(value) (((ObjString*)AS_OBJ(value))->chars)

static inline bool isObjType(Value value, ObjType type) {
    return IS_OBJ(value) && AS_OBJ(value)->type == type;
}

// ---- Constructors ----------------------------------------------------------

ObjFunction* newFunction();
ObjUpvalue* newUpvalue(Value* slot);
ObjClosure* newClosure(ObjFunction* fn);
ObjNative* newNative(NativeFn function);
ObjString* newString(const char* chars, int length);

// ---- Lifecycle -------------------------------------------------------------

void freeObject(Obj* obj);
void printObject(Value value);

#endif
