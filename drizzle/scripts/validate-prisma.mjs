import { spawnSync } from "node:child_process";

const executable = process.platform === "win32" ? "prisma.cmd" : "prisma";
const result = spawnSync(
  executable,
  ["validate", "--schema", "../prisma/schema.prisma"],
  {
    cwd: new URL("..", import.meta.url),
    env: {
      ...process.env,
      DATABASE_URL:
        process.env.DATABASE_URL ?? "postgresql://placeholder.invalid/happy_wakey",
    },
    stdio: "inherit",
  },
);

if (result.error) throw result.error;
process.exitCode = result.status ?? 1;
