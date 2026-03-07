-- name: CreateResource :one
insert into resources (
    id,
    group_id,
    name,
    type,
    status,
    failed_stage,
    error_message,
    created_at,
    updated_at,
    deleted_at
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
returning *;

-- name: GetResource :one
select *
from resources
where id = $1 and deleted_at is null;

-- name: ListResourcesByGroup :many
select *
from resources
where group_id = $1 and deleted_at is null
order by created_at desc;

-- name: UpdateResourceStatus :one
update resources
set status = $2,
    failed_stage = $3,
    error_message = $4,
    updated_at = $5
where id = $1 and deleted_at is null
returning *;

-- name: DeleteResource :exec
update resources
set deleted_at = $2,
    updated_at = $2
where id = $1 and deleted_at is null;
