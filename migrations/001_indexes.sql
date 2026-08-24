CREATE PRIMARY INDEX IF NOT EXISTS `idx_api_keys_primary`
ON `api_monetization`.`app`.`api_keys`;

CREATE INDEX IF NOT EXISTS `idx_api_keys_hash_status`
ON `api_monetization`.`app`.`api_keys`(hash, status);

CREATE PRIMARY INDEX IF NOT EXISTS `idx_wallets_primary`
ON `api_monetization`.`app`.`wallets`;

CREATE PRIMARY INDEX IF NOT EXISTS `idx_usage_primary`
ON `api_monetization`.`app`.`usage`;

CREATE PRIMARY INDEX IF NOT EXISTS `idx_invoices_primary`
ON `api_monetization`.`app`.`invoices`;
