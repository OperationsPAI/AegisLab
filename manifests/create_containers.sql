-- Active: 1750638419546@@10.10.10.220@32432@rcabench
INSERT INTO
    containers (
        type,
        name,
        image,
        tag,
        created_at,
        updated_at
    )
VALUES (
        'algorithm',
        'detector',
        'detector',
        'latest',
        now(),
        now()
    ),
    (
        'algorithm',
        'random',
        'rca-algo-random',
        'latest',
        now(),
        now()
    ),
    (
        'benchmark',
        'clickhouse',
        'clickhouse_dataset',
        '351c949',
        now(),
        now()
    )