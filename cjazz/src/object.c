#include "object.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "chunk.h"
#include "memory.h"
#include "value.h"
#include "vm.h"

// Central allocator: every new object is immediately linked into vm.objects.
static Obj* allocObject(size_t size, ObjType type) {
    Obj* obj      = (Obj*)reallocate(NULL, 0, size);
    obj->type     = type;
    obj->isMarked = false;
    obj->next     = vm.objects;
    vm.objects    = obj;
    return obj;
}

ObjArray* newArray() {
    ObjArray* array = (ObjArray*)allocObject(sizeof(ObjArray), OBJ_ARRAY);
    initValueArray(&array->elements);
    return array;
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
    // Allocate the upvalue pointer array through reallocate so its bytes are
    // counted.  Guard against fn->upvalueCount == 0 to avoid a zero-size
    // GROW_ARRAY call (which reallocate treats as a free, returning NULL).
    ObjUpvalue** upvalues = fn->upvalueCount > 0
                                ? GROW_ARRAY(ObjUpvalue*, NULL, 0, fn->upvalueCount)
                                : NULL;
    for (int i = 0; i < fn->upvalueCount; i++) upvalues[i] = NULL;
    ObjClosure* closure   = (ObjClosure*)allocObject(sizeof(ObjClosure), OBJ_CLOSURE);
    closure->function     = fn;
    closure->upvalues     = upvalues;
    closure->upvalueCount = fn->upvalueCount;
    return closure;
}

ObjNative* newNative(NativeFn function) {
    ObjNative* native = (ObjNative*)allocObject(sizeof(ObjNative), OBJ_NATIVE);
    native->function  = function;
    return native;
}

ObjString* newString(const char* chars, int length) {
    ObjString* str = (ObjString*)allocObject(sizeof(ObjString), OBJ_STRING);
    str->length    = length;
    str->chars     = (char*)reallocate(NULL, 0, length + 1);
    memcpy(str->chars, chars, length);
    str->chars[length] = '\0';
    return str;
}

void freeObject(Obj* obj) {
    switch (obj->type) {
        case OBJ_ARRAY: {
            ObjArray* array = (ObjArray*)obj;
            freeValueArray(&array->elements);
            FREE(ObjArray, array);
            break;
        }
        case OBJ_NATIVE:
            FREE(ObjNative, obj);
            break;
        case OBJ_FUNCTION: {
            ObjFunction* fn = (ObjFunction*)obj;
            freeChunk(&fn->chunk);
            // fn->name was set by strndup (not reallocate) — free directly.
            free(fn->name);
            FREE(ObjFunction, fn);
            break;
        }
        case OBJ_UPVALUE:
            FREE(ObjUpvalue, obj);
            break;
        case OBJ_CLOSURE: {
            ObjClosure* closure = (ObjClosure*)obj;
            FREE_ARRAY(ObjUpvalue*, closure->upvalues, closure->upvalueCount);
            FREE(ObjClosure, closure);
            break;
        }
        case OBJ_STRING: {
            ObjString* str = (ObjString*)obj;
            FREE_ARRAY(char, str->chars, str->length + 1);
            FREE(ObjString, str);
            break;
        }
    }
}

void printObject(Value value) {
    switch (OBJ_TYPE(value)) {
        case OBJ_ARRAY: {
            ObjArray* array = AS_ARRAY(value);
            printf("[");
            for (int i = 0; i < array->elements.count; i++) {
                if (i > 0) printf(", ");
                printValue(array->elements.values[i]);
            }
            printf("]");
            break;
        }
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
        case OBJ_NATIVE:
            printf("<native fn>");
            break;
        case OBJ_UPVALUE:
            printf("<upvalue>");
            break;
        case OBJ_STRING:
            printf("%s", AS_CSTRING(value));
            break;
    }
}
