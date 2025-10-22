CREATE TABLE IF NOT EXISTS `weather` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `date` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '日期',
  `city` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '城市',
  `day_weather` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '白天天气',
  `day_temp` float NOT NULL DEFAULT '0' COMMENT '白天气温',
  `night_weather` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '夜间天气',
  `night_temp` float NOT NULL DEFAULT '0' COMMENT '夜间气温',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `date_city` (`date`,`city`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin COMMENT='天气';