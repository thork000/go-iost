#include "compile.h"
#include <cstring>

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
        "%s\n" // load Int64
        "%s\n" // load util
        "%s\n" // load console
        "%s\n";// load storage

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
////    char *utilsjs = reinterpret_cast<char *>(utils_js);
////    char *consolejs = reinterpret_cast<char *>(console_js);
////    char *storagejs = reinterpret_cast<char *>(storage_js);
//
    char *code = nullptr;
    asprintf(&code, codeFormat, bignumberjs, int64js);
//
    StartupData blob;
    {
        SnapshotCreator creator;
        Isolate* isolate = creator.GetIsolate();
        {
            HandleScope handle_scope(isolate);
            Local<Context> context = Context::New(isolate);
            Context::Scope context_scope(context);

            CompileRun(code);
            creator.SetDefaultContext(context);
        }
////        blob = creator.CreateBlob(SnapshotCreator::FunctionCodeHandling::kClear);
    }
//
//    return CustomStartupData{blob.data, blob.raw_size};
    return CustomStartupData{};
}