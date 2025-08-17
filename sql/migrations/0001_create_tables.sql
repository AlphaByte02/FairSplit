-- Postgres schema for FairSplit
CREATE TABLE users (
  id UUID PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);


CREATE TABLE sessions (
  id UUID PRIMARY KEY,
  created_by_id UUID NOT NULL REFERENCES users (id),
  name TEXT NOT NULL,
  is_closed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  UNIQUE (created_by_id, name)
);


CREATE INDEX idx_session_created_by_id ON sessions (created_by_id);


CREATE TABLE session_participants (
  session_id UUID NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  PRIMARY KEY (session_id, user_id)
);


CREATE INDEX idx_session_participants_session_id ON session_participants (session_id);


CREATE INDEX idx_session_participants_user_id ON session_participants (user_id);


CREATE TABLE transactions (
  id UUID PRIMARY KEY,
  session_id UUID NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
  payer_id UUID NOT NULL REFERENCES users (id),
  amount NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
  description text,
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);


CREATE INDEX idx_transactions_session_id ON transactions (session_id);


CREATE INDEX idx_transactions_payer_id ON transactions (payer_id);


CREATE TABLE transaction_participants (
  transaction_id UUID NOT NULL REFERENCES transactions (id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  PRIMARY KEY (transaction_id, user_id)
);


CREATE INDEX idx_transaction_participants_transaction_id ON transaction_participants (transaction_id);


CREATE INDEX idx_transaction_participants_user_id ON transaction_participants (user_id);
