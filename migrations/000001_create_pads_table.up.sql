CREATE TABLE pads (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    created TIMESTAMPTZ NOT NULL,
    expires TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_pads_created ON pads(created);
