import {
  bigint,
  boolean,
  index,
  integer,
  jsonb,
  numeric,
  pgTable,
  text,
  time,
  timestamp,
  uuid,
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
