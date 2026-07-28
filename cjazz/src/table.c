#include "table.h"

#include <stdlib.h>
#include <string.h>

#include "memory.h"

#define TABLE_MAX_LOAD 0.75

static uint32_t hashString(const char* key) {
    uint32_t hash = 2166136261u;
    for (int i = 0; key[i] != '\0'; i++) {
        hash ^= (uint8_t)key[i];
        hash *= 16777619;
    }
    return hash;
}

void initTable(Table* table) {
    table->count    = 0;
    table->capacity = 0;
    table->entries  = NULL;
}

void freeTable(Table* table) {
    for (int i = 0; i < table->capacity; i++) {
        if (table->entries[i].key != NULL) {
            free(table->entries[i].key);
        }
    }
    FREE_ARRAY(Entry, table->entries, table->capacity);
    initTable(table);
}

static Entry* findEntry(Entry* entries, int capacity, const char* key) {
    uint32_t index   = hashString(key) % (uint32_t)capacity;
    Entry* tombstone = NULL;
    for (;;) {
        Entry* entry = &entries[index];
        if (entry->key == NULL) {
            if (IS_NIL(entry->value)) {
                return tombstone != NULL ? tombstone : entry;
            }
            if (tombstone == NULL)
                tombstone = entry;
        } else if (strcmp(entry->key, key) == 0) {
            return entry;
        }
        index = (index + 1) % (uint32_t)capacity;
    }
}

static void adjustCapacity(Table* table, int capacity) {
    Entry* entries = GROW_ARRAY(Entry, NULL, 0, capacity);
    for (int i = 0; i < capacity; i++) {
        entries[i].key   = NULL;
        entries[i].value = NIL_VAL;
    }

    table->count = 0;
    for (int i = 0; i < table->capacity; i++) {
        Entry* entry = &table->entries[i];
        if (entry->key == NULL)
            continue;
        Entry* dest = findEntry(entries, capacity, entry->key);
        dest->key   = entry->key;
        dest->value = entry->value;
        table->count++;
    }

    FREE_ARRAY(Entry, table->entries, table->capacity);
    table->entries  = entries;
    table->capacity = capacity;
}

bool tableGet(Table* table, const char* key, Value* value) {
    if (table->count == 0)
        return false;
    Entry* entry = findEntry(table->entries, table->capacity, key);
    if (entry->key == NULL)
        return false;
    *value = entry->value;
    return true;
}

bool tableSet(Table* table, const char* key, Value value) {
    if (table->count + 1 > table->capacity * TABLE_MAX_LOAD) {
        adjustCapacity(table, GROW_CAPACITY(table->capacity));
    }
    Entry* entry = findEntry(table->entries, table->capacity, key);
    bool isNew   = (entry->key == NULL);
    if (isNew && IS_NIL(entry->value))
        table->count++;
    if (isNew)
        entry->key = strdup(key);
    entry->value = value;
    return isNew;
}

bool tableDelete(Table* table, const char* key) {
    if (table->count == 0)
        return false;
    Entry* entry = findEntry(table->entries, table->capacity, key);
    if (entry->key == NULL)
        return false;
    free(entry->key);
    entry->key   = NULL;
    entry->value = BOOL_VAL(true);  // tombstone
    return true;
}
