class NormalTest {
  init() {
  }

  storageTest() {
    storage.put("key", "value");
    storage.has("key");
    storage.get("key");
    storage.del("key");
    storage.mapPut("mapkey", "field", "value")
    storage.mapHas("mapkey", "field")
    storage.mapGet("mapkey", "field")
    storage.mapKeys("mapkey")
    storage.mapLen("mapkey")
    storage.mapDel("mapkey", "field")
  }

}

module.exports = NormalTest;
