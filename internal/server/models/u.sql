CREATE TABLE IF NOT EXISTS `m_player` (
  `player_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `public_id` bigint(20) unsigned DEFAULT NULL,
  `account_uid` varchar(32) NOT NULL,
  `server_id` int(10) NOT NULL,
  `platform_id` tinyint(3) unsigned NOT NULL,
  `channel_id` int(10) unsigned NOT NULL,
  `player_name` varchar(16) DEFAULT NULL,
  `role_name` varchar(16) DEFAULT NULL,
  `account_name` varchar(255) DEFAULT NULL,
  `sex` tinyint(3) unsigned NOT NULL,
  `level` smallint(5) unsigned NOT NULL,
  `first_pay_time` int(10) NOT NULL,
  `reg_time` int(10) NOT NULL,
  `last_time` int(10) unsigned NOT NULL,
  `pay_total` int(10) unsigned NOT NULL,
  `reg_device_id` varchar(180) NOT NULL,
  `last_device_id` varchar(180) NOT NULL,
  `is_thaw` int(10) NOT NULL DEFAULT '0',
  `thaw_at` int(10) NOT NULL DEFAULT '0',
  PRIMARY KEY (`player_id`),
  UNIQUE KEY `uuid_server_id` (`account_uid`,`server_id`) USING BTREE,
  UNIQUE KEY `public_id` (`public_id`),
  KEY `idx_reg_time` (`reg_time`) USING BTREE,
  KEY `idx_first_pay_time` (`first_pay_time`) USING BTREE,
  KEY `idx_uname` (`player_name`) USING BTREE,
  KEY `idx_account_name` (`account_name`(191)) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1501 DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `u_item` (
  `player_id` bigint(20) unsigned NOT NULL,
  `item_id` int(10) unsigned NOT NULL,
  `item_type` int(10) unsigned NOT NULL,
  `num` int(10) unsigned NOT NULL,
  `update_at` int(10) unsigned NOT NULL,
  `create_at` int(10) unsigned NOT NULL,
  PRIMARY KEY (`player_id`,`item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `u_player` (
  `player_id` bigint(20) unsigned NOT NULL,
  `platform_id` int(10) unsigned NOT NULL DEFAULT "0",
  `channel_id` int(10) unsigned NOT NULL DEFAULT "0",
  `server_id` int(10) unsigned NOT NULL DEFAULT "0",
  `level` int(10) unsigned NOT NULL DEFAULT "0",
  `event_at` int(10) unsigned NOT NULL DEFAULT "0",
  `reg_time` int(10) unsigned NOT NULL,
  `last_ac` varchar(255) DEFAULT "",
  `last_ac_time` int(10) unsigned NOT NULL DEFAULT "0",
  PRIMARY KEY (`player_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;