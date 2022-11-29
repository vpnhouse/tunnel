UPDATE
    peers
SET
    activity = :activity,
    upstream = :upstream,
    downstream = :downstream,
    updated = :updated
WHERE
    id = :id
