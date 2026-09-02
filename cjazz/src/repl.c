#include "repl.h"

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

void repl(void) {
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
        interpretRepl(line);
        linenoiseFree(line);
    }
}
