#ifndef IOST_V8_COMPILE_H
#define IOST_V8_COMPILE_H

#include "sandbox.h"
#include "bignumber.h"
#include "int64.h"
#include "utils.h"
#include "console.h"
#include "storage.h"
#include "blockchain.h"

int compile(SandboxPtr, const char *code, const char **compiledCode);
CustomStartupData createStartupData();

#endif // IOST_V8_COMPILE_H