-- Replace legacy RGB API/admin hostnames across all string-like columns.
--
-- Dry-run (report matches only):
--   make replace-domains
--
-- Apply changes:
--   make replace-domains-run
--
-- Or manually:
--   docker exec -i metargb-mysql mysql -uroot -proot_password metargb_db < scripts/replace-domains.sql

SET NAMES utf8mb4;
SET @dry_run := IFNULL(@dry_run, 0);

DROP PROCEDURE IF EXISTS metargb_replace_domains;

DELIMITER $$

CREATE PROCEDURE metargb_replace_domains()
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE v_table VARCHAR(64);
    DECLARE v_column VARCHAR(64);

    DECLARE cur CURSOR FOR
        SELECT TABLE_NAME, COLUMN_NAME
        FROM information_schema.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE()
          AND DATA_TYPE IN (
              'char', 'varchar', 'tinytext', 'text', 'mediumtext',
              'longtext', 'json', 'enum', 'set'
          )
        ORDER BY TABLE_NAME, ORDINAL_POSITION;

    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;

    SELECT CONCAT('schema=', DATABASE(), ', dry_run=', @dry_run) AS info;

    OPEN cur;

    read_loop: LOOP
        FETCH cur INTO v_table, v_column;
        IF done THEN
            LEAVE read_loop;
        END IF;

        -- api.rgb.irpsc.com -> api.metarang.com
        SET @count_sql := CONCAT(
            'SELECT COUNT(*) INTO @cnt FROM `',
            REPLACE(v_table, '`', '``'),
            '` WHERE `',
            REPLACE(v_column, '`', '``'),
            '` LIKE ''%api.rgb.irpsc.com%'''
        );
        PREPARE count_stmt FROM @count_sql;
        EXECUTE count_stmt;
        DEALLOCATE PREPARE count_stmt;

        IF @cnt > 0 THEN
            SELECT v_table AS table_name, v_column AS column_name,
                   'api.rgb.irpsc.com' AS old_value, @cnt AS row_count,
                   IF(@dry_run, 'would_update', 'updated') AS action;

            IF @dry_run = 0 THEN
                SET @update_sql := CONCAT(
                    'UPDATE `', REPLACE(v_table, '`', '``'),
                    '` SET `', REPLACE(v_column, '`', '``'),
                    '` = REPLACE(`', REPLACE(v_column, '`', '``'),
                    '`, ''api.rgb.irpsc.com'', ''api.metarang.com'') WHERE `',
                    REPLACE(v_column, '`', '``'),
                    '` LIKE ''%api.rgb.irpsc.com%'''
                );
                PREPARE update_stmt FROM @update_sql;
                EXECUTE update_stmt;
                DEALLOCATE PREPARE update_stmt;
            END IF;
        END IF;

        -- admin.rgb.irpsc.com -> admin.metarang.com
        SET @count_sql := CONCAT(
            'SELECT COUNT(*) INTO @cnt FROM `',
            REPLACE(v_table, '`', '``'),
            '` WHERE `',
            REPLACE(v_column, '`', '``'),
            '` LIKE ''%admin.rgb.irpsc.com%'''
        );
        PREPARE count_stmt FROM @count_sql;
        EXECUTE count_stmt;
        DEALLOCATE PREPARE count_stmt;

        IF @cnt > 0 THEN
            SELECT v_table AS table_name, v_column AS column_name,
                   'admin.rgb.irpsc.com' AS old_value, @cnt AS row_count,
                   IF(@dry_run, 'would_update', 'updated') AS action;

            IF @dry_run = 0 THEN
                SET @update_sql := CONCAT(
                    'UPDATE `', REPLACE(v_table, '`', '``'),
                    '` SET `', REPLACE(v_column, '`', '``'),
                    '` = REPLACE(`', REPLACE(v_column, '`', '``'),
                    '`, ''admin.rgb.irpsc.com'', ''admin.metarang.com'') WHERE `',
                    REPLACE(v_column, '`', '``'),
                    '` LIKE ''%admin.rgb.irpsc.com%'''
                );
                PREPARE update_stmt FROM @update_sql;
                EXECUTE update_stmt;
                DEALLOCATE PREPARE update_stmt;
            END IF;
        END IF;
    END LOOP;

    CLOSE cur;
END$$

DELIMITER ;

CALL metargb_replace_domains();
DROP PROCEDURE IF EXISTS metargb_replace_domains;
