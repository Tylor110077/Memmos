-- name: CreateGroup :one
insert into groups (
    id,
    name,
    created_at,
    updated_at,
    deleted_at
) values (
    $1, $2, $3, $4, $5
)
returning *;

-- name: GetGroup :one
select *
from groups
where id = $1 and deleted_at is null;

-- name: ListGroups :many
select *
from groups
where deleted_at is null
order by created_at asc;

-- name: UpdateGroupName :one
update groups
set name = $2,
    updated_at = $3
where id = $1 and deleted_at is null
returning *;

-- name: DeleteGroup :exec
update groups
set deleted_at = $2,
    updated_at = $2
where id = $1 and deleted_at is null;

-- name: ExistsGroupByName :one
select exists(
    select 1
    from groups
    where lower(name) = lower($1)
      and deleted_at is null
      and id <> $2
);
