CREATE TABLE books (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    isbn VARCHAR(20) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    author_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT books_stock_non_negative
        CHECK (stock >= 0),

    CONSTRAINT books_author_fk
        FOREIGN KEY (author_id)
        REFERENCES authors(id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,

    CONSTRAINT books_category_fk
        FOREIGN KEY (category_id)
        REFERENCES categories(id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);