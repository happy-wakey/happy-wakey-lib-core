import assert from "node:assert/strict";
import test from "node:test";
import { alarmsForSubject, validateSubject } from "./repository.js";

test("subject validation fails closed", () => {
  assert.throws(() => validateSubject(""));
  assert.throws(() => validateSubject("subject with spaces"));
  assert.throws(() => validateSubject("x".repeat(513)));
  assert.doesNotThrow(() => validateSubject("customer:01J00000000000000000000000"));
});

test("alarm reads are built from the internal owner column", () => {
  const query = alarmsForSubject("customer:01J00000000000000000000000");
  assert.ok(query.where);
  assert.ok(query.orderBy);
  assert.equal("ownerId" in query, false);
});
