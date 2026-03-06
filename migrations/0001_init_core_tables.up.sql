create extension if not exists vector;

create table if not exists groups (
    id text primary key,
    name varchar(64) not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

create unique index if not exists idx_groups_name_active
    on groups (lower(name))
    where deleted_at is null;

create table if not exists resources (
    id text primary key,
    group_id text not null references groups(id) on delete cascade,
    name text not null,
    type text not null,
    status text not null,
    failed_stage text,
    error_message text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

create index if not exists idx_resources_group_id on resources(group_id) where deleted_at is null;

create table if not exists resource_artifacts (
    id text primary key,
    resource_id text not null references resources(id) on delete cascade,
    artifact_type text not null,
    storage_key text not null,
    content_type text,
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);

create index if not exists idx_resource_artifacts_resource_id on resource_artifacts(resource_id);

create table if not exists resource_chunks (
    id text primary key,
    resource_id text not null references resources(id) on delete cascade,
    chunk_index integer not null,
    content text not null,
    summary text,
    token_count integer,
    embedding vector(1536),
    created_at timestamptz not null default now()
);

create unique index if not exists idx_resource_chunks_resource_chunk
    on resource_chunks(resource_id, chunk_index);

create table if not exists graphs (
    id text primary key,
    group_id text not null references groups(id) on delete cascade,
    resource_id text references resources(id) on delete cascade,
    graph_type text not null,
    version integer not null,
    status text not null,
    title text,
    summary text,
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    archived_at timestamptz
);

create unique index if not exists idx_graphs_active_resource
    on graphs(resource_id, graph_type)
    where is_active = true and resource_id is not null;

create unique index if not exists idx_graphs_active_framework
    on graphs(group_id, graph_type)
    where is_active = true and graph_type = 'framework';

create table if not exists graph_nodes (
    id text primary key,
    graph_id text not null references graphs(id) on delete cascade,
    parent_node_id text references graph_nodes(id) on delete set null,
    name text not null,
    description text,
    meaning text,
    node_type text not null,
    source_type text not null,
    level integer not null default 0,
    is_expansion boolean not null default false,
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists idx_graph_nodes_graph_id on graph_nodes(graph_id);

create table if not exists graph_edges (
    id text primary key,
    graph_id text not null references graphs(id) on delete cascade,
    from_node_id text not null references graph_nodes(id) on delete cascade,
    to_node_id text not null references graph_nodes(id) on delete cascade,
    relation_type text not null,
    description text,
    is_expansion boolean not null default false,
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);

create index if not exists idx_graph_edges_graph_id on graph_edges(graph_id);

create table if not exists node_examples (
    id text primary key,
    node_id text not null references graph_nodes(id) on delete cascade,
    resource_chunk_id text references resource_chunks(id) on delete set null,
    content text not null,
    source_excerpt text,
    created_at timestamptz not null default now()
);

create index if not exists idx_node_examples_node_id on node_examples(node_id);

create table if not exists conversations (
    id text primary key,
    group_id text not null references groups(id) on delete cascade,
    graph_id text references graphs(id) on delete set null,
    node_id text references graph_nodes(id) on delete set null,
    title text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists idx_conversations_group_id on conversations(group_id);

create table if not exists conversation_messages (
    id text primary key,
    conversation_id text not null references conversations(id) on delete cascade,
    role text not null,
    content text not null,
    cited_chunk_ids text[] not null default '{}',
    cited_node_ids text[] not null default '{}',
    context_snapshot jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);

create index if not exists idx_conversation_messages_conversation_id
    on conversation_messages(conversation_id, created_at);

create table if not exists processing_jobs (
    id text primary key,
    group_id text not null references groups(id) on delete cascade,
    resource_id text references resources(id) on delete cascade,
    job_type text not null,
    status text not null,
    queue_name text not null,
    payload jsonb not null default '{}'::jsonb,
    attempts integer not null default 0,
    max_attempts integer not null default 0,
    dedupe_key text,
    error_message text,
    started_at timestamptz,
    finished_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists idx_processing_jobs_resource_id on processing_jobs(resource_id);

create table if not exists job_events (
    id text primary key,
    job_id text not null references processing_jobs(id) on delete cascade,
    event_type text not null,
    message text,
    payload jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);

create index if not exists idx_job_events_job_id on job_events(job_id, created_at);
