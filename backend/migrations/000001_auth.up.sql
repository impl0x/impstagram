CREATE TYPE account_status AS ENUM ('unverified','verified','banned');

CREATE TYPE auth_channel AS ENUM ('email','phone','username','totp');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    display_name TEXT NOT NULL DEFAULT '',
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL DEFAULT '',
    phone TEXT UNIQUE NOT NULL DEFAULT '',
    totp_secret_key TEXT UNIQUE NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL,
    dob DATE NOT NULL,
    status account_status NOT NULL DEFAULT 'unverified',
    two_fas auth_channel[], -- can be null, if so then 2fa is disabled 
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jwt_id UUID NOT NULL,
    token_hash TEXT NOT NULL,
    user_id UUID NOT NULL,
    ip_address INET,
    os_name TEXT,
    browser_name TEXT,
    device_type TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);