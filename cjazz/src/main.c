#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "common.h"
#include "linenoise.h"
#include "vm.h"

// Build the path to the persistent history file (~/.jazz_history) into buf.
// Returns buf on success, or NULL if $HOME is unset.
static char* historyPath(char* buf, size_t size) {
    const char* home = getenv("HOME");
    if (home == NULL)
        return NULL;
    snprintf(buf, size, "%s/.jazz_history", home);
    return buf;
}

static void repl() {
    char histPath[1024];
    bool haveHist = historyPath(histPath, sizeof(histPath)) != NULL;

    printf("Welcome to Jazz v0.0.1\nType \".exit\" to exit.\n");

    // linenoise falls back to a plain fgets-style read when stdin is not a
    // TTY, so piped/scripted input keeps working unchanged.
    if (haveHist)
        linenoiseHistoryLoad(histPath);

    char* line;
    while ((line = linenoise("> ")) != NULL) {
        if (strcmp(line, ".exit") == 0) {
            linenoiseFree(line);
            break;
        }
        if (line[0] != '\0') {
            linenoiseHistoryAdd(line);
            if (haveHist)
                linenoiseHistorySave(histPath);
        }
        interpret(line);
        linenoiseFree(line);
    }
}

static char* readFile(const char* path) {
    FILE* file = fopen(path, "rb");
    if (file == NULL) {
        fprintf(stderr, "Could not open file \"%s\".\n", path);
        exit(74);
    }

    fseek(file, 0L, SEEK_END);
    size_t fileSize = ftell(file);
    rewind(file);

    char* buffer = (char*)malloc(fileSize + 1);
    if (buffer == NULL) {
        fprintf(stderr, "Not enough memory to read \"%s\".\n", path);
        exit(74);
    }

    size_t bytesRead = fread(buffer, sizeof(char), fileSize, file);
    if (bytesRead < fileSize) {
        fprintf(stderr, "Could not read file \"%s\".\n", path);
        exit(74);
    }

    buffer[bytesRead] = '\0';
    fclose(file);
    return buffer;
}

static void runFile(const char* path) {
    char* source           = readFile(path);
    InterpretResult result = interpret(source);
    free(source);

    if (result == INTERPRET_COMPILE_ERROR)
        exit(65);
    if (result == INTERPRET_RUNTIME_ERROR)
        exit(70);
}

int main(int argc, const char* argv[]) {
    initVM();

    if (argc == 1) {
        repl();
    } else if (argc == 2) {
        runFile(argv[1]);
    } else {
        fprintf(stderr, "Usage: jazz [path]\n");
        exit(64);
    }

    freeVM();
    return 0;
}
