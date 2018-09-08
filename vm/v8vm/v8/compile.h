#ifndef IOST_V8_COMPILE_H
#define IOST_V8_COMPILE_H

#include "sandbox.h"
#include "console.h"
#include "storage.h"
#include "blockchain.h"
#include "instruction.h"

intptr_t externalRef[] = {
        reinterpret_cast<intptr_t>(NewConsoleLog),
        reinterpret_cast<intptr_t>(NewIOSTContractStorage),
        reinterpret_cast<intptr_t>(NewIOSTBlockchain),
        reinterpret_cast<intptr_t>(NewIOSTContractInstruction),
        0};

int compile(SandboxPtr, const char *code, const char **compiledCode);
CustomStartupData createStartupData();

#endif // IOST_V8_COMPILE_H