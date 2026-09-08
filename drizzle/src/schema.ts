import {
  bigint,
  boolean,
  date,
  index,
  integer,
  jsonb,
  numeric,
  pgTable,
  primaryKey,
  real,
  text,
  time,
  timestamp,
  uuid,
  uniqueIndex,
} from "drizzle-orm/pg-core";

export const alarms = pgTable(
  "happy_wakey_alarms",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    ownerId: text("owner_id").notNull(),
    label: text("label").notNull(),
    localTime: time("local_time").notNull(),
    timeZone: text("time_zone").notNull(),
    weekdays: jsonb("weekdays").$type<readonly number[]>().notNull(),
    enabled: boolean("enabled").notNull().default(true),
    sound: text("sound").notNull(),
    volume: numeric("volume", { precision: 4, scale: 3 }).notNull(),
    gradualSeconds: integer("gradual_seconds").notNull().default(0),
    tags: jsonb("tags").$type<readonly string[]>().notNull().default([]),
    generation: bigint("generation", { mode: "number" }).notNull().default(0),
    createdAt: timestamp("created_at", { withTimezone: true, mode: "date" })
      .notNull()
      .defaultNow(),
    updatedAt: timestamp("updated_at", { withTimezone: true, mode: "date" })
      .notNull()
      .defaultNow(),
  },
  (table) => [
    index("happy_wakey_alarms_owner_enabled").on(
      table.ownerId,
      table.enabled,
      table.updatedAt.desc(),
    ),
  ],
);

export type AlarmRow = typeof alarms.$inferSelect;
export type NewAlarmRow = typeof alarms.$inferInsert;

export const tenants = pgTable("happy_wakey_tenants", {
  id: text("id").primaryKey(),
  accountKind: text("account_kind").notNull(),
  displayName: text("display_name").notNull(),
  seatLimit: integer("seat_limit").notNull().default(1),
  policyVersion: text("policy_version").notNull(),
  createdAt: timestamp("created_at", { withTimezone: true, mode: "date" }).notNull().defaultNow(),
});

export const tenantMemberships = pgTable(
  "happy_wakey_tenant_memberships",
  {
    tenantId: text("tenant_id").notNull().references(() => tenants.id, { onDelete: "cascade" }),
    subjectId: text("subject_id").notNull(),
    role: text("role").notNull(),
    state: text("state").notNull().default("active"),
    joinedAt: timestamp("joined_at", { withTimezone: true, mode: "date" }).notNull().defaultNow(),
  },
  (table) => [primaryKey({ columns: [table.tenantId, table.subjectId] })],
);

export const connectorConsents = pgTable(
  "happy_wakey_connector_consents",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    tenantId: text("tenant_id").notNull().references(() => tenants.id, { onDelete: "cascade" }),
    subjectId: text("subject_id").notNull(),
    connector: text("connector").notNull(),
    state: text("state").notNull(),
    scopes: jsonb("scopes").$type<readonly string[]>().notNull().default([]),
    sourceAccountRef: text("source_account_ref").notNull(),
    credentialRef: text("credential_ref"),
    grantedAt: timestamp("granted_at", { withTimezone: true, mode: "date" }),
    expiresAt: timestamp("expires_at", { withTimezone: true, mode: "date" }),
    revokedAt: timestamp("revoked_at", { withTimezone: true, mode: "date" }),
    policyVersion: text("policy_version").notNull(),
    updatedAt: timestamp("updated_at", { withTimezone: true, mode: "date" }).notNull().defaultNow(),
  },
  (table) => [
    uniqueIndex("happy_wakey_connector_consent_identity").on(
      table.tenantId,
      table.subjectId,
      table.connector,
      table.sourceAccountRef,
    ),
  ],
);

export const sourceItemCandidates = pgTable(
  "happy_wakey_source_item_candidates",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    tenantId: text("tenant_id").notNull().references(() => tenants.id, { onDelete: "cascade" }),
    subjectId: text("subject_id").notNull(),
    consentId: uuid("consent_id").notNull().references(() => connectorConsents.id, { onDelete: "cascade" }),
    provider: text("provider").notNull(),
    sourceItemRef: text("source_item_ref").notNull(),
    senderClass: text("sender_class").notNull(),
    receivedAt: timestamp("received_at", { withTimezone: true, mode: "date" }).notNull(),
    threadRef: text("thread_ref").notNull(),
    encryptedContentRef: text("encrypted_content_ref").notNull(),
    contentSha256: text("content_sha256").notNull(),
    hasDirectReplyRequest: boolean("has_direct_reply_request").notNull().default(false),
    dueAt: timestamp("due_at", { withTimezone: true, mode: "date" }),
    retentionExpiresAt: timestamp("retention_expires_at", { withTimezone: true, mode: "date" }).notNull(),
    createdAt: timestamp("created_at", { withTimezone: true, mode: "date" }).notNull().defaultNow(),
  },
  (table) => [
    uniqueIndex("happy_wakey_candidate_source_identity").on(
      table.tenantId,
      table.subjectId,
      table.provider,
      table.sourceItemRef,
    ),
    index("happy_wakey_candidates_pending").on(table.tenantId, table.subjectId, table.receivedAt.desc()),
  ],
);

export const usefulnessDecisions = pgTable(
  "happy_wakey_usefulness_decisions",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    tenantId: text("tenant_id").notNull().references(() => tenants.id, { onDelete: "cascade" }),
    subjectId: text("subject_id").notNull(),
    sourceItemId: uuid("source_item_id").notNull().references(() => sourceItemCandidates.id, { onDelete: "cascade" }),
    model: text("model").notNull(),
    policyVersion: text("policy_version").notNull(),
    contentSha256: text("content_sha256").notNull(),
    score: numeric("score", { precision: 5, scale: 4 }).notNull(),
    designatedUseful: boolean("designated_useful").notNull(),
    reasons: jsonb("reasons").$type<readonly string[]>().notNull().default([]),
    rationale: text("rationale").notNull(),
    idempotencyKey: text("idempotency_key").notNull(),
    evaluatedAt: timestamp("evaluated_at", { withTimezone: true, mode: "date" }).notNull().defaultNow(),
  },
  (table) => [uniqueIndex("happy_wakey_usefulness_idempotency").on(table.tenantId, table.subjectId, table.idempotencyKey)],
);

export const safeDeepLinks = pgTable(
  "happy_wakey_safe_deep_links",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    tenantId: text("tenant_id").notNull().references(() => tenants.id, { onDelete: "cascade" }),
    subjectId: text("subject_id").notNull(),
    sourceItemId: uuid("source_item_id").notNull().references(() => sourceItemCandidates.id, { onDelete: "cascade" }),
    decisionId: uuid("decision_id").notNull().references(() => usefulnessDecisions.id, { onDelete: "cascade" }),
    provider: text("provider").notNull(),
    targetUrl: text("target_url").notNull(),
    expiresAt: timestamp("expires_at", { withTimezone: true, mode: "date" }).notNull(),
    requiresReauthentication: boolean("requires_reauthentication").notNull().default(true),
    feedFallbackAllowed: boolean("feed_fallback_allowed").notNull().default(false),
    createdAt: timestamp("created_at", { withTimezone: true, mode: "date" }).notNull().defaultNow(),
  },
  (table) => [
    uniqueIndex("happy_wakey_safe_link_identity").on(
      table.tenantId,
      table.subjectId,
      table.sourceItemId,
      table.decisionId,
    ),
    index("happy_wakey_safe_links_active").on(table.tenantId, table.subjectId, table.expiresAt),
  ],
);

export const morningBriefings = pgTable(
  "happy_wakey_morning_briefings",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    tenantId: text("tenant_id").notNull().references(() => tenants.id, { onDelete: "cascade" }),
    subjectId: text("subject_id").notNull(),
    localDate: date("local_date", { mode: "string" }).notNull(),
    timeZone: text("time_zone").notNull(),
    title: text("title").notNull(),
    cards: jsonb("cards").$type<readonly unknown[]>().notNull().default([]),
    cardCount: integer("card_count").notNull().default(0),
    suppressedItemCount: integer("suppressed_item_count").notNull().default(0),
    audioUrl: text("audio_url"),
    audioDurationSeconds: integer("audio_duration_seconds"),
    generatedAt: timestamp("generated_at", { withTimezone: true, mode: "date" }).notNull().defaultNow(),
    validUntil: timestamp("valid_until", { withTimezone: true, mode: "date" }).notNull(),
  },
  (table) => [uniqueIndex("happy_wakey_briefing_subject_day").on(table.tenantId, table.subjectId, table.localDate)],
);

export const embeddings = pgTable(
  "happy_wakey_embeddings",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    tenantId: text("tenant_id").notNull().references(() => tenants.id, { onDelete: "cascade" }),
    subjectId: text("subject_id").notNull(),
    resourceType: text("resource_type").notNull(),
    resourceId: text("resource_id").notNull(),
    model: text("model").notNull(),
    contentHash: text("content_hash").notNull(),
    dimensions: integer("dimensions").notNull(),
    embedding: real("embedding").array().notNull(),
    retentionClass: text("retention_class").notNull().default("ephemeral"),
    idempotencyKey: text("idempotency_key").notNull(),
    createdAt: timestamp("created_at", { withTimezone: true, mode: "date" }).notNull().defaultNow(),
    updatedAt: timestamp("updated_at", { withTimezone: true, mode: "date" }).notNull().defaultNow(),
  },
  (table) => [
    uniqueIndex("happy_wakey_embedding_idempotency").on(table.tenantId, table.subjectId, table.idempotencyKey),
    index("happy_wakey_embeddings_resource").on(
      table.tenantId,
      table.subjectId,
      table.resourceType,
      table.resourceId,
      table.updatedAt.desc(),
    ),
  ],
);

export const correlationFindings = pgTable("happy_wakey_correlation_findings", {
  id: uuid("id").primaryKey().defaultRandom(),
  tenantId: text("tenant_id").notNull().references(() => tenants.id, { onDelete: "cascade" }),
  subjectId: text("subject_id").notNull(),
  metricA: text("metric_a").notNull(),
  metricB: text("metric_b").notNull(),
  coefficient: numeric("coefficient", { precision: 8, scale: 7 }).notNull(),
  pValue: numeric("p_value", { precision: 8, scale: 7 }).notNull(),
  confidenceLow: numeric("confidence_low", { precision: 8, scale: 7 }).notNull(),
  confidenceHigh: numeric("confidence_high", { precision: 8, scale: 7 }).notNull(),
  sampleSize: integer("sample_size").notNull(),
  correction: text("correction").notNull(),
  causalClaimAllowed: boolean("causal_claim_allowed").notNull().default(false),
  discoveredAt: timestamp("discovered_at", { withTimezone: true, mode: "date" }).notNull().defaultNow(),
});

export const chatSessions = pgTable("happy_wakey_chat_sessions", {
  id: uuid("id").primaryKey().defaultRandom(),
  tenantId: text("tenant_id").notNull().references(() => tenants.id, { onDelete: "cascade" }),
  subjectId: text("subject_id").notNull(),
  audience: text("audience").notNull(),
  allowedSearchScopes: jsonb("allowed_search_scopes").$type<readonly string[]>().notNull().default([]),
  oresChatSessionRef: text("ores_chat_session_ref").notNull(),
  openedAt: timestamp("opened_at", { withTimezone: true, mode: "date" }).notNull().defaultNow(),
  expiresAt: timestamp("expires_at", { withTimezone: true, mode: "date" }).notNull(),
  closedAt: timestamp("closed_at", { withTimezone: true, mode: "date" }),
});

export type MorningBriefingRow = typeof morningBriefings.$inferSelect;
export type SafeDeepLinkRow = typeof safeDeepLinks.$inferSelect;
export type UsefulnessDecisionRow = typeof usefulnessDecisions.$inferSelect;
