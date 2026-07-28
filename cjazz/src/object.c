#include "object.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "chunk.h"
#include "value.h"
#include "vm.h"

extern VM vm;

// Central allocator: every new object is immediately linked into vm.objects
// so freeVM can walk the list and release everything.
static Obj* allocObject(size_t size, ObjType type) {
    Obj* obj   = (Obj*)malloc(size);
    obj->type  = type;
    obj->next  = vm.objects;
    vm.objects = obj;
    return obj;
}

ObjFunction* newFunction() {
    ObjFunction* fn = (ObjFunction*)allocObject(sizeof(ObjFunction), OBJ_FUNCTION);
    fn->arity       = 0;
    fn->name        = NULL;
    initChunk(&fn->chunk);
    return fn;
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
            free(fn);
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
        case OBJ_STRING:
            printf("%s", AS_CSTRING(value));
            break;
    }
}
