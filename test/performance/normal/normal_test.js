class NormalTest {
    init() {
    }

    storageTest() {
        storage.put("key", "value");
        storage.has("key");
        storage.get("key");
        storage.del("key");
        storage.mapPut("mapkey", "field", "value");
        storage.mapHas("mapkey", "field");
        storage.mapGet("mapkey", "field");
        storage.mapKeys("mapkey");
        storage.mapLen("mapkey");
        storage.mapDel("mapkey", "field");
    }

    blockchainTest() {
        block.number;
        block.time;
        tx.time;
        tx.hash;
        const blockInfo = JSON.parse(BlockChain.blockInfo());
        const txInfo = JSON.parse(BlockChain.txInfo());
        // BlockChain.topUp();
        const contextInfo = JSON.parse(BlockChain.contextInfo());
        const cn = BlockChain.contractName();
        const p = BlockChain.publisher();
    }

    varTest() {
        let a = new Float64("12345.6789");
        let b = a.minus(100);
        const c = a.multi(12);
        const d = a.div(12);
        a.mod(30);
        const e = a.pow(2);
        a.eq(10);
        a.gt(10);
        a.gte(10);
        a.lt(10);
        a.lte(10);
        a.negated(10)
        a.isZero()
        a.isPositive()
        a.isNegative()
        a.toString()
        a.toFixed(10)
        const number = new Int64("1234500000");
        const number2 = number.minus(680);
        const number3 = number.multi(12);
        const number4 = number.div(12);
        const number5 = number.mod(12);
        const number6 = number.pow(2);
    }


    transfer(from, to, amount) {
        BlockChain.transfer(from, to, amount, "")
        BlockChain.deposit(from, amount, "")
        BlockChain.withdraw(to, amount, "")
    }

    test() {
        this.storageTest();
        this.blockchainTest();
        this.varTest();
    }
}

module.exports = NormalTest;