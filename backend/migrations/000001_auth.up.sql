CREATE TYPE account_status AS ENUM ('unverified','verified','banned');

CREATE TYPE auth_channel AS ENUM ('email','phone','username','totp');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE,
    phone TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    dob DATE NOT NULL,
    status account_status NOT NULL DEFAULT 'unverified',
    totp_secret_key TEXT,
    two_fas auth_channel[], -- can be null, if so then 2fa is disabled
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jwt_id UUID NOT NULL UNIQUE,
    token_hash TEXT NOT NULL,
    user_id UUID NOT NULL,
    ip_address INET,
    os_name TEXT,
    browser_name TEXT,
    device_type TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE profiles(
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    username TEXT UNIQUE NOT NULL,
    display_name TEXT,
    avatar_url TEXT,
    is_private BOOLEAN NOT NULL DEFAULT false,
    bio TEXT,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
