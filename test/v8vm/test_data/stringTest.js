'use strict';
class stringTest {
    constructor() {
    }

    string() {
        let s = "ioststringtest";
        return s;
    }

    valueOf() {
        let s = "ioststringtest";
        return s.valueOf();
    }

    concat() {
        let s1 = "iost";
        let s2 = "stringtest";
        return s1.concat(s2);
    }

    includes() {
        let s = "ioststringtest";
        return s.includes("o");
    }

    endsWith() {
        let s = "ioststringtest";
        return s.endsWith("st");
    }

    indexOf() {
        let s = "ioststringtest";
        return s.indexOf("t");
    }

    lastIndexOf() {
        let s = "ioststringtest";
        return s.lastIndexOf("t");
    }

    replace() {
        let s = "ioststringtest";
        return s.replace("t", "u");
    }

    search() {
        let s = "ioststringtest";
        return s.search("str");
    }

    split() {
        let s = "ioststringtest";
        return s.split("s").toString();
    }

    startsWith() {
        let s = "ioststringtest";
        return s.startsWith("iost");
    }

    slice() {
        let s = "ioststringtest";
        return s.slice(10);
    }

    toLowerCase() {
        let s = "IostStringTest";
        return s.toLowerCase();
    }

    toUpperCase() {
        let s = "ioststringtest";
        return s.toUpperCase();
    }

    trim() {
        let s = "   ioststringtest   ";
        return s.trim();
    }

    trimLeft() {
        let s = "   ioststringtest   ";
        return s.trimLeft();
    }

    trimRight() {
        let s = "   ioststringtest   ";
        return s.trimRight();
    }

    repeat() {
        let s = "ioststringtest";
        return s.repeat(3);
    }
}

module.exports = stringTest;
