CREATE TABLE IF NOT EXISTS `carbon` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `area` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '区域，如全国、华东、福建',
  `year` int(11) NOT NULL DEFAULT '0' COMMENT '年份',
  `value` float NOT NULL DEFAULT '0' COMMENT '碳排因子',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `area_year` (`area`,`year`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin COMMENT='净购入电力碳排因子';