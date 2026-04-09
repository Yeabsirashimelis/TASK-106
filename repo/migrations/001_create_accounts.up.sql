CREATE TYPE account_role AS ENUM (
    'Administrator',
    'Scheduler',
    'Instructor',
    'Reviewer',
    'Finance Clerk',
    'Auditor'
);

CREATE TYPE account_status AS ENUM (
    'Active',
    'Frozen',
    'Deactivated'
);

CREATE TABLE accounts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          account_role NOT NULL,
    status        account_status NOT NULL DEFAULT 'Active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_username ON accounts (username);
CREATE INDEX idx_accounts_status ON accounts (status);
