import { asc, eq } from "drizzle-orm";
import { alarms } from "./schema.js";

const MAX_SUBJECT_BYTES = 512;

/**
 * Return only the predicates and ordering for a subject-scoped alarm read.
 * The web/API adapter owns the concrete database connection and can receive a
 * subject only after Shared Auth has established it.
 */
export function alarmsForSubject(subject: string) {
  validateSubject(subject);
  return {
    where: eq(alarms.ownerId, subject),
    orderBy: asc(alarms.createdAt),
  } as const;
}

export function validateSubject(subject: string): void {
  if (
    subject.length === 0 ||
    Buffer.byteLength(subject, "utf8") > MAX_SUBJECT_BYTES ||
    /\s/u.test(subject)
  ) {
    throw new TypeError(
      "verified subject must be non-empty, bounded, and whitespace-free",
    );
  }
}
