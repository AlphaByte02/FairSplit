-- Postgres schema for SplitFlow
CREATE TABLE
  users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
  );

CREATE TABLE
  sessions (
    id UUID PRIMARY KEY,
    created_by_id UUID NOT NULL REFERENCES users (id),
    name TEXT NOT NULL,
    is_closed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
  );

CREATE INDEX idx_session_created_by_id ON sessions (created_by_id);

CREATE TABLE
  session_participants (
    session_id UUID NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (session_id, user_id)
  );

CREATE INDEX idx_session_participants_session_id ON session_participants (session_id);

CREATE INDEX idx_session_participants_user_id ON session_participants (user_id);

CREATE TABLE
  payments (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    payer_id UUID NOT NULL REFERENCES users (id),
    amount NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
    description text,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
  );

CREATE INDEX idx_payments_session_id ON payments (session_id);

CREATE INDEX idx_payments_payer_id ON payments (payer_id);

CREATE TABLE
  payment_participants (
    payment_id UUID NOT NULL REFERENCES payments (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (payment_id, user_id)
  );

CREATE INDEX idx_payment_participants_payment_id ON payment_participants (payment_id);

CREATE INDEX idx_payment_participants_user_id ON payment_participants (user_id);
