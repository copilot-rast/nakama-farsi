-- Planet-scoped Jazebe groups use Nakama user IDs. Names remain reserved after disbanding.
-- +migrate Up
CREATE TABLE jazebe_group (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    name VARCHAR(64) NOT NULL,
    description VARCHAR(280),
    icon VARCHAR(32),
    color VARCHAR(9),
    mode VARCHAR(8) NOT NULL CHECK (mode IN ('open','approval')),
    member_cap INTEGER NOT NULL CHECK (member_cap >= 1),
    member_count INTEGER NOT NULL DEFAULT 1 CHECK (member_count >= 0),
    creator_user_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(12) NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended','disbanded')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    disbanded_at TIMESTAMPTZ,
    CONSTRAINT jazebe_group_color CHECK (color IS NULL OR color ~ '^#[0-9a-fA-F]{6}$'),
    CONSTRAINT jazebe_group_members_bound CHECK (member_count <= member_cap)
);
CREATE UNIQUE INDEX uq_jazebe_group_planet_name ON jazebe_group(planet_id,lower(name));
CREATE INDEX idx_jazebe_group_planet_status_members ON jazebe_group(planet_id,status,member_count DESC);
CREATE TABLE jazebe_group_membership (
    id UUID PRIMARY KEY,
    group_id UUID NOT NULL REFERENCES jazebe_group(id),
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    user_id UUID NOT NULL REFERENCES users(id),
    role VARCHAR(8) NOT NULL CHECK (role IN ('creator','admin','member')),
    status VARCHAR(8) NOT NULL DEFAULT 'active' CHECK (status IN ('active','left','removed')),
    source VARCHAR(20) NOT NULL CHECK (source IN ('creation','list_join','invite_link','approved_request')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    left_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX uq_jazebe_group_active_user ON jazebe_group_membership(planet_id,user_id) WHERE status='active';
CREATE UNIQUE INDEX uq_jazebe_group_creator ON jazebe_group_membership(group_id) WHERE role='creator' AND status='active';
CREATE INDEX idx_jazebe_group_member_group ON jazebe_group_membership(group_id,status,role);
CREATE INDEX idx_jazebe_group_member_user ON jazebe_group_membership(user_id,status);
CREATE TABLE jazebe_group_activity (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    group_id UUID NOT NULL REFERENCES jazebe_group(id),
    kind VARCHAR(12) NOT NULL CHECK (kind IN ('join','leave','ticket','game','prize','mission','role')),
    actor_user_id UUID REFERENCES users(id),
    message_key VARCHAR(64) NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_jazebe_group_activity_group_created ON jazebe_group_activity(group_id,created_at DESC);
CREATE TABLE jazebe_group_join_request (
    id UUID PRIMARY KEY,
    planet_id UUID NOT NULL REFERENCES jazebe_planet(id),
    group_id UUID NOT NULL REFERENCES jazebe_group(id),
    user_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected','cancelled','auto_cancelled')),
    requested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    decided_by_user_id UUID REFERENCES users(id),
    decided_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX uq_jazebe_group_request_pending ON jazebe_group_join_request(planet_id,user_id) WHERE status='pending';
CREATE INDEX idx_jazebe_group_request_group ON jazebe_group_join_request(group_id,status);

-- +migrate Down
-- Group membership and contribution history may represent monetary ownership.
SELECT 1;
