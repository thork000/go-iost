#include "compile.h"
#include <cstring>
#include <iostream>

#include "js_bignumber.h"
#include "js_int64.h"
#include "js_utils.h"
#include "js_console.h"
#include "js_storage.h"
#include "js_blockchain.h"

static char injectGasFormat[] =
    "(function(){\n"
    "const source = \"%s\";\n"
    "return injectGas(source);\n"
    "})();";

static char codeFormat[] =
        "let module = {};\n"
        "module.exports = {};\n"
        "%s\n" // load BigNumber
        "let BigNumber = module.exports;\n"
        "%s\n"  // load Int64
        "%s\n"  // load util
        "%s\n"  // load console
        "%s\n";  // load storage

int compile(SandboxPtr ptr, const char *code, const char **compiledCode) {
    Sandbox *sbx = static_cast<Sandbox*>(ptr);
    Isolate *isolate = sbx->isolate;

    Locker locker(isolate);
    Isolate::Scope isolate_scope(isolate);
    HandleScope handle_scope(isolate);

    Local<Context> context = sbx->context.Get(isolate);
    Context::Scope context_scope(context);

    char *injectCode = nullptr;
    asprintf(&injectCode, injectGasFormat, code);

    Local<String> source = String::NewFromUtf8(isolate, injectCode, NewStringType::kNormal).ToLocalChecked();
    free(injectCode);
    Local<String> fileName = String::NewFromUtf8(isolate, "__inject_ga.js", NewStringType::kNormal).ToLocalChecked();
    Local<Script> script = Script::Compile(source, fileName);

    if (!script.IsEmpty()) {
        Local<Value> result = script->Run();
        if (!result.IsEmpty()) {
            String::Utf8Value retStr(result);
            *compiledCode = strdup(*retStr);
            return 0;
        }
    }
    return 1;
}

static inline Local<String> v8_str(const char* x) {
  return String::NewFromUtf8(Isolate::GetCurrent(), x,
                                 NewStringType::kNormal)
      .ToLocalChecked();
}

static inline Local<Script> v8_compile(Local<String> x) {
  Local<Script> result;
  if (Script::Compile(Isolate::GetCurrent()->GetCurrentContext(), x)
          .ToLocal(&result)) {
    return result;
  }
  return Local<v8::Script>();
}

static inline Local<Value> CompileRun(Local<String> source) {
  Local<Value> result;
  if (v8_compile(source)
          ->Run(Isolate::GetCurrent()->GetCurrentContext())
          .ToLocal(&result)) {
    return result;
  }
  return Local<Value>();
}

static inline Local<Value> CompileRun(const char* source) {
  return CompileRun(v8_str(source));
}

CustomStartupData createStartupData() {
    char *bignumberjs = reinterpret_cast<char *>(libjs_bignumber_js);
    char *int64js = reinterpret_cast<char *>(libjs_int64_js);
    char *utilsjs = reinterpret_cast<char *>(libjs_utils_js);
    char *consolejs = reinterpret_cast<char *>(libjs_console_js);
    char *storagejs = reinterpret_cast<char *>(libjs_storage_js);
    char *blockchainjs = reinterpret_cast<char *>(libjs_blockchain_js);

    char *code = nullptr;
    asprintf(&code, codeFormat,
        bignumberjs,
        int64js,
        utilsjs,
        consolejs,
        storagejs);

    StartupData blob;
    {
        SnapshotCreator creator(externalRef);
        Isolate* isolate = creator.GetIsolate();
        {
            HandleScope handle_scope(isolate);

            Local<ObjectTemplate> globalTpl = ObjectTemplate::New(isolate);
            globalTpl->SetInternalFieldCount(1);

            Local<FunctionTemplate> callback = FunctionTemplate::New(isolate, NewConsoleLog);
            globalTpl->Set(v8_str("_cLog"), callback);

            Local<FunctionTemplate> storageClass =
                    FunctionTemplate::New(isolate, NewIOSTContractStorage);
            Local<String> storageClassName = String::NewFromUtf8(isolate, "IOSTStorage");
//            storageClass->SetClassName(storageClassName);

//            Local<ObjectTemplate> storageTpl = storageClass->InstanceTemplate();
//            storageTpl->SetInternalFieldCount(1);
//            storageTpl->Set(
//                String::NewFromUtf8(isolate, "put"),
//                FunctionTemplate::New(isolate, IOSTContractStorage_Put)
//            );
//            storageTpl->Set(
//                String::NewFromUtf8(isolate, "get"),
//                FunctionTemplate::New(isolate, IOSTContractStorage_Get)
//            );
//            storageTpl->Set(
//                    String::NewFromUtf8(isolate, "del"),
//                    FunctionTemplate::New(isolate, IOSTContractStorage_Del)
//            );
//            storageTpl->Set(
//                String::NewFromUtf8(isolate, "globalGet"),
//                FunctionTemplate::New(isolate, IOSTContractStorage_GGet)
//            );
            globalTpl->Set(storageClassName, storageClass);

//            Local<FunctionTemplate> blockchainClass =
//                FunctionTemplate::New(isolate, NewIOSTBlockchain);
//            Local<String> blockchainClassName = String::NewFromUtf8(isolate, "IOSTBlockchain");
//            blockchainClass->SetClassName(blockchainClassName);
//
//            Local<ObjectTemplate> blockchainTpl = blockchainClass->InstanceTemplate();
//            blockchainTpl->SetInternalFieldCount(1);
//            blockchainTpl->Set(
//                String::NewFromUtf8(isolate, "transfer"),
//                FunctionTemplate::New(isolate, IOSTBlockchain_transfer)
//            );
//            blockchainTpl->Set(
//                String::NewFromUtf8(isolate, "withdraw"),
//                FunctionTemplate::New(isolate, IOSTBlockchain_withdraw)
//            );
//            blockchainTpl->Set(
//                String::NewFromUtf8(isolate, "deposit"),
//                FunctionTemplate::New(isolate, IOSTBlockchain_deposit)
//            );
//            blockchainTpl->Set(
//                String::NewFromUtf8(isolate, "topUp"),
//                FunctionTemplate::New(isolate, IOSTBlockchain_topUp)
//            );
//            blockchainTpl->Set(
//                String::NewFromUtf8(isolate, "countermand"),
//                FunctionTemplate::New(isolate, IOSTBlockchain_countermand)
//            );
//            blockchainTpl->Set(
//                String::NewFromUtf8(isolate, "blockInfo"),
//                FunctionTemplate::New(isolate, IOSTBlockchain_blockInfo)
//            );
//            blockchainTpl->Set(
//                String::NewFromUtf8(isolate, "txInfo"),
//                FunctionTemplate::New(isolate, IOSTBlockchain_txInfo)
//            );
//            blockchainTpl->Set(
//                String::NewFromUtf8(isolate, "call"),
//                FunctionTemplate::New(isolate, IOSTBlockchain_call)
//            );
//            blockchainTpl->Set(
//                String::NewFromUtf8(isolate, "callWithReceipt"),
//                FunctionTemplate::New(isolate, IOSTBlockchain_callWithReceipt)
//            );
//            blockchainTpl->Set(
//                String::NewFromUtf8(isolate, "requireAuth"),
//                FunctionTemplate::New(isolate, IOSTBlockchain_requireAuth)
//            );
//            blockchainTpl->Set(
//                String::NewFromUtf8(isolate, "grantServi"),
//                FunctionTemplate::New(isolate, IOSTBlockchain_grantServi)
//            );
//            globalTpl->Set(blockchainClassName, blockchainClass);

//            Local<FunctionTemplate> instructionClass =
//                FunctionTemplate::New(isolate, NewIOSTContractInstruction);
//            Local<String> instructionClassName = String::NewFromUtf8(isolate, "IOSTInstruction");
//            instructionClass->SetClassName(instructionClassName);

//            Local<ObjectTemplate> instructionTpl = instructionClass->InstanceTemplate();
//            instructionTpl->SetInternalFieldCount(1);
//            instructionTpl->Set(
//                String::NewFromUtf8(isolate, "incr"),
//                FunctionTemplate::New(isolate, IOSTContractInstruction_Incr)
//            );
//            instructionTpl->Set(
//                String::NewFromUtf8(isolate, "count"),
//                FunctionTemplate::New(isolate, IOSTContractInstruction_Count)
//            );

//            Local<Value> instructionFunc = instructionClass->GetFunction(context).ToLocalChecked();
//            context->Global()->Set(context, instructionClassName, instructionFunc);
//            globalTpl->Set(instructionClassName, instructionClass);

            Local<Context> context = Context::New(isolate, nullptr, globalTpl);
            Context::Scope context_scope(context);

            CompileRun(code);

            creator.SetDefaultContext(context);
        }
        blob = creator.CreateBlob(SnapshotCreator::FunctionCodeHandling::kClear);
    }

    return CustomStartupData{blob.data, blob.raw_size};
}