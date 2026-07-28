#include "object.h"

#include <stdio.h>
#include <stdlib.h>

#include "chunk.h"
#include "value.h"

ObjFunction* newFunction() {
    ObjFunction* fn = (ObjFunction*)malloc(sizeof(ObjFunction));
    fn->obj.type    = OBJ_FUNCTION;
    fn->obj.next    = NULL;
    fn->arity       = 0;
    fn->name        = NULL;
    initChunk(&fn->chunk);
    return fn;
}

void freeObject(Obj* obj) {
    switch (obj->type) {
        case OBJ_FUNCTION: {
            ObjFunction* fn = (ObjFunction*)obj;
            freeChunk(&fn->chunk);
            free(fn);
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
    }
}
