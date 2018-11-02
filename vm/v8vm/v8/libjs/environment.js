'use strict';

// String
const stringAllowedMethods = [
    'charCodeAt',
    'length',
    'constructor',
    'toString',
    'valueOf',
    'concat',
    'includes',
    'endsWith',
    'indexOf',
    'lastIndexOf',
    'replace',
    'search',
    'split',
    'startsWith',
    'slice',
    'substring',
    'toLowerCase',
    'toUpperCase',
    'trim',
    'trimLeft',
    'trimRight',
    'repeat'
];

const stringMethods = Object.getOwnPropertyNames(String.prototype);
stringMethods.forEach((method) => {
    if (!stringAllowedMethods.includes(method)) {
        String.prototype[method] = null;
        return;
    }

    const origMethod = String.prototype[method];
    String.prototype[method] = function() {
        // console.log("String: " + method)
        return origMethod.call(this, ...arguments);
    }
});

// Array
const arrayAllowedMethods = [
    'constructor',
    'toString',
    'concat',
    'every',
    'filter',
    'find',
    'findIndex',
    'forEach',
    'includes',
    'indexOf',
    'join',
    'keys',
    'lastIndexOf',
    'map',
    'pop',
    'push',
    'reverse',
    'shift',
    'slice',
    'sort',
    'splice',
    'unshift'
];

const arrayMethods = Object.getOwnPropertyNames(Array.prototype);
arrayMethods.forEach((method) => {
    if (!arrayAllowedMethods.includes(method)) {
        Array.prototype[method] = null;
        return;
    }

    if (method === 'length') {
        return;
    }

    const origMethod = Array.prototype[method];
    Array.prototype[method] = function() {
        // console.log("Array: ", method)
        return origMethod.call(this, ...arguments);
    }
});

// JSON
const JSONMethods = Object.getOwnPropertyNames(JSON);
JSONMethods.forEach((method) => {
    const origMethod = JSON[method];
    JSON[method] = function() {
        // console.log("JSON: " + method)
        return origMethod.call(this, ...arguments);
    }
});

// Functions
parseFloat = null;
parseInt = null;
decodeURI = null;
decodeURIComponent = null;
encodeURI = null;
encodeURIComponent = null;
escape = null;
unescape = null;

// Fundamental Objects
Function = null;
Boolean = null;
EvalError = null;
RangeError = null;
ReferenceError = null;
SyntaxError = null;
TypeError = null;
URIError = null;

// Numbers and dates
Number = null;
Math = null;
Date = null;

// Text processing
RegExp = null;

// Indexed collections
Int8Array = null;
Uint8Array = null;
Uint8ClampedArray = null;
Int16Array = null;
Uint16Array = null;
Int32Array = null;
Uint32Array = null;
Float32Array = null;
Float64Array = null;

// Keyed collections
Map = null;
Set = null;
WeakMap = null;
WeakSet = null;

// Structured data
ArrayBuffer = null;
SharedArrayBuffer = null;
Atomics = null;
DataView = null;

// Control abstraction objects
Promise = null;
Generator = null;
GeneratorFunction = null;
AsyncFunction = null;

// ReflectionSection
Reflect = null;
Proxy = null;

// InternationalizationSection
Intl = null;

// WebAssembly
WebAssembly = null;