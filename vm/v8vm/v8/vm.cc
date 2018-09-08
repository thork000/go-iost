#include "vm.h"
#include "v8.h"
#include "sandbox.h"
#include "snapshot_blob.bin.h"
#include "natives_blob.bin.h"

#include "console.h"
#include "storage.h"
#include "blockchain.h"
#include "instruction.h"

#include "libplatform/libplatform.h"

#include <assert.h>
#include <stdlib.h>
#include <stdio.h>

using namespace v8;

intptr_t externalRefx[] = {
        reinterpret_cast<intptr_t>(NewConsoleLog),
        reinterpret_cast<intptr_t>(NewIOSTContractStorage),
        reinterpret_cast<intptr_t>(NewIOSTBlockchain),
        reinterpret_cast<intptr_t>(NewIOSTContractInstruction),
        0};

void init() {
    V8::InitializeICU();

    Platform *platform = platform::CreateDefaultPlatform();
    V8::InitializePlatform(platform);
#ifdef __linux__
    StartupData nativesData, snapshotData;
    nativesData.data = reinterpret_cast<char *>(natives_blob_bin);
    nativesData.raw_size = natives_blob_bin_len;
    snapshotData.data = reinterpret_cast<char *>(snapshot_blob_bin);
    snapshotData.raw_size = snapshot_blob_bin_len;
    V8::SetNativesDataBlob(&nativesData);
    V8::SetSnapshotDataBlob(&snapshotData);
#endif
    V8::Initialize();
    return;
}

IsolatePtr newIsolate(CustomStartupData customStartupData) {
  Isolate::CreateParams params;

  StartupData* blob = new StartupData;
  blob->data = customStartupData.data;
  blob->raw_size = customStartupData.raw_size;

  params.snapshot_blob = blob;
  params.array_buffer_allocator = ArrayBuffer::Allocator::NewDefaultAllocator();
  params.external_references = externalRefx;

  return static_cast<IsolatePtr>(Isolate::New(params));
}

void releaseIsolate(IsolatePtr ptr) {
    if (ptr == nullptr) {
        return;
    }

    Isolate *isolate = static_cast<Isolate*>(ptr);
    isolate->Dispose();
    return;
}

ValueTuple Execute(SandboxPtr ptr, const char *code) {
    ValueTuple ret = Execution(ptr, code);
    return ret;
}
