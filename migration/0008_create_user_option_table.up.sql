CREATE TABLE IF NOT EXISTS `user_option` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `account` varchar(50) COLLATE utf8_bin NOT NULL COMMENT '工程ID',
  `area` varchar(100) COLLATE utf8_bin NOT NULL COMMENT '区域、设备ID、支路名称或ID等',
  `category` varchar(100) COLLATE utf8_bin NOT NULL COMMENT '用电类别，如福建>工商业,两部制>1-10（20）千伏',
  `power_factor` double NOT NULL DEFAULT '0' COMMENT '[大陆]功率因素标准，限定0.8/0.85/0.9',
  `capacity` double NOT NULL DEFAULT '0' COMMENT '[大陆]合同容量(kVA)',
  `demand` double NOT NULL DEFAULT '0' COMMENT '[大陆]合同需量(kW)',
  `installed_cap` double NOT NULL DEFAULT '0' COMMENT '[台湾]装置契约(kW)',
  `regular_cap` double NOT NULL DEFAULT '0' COMMENT '[台湾]经常契约(kW)',
  `non_summer_cap` double NOT NULL DEFAULT '0' COMMENT '[台湾]非夏月契约(kW)',
  `semi_peak_cap` double NOT NULL DEFAULT '0' COMMENT '[台湾]半尖峰契约(kW)',
  `sat_semi_peak_cap` double NOT NULL DEFAULT '0' COMMENT '[台湾]周六半尖峰契约(kW)',
  `off_peak_cap` double NOT NULL DEFAULT '0' COMMENT '[台湾]离峰契约(kW)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `account_area` (`account`,`area`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin COMMENT='用电档案';