#ifndef IOST_V8_BLOCKCHAIN_H
#define IOST_V8_BLOCKCHAIN_H

#include "sandbox.h"
#include "stddef.h"

using namespace v8;

void InitBlockchain(Isolate *isolate, Local<ObjectTemplate> globalTpl);
void NewIOSTBlockchain(const FunctionCallbackInfo<Value> &args);

void IOSTBlockchain_transfer(const FunctionCallbackInfo<Value> &args);
void IOSTBlockchain_withdraw(const FunctionCallbackInfo<Value> &args);
void IOSTBlockchain_deposit(const FunctionCallbackInfo<Value> &args);
void IOSTBlockchain_topUp(const FunctionCallbackInfo<Value> &args);
void IOSTBlockchain_countermand(const FunctionCallbackInfo<Value> &args);
void IOSTBlockchain_blockInfo(const FunctionCallbackInfo<Value> &args);
void IOSTBlockchain_txInfo(const FunctionCallbackInfo<Value> &args);
void IOSTBlockchain_call(const FunctionCallbackInfo<Value> &args);
void IOSTBlockchain_callWithReceipt(const FunctionCallbackInfo<Value> &args);
void IOSTBlockchain_requireAuth(const FunctionCallbackInfo<Value> &args);
void IOSTBlockchain_grantServi(const FunctionCallbackInfo<Value> &args);

// This Class wraps Go BlockChain function so JS contract can call them.
class IOSTBlockchain {
private:
    SandboxPtr sbxPtr;
public:
    IOSTBlockchain(SandboxPtr ptr): sbxPtr(ptr) {}

    int Transfer(const char *, const char *, const char *);
    int Withdraw(const char *, const char *);
    int Deposit(const char *, const char *);
    int TopUp(const char *, const char *, const char *);
    int Countermand(const char *, const char *, const char *);
    char *BlockInfo();
    char *TxInfo();
    char *Call(const char *, const char *, const char *);
    char *CallWithReceipt(const char *, const char *, const char *);
    bool RequireAuth(const char *pubKey);
    int GrantServi(const char *, const char *);
};

#endif // IOST_V8_BLOCKCHAIN_H