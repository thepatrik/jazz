#include "object.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "chunk.h"
#include "value.h"
#include "vm.h"

extern VM vm;

// Central allocator: every new object is immediately linked into vm.objects.
static Obj* allocObject(size_t size, ObjType type) {
    Obj* obj   = (Obj*)malloc(size);
    obj->type  = type;
    obj->next  = vm.objects;
    vm.objects = obj;
    return obj;
}

ObjFunction* newFunction() {
    ObjFunction* fn  = (ObjFunction*)allocObject(sizeof(ObjFunction), OBJ_FUNCTION);
    fn->arity        = 0;
    fn->upvalueCount = 0;
    fn->name         = NULL;
    initChunk(&fn->chunk);
    return fn;
}

ObjUpvalue* newUpvalue(Value* slot) {
    ObjUpvalue* upvalue = (ObjUpvalue*)allocObject(sizeof(ObjUpvalue), OBJ_UPVALUE);
    upvalue->location   = slot;
    upvalue->closed     = NIL_VAL;
    upvalue->next       = NULL;
    return upvalue;
}

ObjClosure* newClosure(ObjFunction* fn) {
    ObjUpvalue** upvalues = (ObjUpvalue**)malloc(sizeof(ObjUpvalue*) * fn->upvalueCount);
    for (int i = 0; i < fn->upvalueCount; i++) upvalues[i] = NULL;
    ObjClosure* closure   = (ObjClosure*)allocObject(sizeof(ObjClosure), OBJ_CLOSURE);
    closure->function     = fn;
    closure->upvalues     = upvalues;
    closure->upvalueCount = fn->upvalueCount;
    return closure;
}

ObjString* newString(const char* chars, int length) {
    ObjString* str = (ObjString*)allocObject(sizeof(ObjString), OBJ_STRING);
    str->length    = length;
    str->chars     = (char*)malloc(length + 1);
    memcpy(str->chars, chars, length);
    str->chars[length] = '\0';
    return str;
}

void freeObject(Obj* obj) {
    switch (obj->type) {
        case OBJ_FUNCTION: {
            ObjFunction* fn = (ObjFunction*)obj;
            freeChunk(&fn->chunk);
            free(fn->name);
            free(fn);
            break;
        }
        case OBJ_UPVALUE:
            free(obj);
            break;
        case OBJ_CLOSURE: {
            ObjClosure* closure = (ObjClosure*)obj;
            free(closure->upvalues);
            free(closure);
            break;
        }
        case OBJ_STRING: {
            ObjString* str = (ObjString*)obj;
            free(str->chars);
            free(str);
            break;
        }
    }
}

void printObject(Value value) {
    switch (OBJ_TYPE(value)) {
        case OBJ_FUNCTION: {
            ObjFunction* fn = AS_FUNCTION(value);
            if (fn->name == NULL) {
                printf("<script>");
            } else {
                printf("<fn %s>", fn->name);
            }
            break;
        }
        case OBJ_CLOSURE: {
            ObjClosure* closure = AS_CLOSURE(value);
            if (closure->function->name == NULL) {
                printf("<script>");
            } else {
                printf("<fn %s>", closure->function->name);
            }
            break;
        }
        case OBJ_UPVALUE:
            printf("<upvalue>");
            break;
        case OBJ_STRING:
            printf("%s", AS_CSTRING(value));
            break;
    }
}
