'use strict';

class For {
    constructor() {
    }

    doIn(num) {
        for (let i in Array.apply(null, { length: num })) {
        }
    }
};

module.exports = For;
