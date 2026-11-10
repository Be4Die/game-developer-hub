-- Тестовые учётные записи для ЛОКАЛЬНОЙ разработки (пароль: password123).
-- Не входит в схему и НЕ должно применяться в продуктивной среде.
-- Применение: task stack:db:seed-dev-users
INSERT INTO users (email, password_hash, display_name, role, status, email_verified)
VALUES
    ('dev@welwise.com', decode('243261243130246134693851654a7948306544747170573176594c73753370472e46334c6c674b5271706e753956593334325045416b346159397257', 'hex'), 'Test Developer', 1, 1, TRUE),
    ('moderator@welwise.com', decode('243261243130246134693851654a7948306544747170573176594c73753370472e46334c6c674b5271706e753956593334325045416b346159397257', 'hex'), 'Test Moderator', 2, 1, TRUE)
ON CONFLICT (email) DO NOTHING;
