create table groups (
    id text primary key,
    name varchar(64) not null,
    created_at timestamptz not null,
    updated_at timestamptz not null,
    deleted_at timestamptz
);

create table resources (
    id text primary key,
    group_id text not null references groups(id),
    name text not null,
    type text not null,
    status text not null,
    failed_stage text,
    error_message text,
    created_at timestamptz not null,
    updated_at timestamptz not null,
    deleted_at timestamptz
);

create table graphs (
    id text primary key,
    group_id text not null references groups(id),
    resource_id text references resources(id),
    graph_type text not null,
    version integer not null,
    status text not null,
    title text,
    summary text,
    is_active boolean not null,
    created_at timestamptz not null,
    updated_at timestamptz not null,
    archived_at timestamptz
);

create table graph_nodes (
    id text primary key,
    graph_id text not null references graphs(id),
    parent_node_id text references graph_nodes(id),
    name text not null,
    description text,
    meaning text,
    node_type text not null,
    source_type text not null,
    level integer not null,
    is_expansion boolean not null,
    metadata jsonb not null,
    created_at timestamptz not null,
    updated_at timestamptz not null
);

create table graph_edges (
    id text primary key,
    graph_id text not null references graphs(id),
    from_node_id text not null references graph_nodes(id),
    to_node_id text not null references graph_nodes(id),
    relation_type text not null,
    description text,
    is_expansion boolean not null,
    metadata jsonb not null,
    created_at timestamptz not null
);

create table processing_jobs (
    id text primary key,
    group_id text not null references groups(id),
    resource_id text references resources(id),
    job_type text not null,
    status text not null,
    queue_name text not null,
    payload jsonb not null,
    attempts integer not null,
    max_attempts integer not null,
    dedupe_key text,
    error_message text,
    started_at timestamptz,
    finished_at timestamptz,
    created_at timestamptz not null,
    updated_at timestamptz not null
);
