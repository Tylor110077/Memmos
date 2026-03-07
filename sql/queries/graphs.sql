-- name: CreateGraph :one
insert into graphs (
    id,
    group_id,
    resource_id,
    graph_type,
    version,
    status,
    title,
    summary,
    is_active,
    created_at,
    updated_at,
    archived_at
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
returning *;

-- name: ArchiveActiveGraphsForResource :exec
update graphs
set is_active = false,
    archived_at = $3,
    updated_at = $3
where resource_id = $1
  and graph_type = $2
  and is_active = true;

-- name: GetActiveGraphByResource :one
select *
from graphs
where resource_id = $1
  and graph_type = $2
  and is_active = true;

-- name: ListNodesByGraph :many
select *
from graph_nodes
where graph_id = $1
order by level asc, created_at asc;

-- name: ListEdgesByGraph :many
select *
from graph_edges
where graph_id = $1
order by created_at asc;

-- name: CreateGraphNode :one
insert into graph_nodes (
    id,
    graph_id,
    parent_node_id,
    name,
    description,
    meaning,
    node_type,
    source_type,
    level,
    is_expansion,
    metadata,
    created_at,
    updated_at
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
returning *;

-- name: CreateGraphEdge :one
insert into graph_edges (
    id,
    graph_id,
    from_node_id,
    to_node_id,
    relation_type,
    description,
    is_expansion,
    metadata,
    created_at
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
returning *;
