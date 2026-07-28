#ifndef jazz_object_h
#define jazz_object_h

#include "chunk.h"
#include "common.h"
#include "value.h"

typedef enum {
    OBJ_FUNCTION,
    OBJ_STRING,
} ObjType;

typedef struct Obj {
    ObjType type;
    struct Obj* next;
} Obj;

// ---- ObjFunction -----------------------------------------------------------

typedef struct {
    Obj obj;
    int arity;
    Chunk chunk;
    char* name;
} ObjFunction;

// ---- ObjString -------------------------------------------------------------

typedef struct {
    Obj obj;
    int length;
    char* chars;
} ObjString;

// ---- Type-check macros -----------------------------------------------------

#define OBJ_TYPE(value) (AS_OBJ(value)->type)
#define IS_FUNCTION(value) (isObjType(value, OBJ_FUNCTION))
#define IS_STRING(value) (isObjType(value, OBJ_STRING))
#define AS_FUNCTION(value) ((ObjFunction*)AS_OBJ(value))
#define AS_STRING(value) ((ObjString*)AS_OBJ(value))
#define AS_CSTRING(value) (((ObjString*)AS_OBJ(value))->chars)

static inline bool isObjType(Value value, ObjType type) {
    return IS_OBJ(value) && AS_OBJ(value)->type == type;
}

// ---- Constructors ----------------------------------------------------------

ObjFunction* newFunction();
ObjString* newString(const char* chars, int length);

// ---- Lifecycle -------------------------------------------------------------

void freeObject(Obj* obj);
void printObject(Value value);

#endif
