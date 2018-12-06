'use strict';
class stringTest {
    constructor() {
    }

    string() {
        var s = "ioststringtest";
        return s;
    }

    valueOf() {
        var s = "ioststringtest";
        return s.valueOf();
    }

    concat() {
        var s1 = "iost";
        var s2 = "stringtest";
        return s1.concat(s2);
    }

    includes() {
        var s = "ioststringtest";
        return s.includes("o");
    }

    endsWith() {
        var s = "ioststringtest";
        return s.endsWith("st");
    }

    indexOf() {
        var s = "ioststringtest";
        return s.indexOf("t");
    }

    lastIndexOf() {
        var s = "ioststringtest";
        return s.lastIndexOf("t");
    }

    replace() {
        var s = "ioststringtest";
        return s.replace("t", "u");
    }

    search() {
        var s = "ioststringtest";
        return s.search("str");
    }

    split() {
        var s = "ioststringtest";
        return s.split("s").toString();
    }

    startsWith() {
        var s = "ioststringtest";
        return s.startsWith("iost");
    }

    slice() {
        var s = "ioststringtest";
        return s.slice(10);
    }

    toLowerCase() {
        var s = "IostStringTest";
        return s.toLowerCase();
    }

    toUpperCase() {
        var s = "ioststringtest";
        return s.toUpperCase();
    }

    trim() {
        var s = "   ioststringtest   ";
        return s.trim();
    }

    trimLeft() {
        var s = "   ioststringtest   ";
        return s.trimLeft();
    }

    trimRight() {
        var s = "   ioststringtest   ";
        return s.trimRight();
    }

    repeat() {
        var s = "ioststringtest";
        return s.repeat(3);
    }
}

module.exports = stringTest;
