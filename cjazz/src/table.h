#ifndef jazz_table_h
#define jazz_table_h

#include "common.h"
#include "value.h"

typedef struct {
    char*  key;
    Value  value;
} Entry;

typedef struct {
    int    count;
    int    capacity;
    Entry* entries;
} Table;

void initTable(Table* table);
void freeTable(Table* table);
bool tableGet(Table* table, const char* key, Value* value);
bool tableSet(Table* table, const char* key, Value value);
bool tableDelete(Table* table, const char* key);

#endif
