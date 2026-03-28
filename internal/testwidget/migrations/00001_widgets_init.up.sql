-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "widgets" (
  "id" TEXT NOT NULL PRIMARY KEY,
  "title" TEXT,
  "version" BIGINT,
  "deleted" BOOLEAN
);
ALTER TABLE "widgets" REPLICA IDENTITY FULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Additive policy: no DROP TABLE in default down (manual rollback if needed).
SELECT 1;
-- +goose StatementEnd
