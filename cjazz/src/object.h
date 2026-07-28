#ifndef jazz_object_h
#define jazz_object_h

#include "chunk.h"
#include "common.h"
#include "value.h"

typedef enum {
    OBJ_FUNCTION,
} ObjType;

typedef struct Obj {
    ObjType type;
    struct Obj* next;
} Obj;

typedef struct {
    Obj obj;
    int arity;
    Chunk chunk;
    char* name;
} ObjFunction;

#define OBJ_TYPE(value) (AS_OBJ(value)->type)
#define IS_FUNCTION(value) (isObjType(value, OBJ_FUNCTION))
#define AS_FUNCTION(value) ((ObjFunction*)AS_OBJ(value))

static inline bool isObjType(Value value, ObjType type) {
    return IS_OBJ(value) && AS_OBJ(value)->type == type;
}

ObjFunction* newFunction();
void freeObject(Obj* obj);
void printObject(Value value);

#endif
