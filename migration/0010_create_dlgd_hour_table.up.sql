CREATE TABLE `dlgd_hour` (
	`id` INT(11) NOT NULL AUTO_INCREMENT,
	`area` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '区域' COLLATE 'utf8_bin',
	`doc_no` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '电价政策文号' COLLATE 'utf8_bin',
	`confirm` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '确认有效，默认0：未确认',
	`name` VARCHAR(50) NOT NULL COMMENT '时段名称' COLLATE 'utf8_bin',
	`value` VARCHAR(100) NOT NULL COMMENT '时段值' COLLATE 'utf8_bin',
	`temp` VARCHAR(50) NOT NULL COMMENT '温度条件，如：其他月份中广州日最高气温达到35℃' COLLATE 'utf8_bin',
	`months` VARCHAR(50) NOT NULL COMMENT '月份条件' COLLATE 'utf8_bin',
	`weekend_months` VARCHAR(50) NOT NULL COMMENT '月份-休息日条件' COLLATE 'utf8_bin',
	`holidays` VARCHAR(50) NOT NULL COMMENT '节假日条件，如：春节,劳动节,国庆节' COLLATE 'utf8_bin',
	`categories` VARCHAR(50) NOT NULL COMMENT '用电类别条件，如：容量315千伏安及以上,两部制' COLLATE 'utf8_bin',
	PRIMARY KEY (`id`) USING BTREE
)
COMMENT='代理购电-时段划分，字符串数组字段用,分割'
COLLATE='utf8_bin'
ENGINE=InnoDB
;