-- Attribute building satisfaction spending to a specific user.
-- Apply to existing databases that were created from an older schema.sql.

ALTER TABLE `buildings`
  ADD COLUMN `user_id` bigint(20) unsigned NULL AFTER `feature_id`,
  ADD KEY `idx_buildings_user_id` (`user_id`);
