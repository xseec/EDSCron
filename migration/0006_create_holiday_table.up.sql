CREATE TABLE IF NOT EXISTS `holiday` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `area` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '地区，如china',
  `date` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '日期，如2025-03-18',
  `category` enum('假日','调休工作日','离峰日') COLLATE utf8_bin NOT NULL DEFAULT '假日',
  `detail` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '假日类型，如元旦',
  `alias` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '地区别称，如中国',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `area_date` (`area`,`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin COMMENT='节假日';