#ifndef jazz_value_h
#define jazz_value_h

#include "common.h"

// Forward declaration – full definition in object.h
struct Obj;

typedef enum {
    VAL_BOOL,
    VAL_NIL,
    VAL_NUMBER,
    VAL_IDENT,  // internal: identifier names stored in the constants pool
    VAL_OBJ,
} ValueType;

typedef struct {
    ValueType type;
    union {
        bool boolean;
        double number;
        char* ident;
        struct Obj* obj;
    } as;
} Value;

#define IS_BOOL(value) ((value).type == VAL_BOOL)
#define IS_NIL(value) ((value).type == VAL_NIL)
#define IS_NUMBER(value) ((value).type == VAL_NUMBER)
#define IS_IDENT(value) ((value).type == VAL_IDENT)
#define IS_OBJ(value) ((value).type == VAL_OBJ)

#define AS_BOOL(value) ((value).as.boolean)
#define AS_NUMBER(value) ((value).as.number)
#define AS_IDENT(value) ((value).as.ident)
#define AS_OBJ(value) ((value).as.obj)

#define BOOL_VAL(value) ((Value){VAL_BOOL, {.boolean = value}})
#define NIL_VAL ((Value){VAL_NIL, {.number = 0}})
#define NUMBER_VAL(value) ((Value){VAL_NUMBER, {.number = value}})
#define IDENT_VAL(value) ((Value){VAL_IDENT, {.ident = value}})
#define OBJ_VAL(value) ((Value){VAL_OBJ, {.obj = (struct Obj*)(value)}})

typedef struct {
    int capacity;
    int count;
    Value* values;
} ValueArray;

void initValueArray(ValueArray* array);
void writeValueArray(ValueArray* array, Value value);
void freeValueArray(ValueArray* array);
void printValue(Value value);

#endif
