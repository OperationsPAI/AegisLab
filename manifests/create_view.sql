CREATE VIEW fault_injection_no_issues AS
SELECT DISTINCT fis.id DatasetID, fis.display_config, fis.engine_config, fis.pre_duration, fis.injection_name
FROM fault_injection_schedules fis
         JOIN execution_results er ON fis.id = er.dataset
         JOIN detectors d ON er.id = d.execution_id
WHERE d.issues='{}';

CREATE VIEW fault_injection_with_issues AS
SELECT fis.id DatasetID, fis.display_config, fis.engine_config, fis.pre_duration, fis.injection_name, d.issues
FROM fault_injection_schedules fis
         JOIN execution_results er ON fis.id = er.dataset
         JOIN detectors d ON er.id = d.execution_id
WHERE d.issues != '{}';