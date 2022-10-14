CREATE TABLE `push_jobs` (
 `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
 `queue` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '队列名称',
 `payload` longtext COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'json消息',
 `update_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最新更新时间',
 `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
 PRIMARY KEY (`id`),
 KEY `jobs_queue_index` (`queue`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 DEFAULT COLLATE = utf8mb4_unicode_ci
ALTER TABLE push_jobs ADD del ENUM("0","1") NOT NULL DEFAULT "0";
ALTER TABLE push_jobs ADD attempts int NOT NULL  default 0;
ALTER TABLE push_jobs ADD last_error longtext;
