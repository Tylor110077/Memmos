-- name: CreateProcessingJob :one
insert into processing_jobs (
    id,
    group_id,
    resource_id,
    job_type,
    status,
    queue_name,
    payload,
    attempts,
    max_attempts,
    dedupe_key,
    error_message,
    started_at,
    finished_at,
    created_at,
    updated_at
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
)
returning *;

-- name: GetProcessingJob :one
select *
from processing_jobs
where id = $1;

-- name: ListProcessingJobsByResource :many
select *
from processing_jobs
where resource_id = $1
order by created_at desc;

-- name: UpdateProcessingJobStatus :one
update processing_jobs
set status = $2,
    error_message = $3,
    started_at = $4,
    finished_at = $5,
    attempts = $6,
    updated_at = $7
where id = $1
returning *;
