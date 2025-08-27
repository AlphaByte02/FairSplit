-- Postgres schema for FairSplit
CREATE OR REPLACE FUNCTION update_modified_column () RETURNS TRIGGER AS $$
BEGIN
NEW.updated_at = now();
RETURN NEW;
END;
$$ language 'plpgsql';


CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    username TEXT UNIQUE,
    picture TEXT,
    paypal_username TEXT,
    iban TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TRIGGER update_users_updated_at BEFORE
UPDATE ON users FOR EACH ROW
EXECUTE PROCEDURE update_modified_column ();


CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    created_by_id UUID NOT NULL REFERENCES users (id),
    name TEXT NOT NULL,
    is_closed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (created_by_id, name)
);


CREATE TRIGGER update_sessions_updated_at BEFORE
UPDATE ON sessions FOR EACH ROW
EXECUTE PROCEDURE update_modified_column ();


CREATE INDEX idx_session_created_by_id ON sessions (created_by_id);


CREATE TABLE session_participants (
    session_id UUID NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, user_id)
);


CREATE INDEX idx_session_participants_session_id ON session_participants (session_id);


CREATE INDEX idx_session_participants_user_id ON session_participants (user_id);


CREATE TABLE final_balances (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    creditor_id UUID NOT NULL REFERENCES users (id),
    debtor_id UUID NOT NULL REFERENCES users (id),
    amount NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
    is_paid BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TRIGGER update_final_balances_updated_at BEFORE
UPDATE ON final_balances FOR EACH ROW
EXECUTE PROCEDURE update_modified_column ();


CREATE INDEX idx_final_balances_session_id ON final_balances (session_id);


CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    payer_id UUID NOT NULL REFERENCES users (id),
    amount NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
    description TEXT,
    created_by_id UUID NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TRIGGER update_transactions_updated_at BEFORE
UPDATE ON transactions FOR EACH ROW
EXECUTE PROCEDURE update_modified_column ();


CREATE INDEX idx_transactions_session_id ON transactions (session_id);


CREATE INDEX idx_transactions_payer_id ON transactions (payer_id);


CREATE TABLE transaction_participants (
    transaction_id UUID NOT NULL REFERENCES transactions (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id),
    PRIMARY KEY (transaction_id, user_id)
);


CREATE INDEX idx_transaction_participants_transaction_id ON transaction_participants (transaction_id);


CREATE INDEX idx_transaction_participants_user_id ON transaction_participants (user_id);
