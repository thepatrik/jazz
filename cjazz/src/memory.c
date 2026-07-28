#include "memory.h"

#include <stdlib.h>

#include "object.h"
#include "table.h"
#include "vm.h"

// ---- Core allocator --------------------------------------------------------

void* reallocate(void* pointer, size_t oldSize, size_t newSize) {
    // Track net allocation delta.
    if (newSize > oldSize)
        vm.bytesAllocated += newSize - oldSize;
    else
        vm.bytesAllocated -= oldSize - newSize;

    // Trigger GC only on growth; never on free/shrink (avoids re-entrance
    // during sweep, where freeObject calls FREE which calls reallocate).
    if (newSize > oldSize && vm.bytesAllocated > vm.nextGC)
        collectGarbage();

    if (newSize == 0) {
        free(pointer);
        return NULL;
    }
    void* result = realloc(pointer, newSize);
    if (result == NULL)
        exit(1);
    return result;
}

// ---- Mark phase ------------------------------------------------------------

void markObject(Obj* obj) {
    if (obj == NULL || obj->isMarked)
        return;
    obj->isMarked = true;
    // Grow the gray worklist using raw realloc so that the growth itself
    // does NOT call reallocate and cannot re-trigger collectGarbage.
    if (vm.grayCount >= vm.grayCapacity) {
        vm.grayCapacity = vm.grayCapacity < 8 ? 8 : vm.grayCapacity * 2;
        vm.grayStack    = (Obj**)realloc(vm.grayStack, sizeof(Obj*) * vm.grayCapacity);
        if (vm.grayStack == NULL)
            exit(1);
    }
    vm.grayStack[vm.grayCount++] = obj;
}

void markValue(Value value) {
    if (IS_OBJ(value))
        markObject(AS_OBJ(value));
}

static void markArray(ValueArray* array) {
    for (int i = 0; i < array->count; i++)
        markValue(array->values[i]);
}

static void blackenObject(Obj* obj) {
    switch (obj->type) {
        case OBJ_ARRAY:
            markArray(&((ObjArray*)obj)->elements);
            break;
        case OBJ_UPVALUE:
            // Open upvalue: location points into the stack (already a root).
            // Closed upvalue: 'closed' holds the captured value — mark it.
            markValue(((ObjUpvalue*)obj)->closed);
            break;
        case OBJ_FUNCTION: {
            ObjFunction* fn = (ObjFunction*)obj;
            // fn->name is a raw strndup char* — not a GC object, skip it.
            markArray(&fn->chunk.constants);
            break;
        }
        case OBJ_CLOSURE: {
            ObjClosure* closure = (ObjClosure*)obj;
            markObject((Obj*)closure->function);
            for (int i = 0; i < closure->upvalueCount; i++)
                markObject((Obj*)closure->upvalues[i]);
            break;
        }
        case OBJ_NATIVE:
        case OBJ_STRING:
            break;  // leaf nodes — no heap references
    }
}

// ---- GC roots --------------------------------------------------------------

static void markRoots() {
    // Value stack
    for (Value* slot = vm.stack; slot < vm.stackTop; slot++)
        markValue(*slot);
    // Active call frames
    for (int i = 0; i < vm.frameCount; i++)
        markObject((Obj*)vm.frames[i].closure);
    // Open upvalue chain
    for (ObjUpvalue* uv = vm.openUpvalues; uv != NULL; uv = uv->next)
        markObject((Obj*)uv);
    // Global variables
    for (int i = 0; i < vm.globals.capacity; i++) {
        Entry* entry = &vm.globals.entries[i];
        if (entry->key != NULL)
            markValue(entry->value);
    }
}

// ---- Trace phase -----------------------------------------------------------

static void traceReferences() {
    while (vm.grayCount > 0) {
        Obj* obj = vm.grayStack[--vm.grayCount];
        blackenObject(obj);
    }
}

// ---- Sweep phase -----------------------------------------------------------

static void sweep() {
    Obj* prev = NULL;
    Obj* obj  = vm.objects;
    while (obj != NULL) {
        if (obj->isMarked) {
            obj->isMarked = false;  // reset for next cycle
            prev          = obj;
            obj           = obj->next;
        } else {
            Obj* unreached = obj;
            obj            = obj->next;
            if (prev != NULL)
                prev->next = obj;
            else
                vm.objects = obj;
            freeObject(unreached);
        }
    }
}

// ---- Entry point -----------------------------------------------------------

void collectGarbage() {
    markRoots();
    traceReferences();
    sweep();
    vm.nextGC = vm.bytesAllocated * GC_HEAP_GROW_FACTOR;
}
