-- Проверка и добавление столбца user_id в таблицу crash_games
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'crash_games' AND column_name = 'user_id'
    ) THEN
        ALTER TABLE crash_games ADD COLUMN user_id BIGINT NOT NULL DEFAULT 0;
    END IF;
END $$; 