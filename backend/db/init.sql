CREATE TABLE series (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Plan to Watch',
    last_episode INTEGER NOT NULL DEFAULT 0,
    total_episodes INTEGER NOT NULL DEFAULT 0,
    ranking INTEGER NOT NULL DEFAULT 0
);

-- Insertar datos 
INSERT INTO series (title, status, last_episode, total_episodes, ranking) VALUES
('Haikyu', 'Completed', 75, 75, 10),
('One Piece', 'Plan to Watch', 0, 1000, 9);
('Kamisama-san', 'Watching', 15, 25, 2)