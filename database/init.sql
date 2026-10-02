-- Инициализация структуры базы данных WWchan

CREATE TABLE IF NOT EXISTS ref_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ref_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL
);

-- Вставляем начальные категории (типы постов)
INSERT INTO ref_categories (name, slug) VALUES 
('Общение', 'general'),
('Игры', 'games'),
('Музыка', 'music'),
('Творчество', 'art'),
('Технологии', 'tech'),
('Мемы', 'memes'),
('Кино и сериалы', 'movies'),
('Оффтоп', 'offtop')
ON CONFLICT (slug) DO NOTHING;

CREATE TABLE IF NOT EXISTS op_posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES ref_users(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES ref_categories(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    image_url VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS op_comments (
    id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES op_posts(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES ref_users(id) ON DELETE CASCADE,
    parent_comment_id INTEGER REFERENCES op_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    image_url VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS op_reactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES ref_users(id) ON DELETE CASCADE,
    post_id INTEGER REFERENCES op_posts(id) ON DELETE CASCADE,
    comment_id INTEGER REFERENCES op_comments(id) ON DELETE CASCADE,
    reaction_value SMALLINT NOT NULL CHECK (reaction_value IN (1, -1)),
    -- Убедимся, что реакция ставится либо на пост, либо на коммент, но не на оба сразу
    CHECK ((post_id IS NOT NULL AND comment_id IS NULL) OR (post_id IS NULL AND comment_id IS NOT NULL)),
    -- Один юзер может поставить только одну реакцию на пост или коммент
    UNIQUE(user_id, post_id),
    UNIQUE(user_id, comment_id)
);

-- ХРАНИМЫЕ ФУНКЦИИ --
-- Требование: "На стороне базы данных для облегчения функционала бэкенда создаются хранимые функции."

-- 1. Создание поста
CREATE OR REPLACE FUNCTION func_create_post(
    p_user_id INT,
    p_category_id INT,
    p_title VARCHAR,
    p_content TEXT,
    p_image_url VARCHAR
) RETURNS INT AS $$
DECLARE
    new_post_id INT;
BEGIN
    INSERT INTO op_posts (user_id, category_id, title, content, image_url)
    VALUES (p_user_id, p_category_id, p_title, p_content, p_image_url)
    RETURNING id INTO new_post_id;
    
    RETURN new_post_id;
END;
$$ LANGUAGE plpgsql;

-- 2. Создание комментария
CREATE OR REPLACE FUNCTION func_create_comment(
    p_post_id INT,
    p_user_id INT,
    p_parent_comment_id INT,
    p_content TEXT,
    p_image_url VARCHAR
) RETURNS INT AS $$
DECLARE
    new_comment_id INT;
BEGIN
    INSERT INTO op_comments (post_id, user_id, parent_comment_id, content, image_url)
    VALUES (p_post_id, p_user_id, p_parent_comment_id, p_content, p_image_url)
    RETURNING id INTO new_comment_id;
    
    RETURN new_comment_id;
END;
$$ LANGUAGE plpgsql;

-- 3. Установка реакции (upsert)
CREATE OR REPLACE FUNCTION func_set_reaction(
    p_user_id INT,
    p_post_id INT,
    p_comment_id INT,
    p_value SMALLINT
) RETURNS VOID AS $$
BEGIN
    IF p_post_id IS NOT NULL THEN
        INSERT INTO op_reactions (user_id, post_id, reaction_value)
        VALUES (p_user_id, p_post_id, p_value)
        ON CONFLICT (user_id, post_id) DO UPDATE 
        SET reaction_value = EXCLUDED.reaction_value;
    ELSIF p_comment_id IS NOT NULL THEN
        INSERT INTO op_reactions (user_id, comment_id, reaction_value)
        VALUES (p_user_id, p_comment_id, p_value)
        ON CONFLICT (user_id, comment_id) DO UPDATE 
        SET reaction_value = EXCLUDED.reaction_value;
    END IF;
END;
$$ LANGUAGE plpgsql;
