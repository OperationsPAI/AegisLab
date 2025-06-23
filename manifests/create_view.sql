DROP VIEW IF EXISTS fault_injection_no_issues;

DROP VIEW IF EXISTS fault_injection_with_issues;

CREATE OR REPLACE VIEW fault_injection_no_issues AS
SELECT DISTINCT
    fis.id AS dataset_id,
    fis.fault_type,
    fis.display_config,
    fis.engine_config,
    fis.pre_duration,
    fis.injection_name,
    fis.created_at
FROM
    fault_injection_schedules fis
    JOIN (
        SELECT id, dataset, algorithm, ROW_NUMBER() OVER (
                PARTITION BY
                    dataset, algorithm
                ORDER BY created_at DESC, id DESC
            ) as rn
        FROM execution_results
    ) er_ranked ON fis.injection_name = er_ranked.dataset
    AND er_ranked.rn = 1
    JOIN detectors d ON er_ranked.id = d.execution_id
WHERE
    d.issues = '{}';

CREATE OR REPLACE VIEW fault_injection_with_issues AS
SELECT DISTINCT
    fis.id AS dataset_id,
    fis.fault_type,
    fis.display_config,
    fis.engine_config,
    fis.pre_duration,
    fis.injection_name,
    fis.created_at,
    d.issues
FROM
    fault_injection_schedules fis
    JOIN (
        SELECT id, dataset, algorithm, ROW_NUMBER() OVER (
                PARTITION BY
                    dataset, algorithm
                ORDER BY created_at DESC, id DESC
            ) as rn
        FROM execution_results
    ) er_ranked ON fis.injection_name = er_ranked.dataset
    AND er_ranked.rn = 1
    JOIN detectors d ON er_ranked.id = d.execution_id
WHERE
    d.issues != '{}';