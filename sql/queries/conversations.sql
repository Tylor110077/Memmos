-- name: CreateConversation :one
insert into conversations (
    id,
    group_id,
    graph_id,
    node_id,
    title,
    created_at,
    updated_at
) values (
    $1, $2, $3, $4, $5, $6, $7
)
returning *;

-- name: GetConversation :one
select *
from conversations
where id = $1;

-- name: CreateConversationMessage :one
insert into conversation_messages (
    id,
    conversation_id,
    role,
    content,
    cited_chunk_ids,
    cited_node_ids,
    context_snapshot,
    created_at
) values (
    $1, $2, $3, $4, $5, $6, $7, $8
)
returning *;

-- name: ListConversationMessages :many
select *
from conversation_messages
where conversation_id = $1
order by created_at asc;
