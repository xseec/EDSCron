CREATE TABLE `dlgd` (
	`id` INT(11) NOT NULL AUTO_INCREMENT,
	`area` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '区域' COLLATE 'utf8_bin',
	`start_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '执行起始时间（含）',
	`end_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '执行截止时间（不含）',
	`category` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '用电分类，如单一制、两部制' COLLATE 'utf8_bin',
	`voltage` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '电压等级，如1-10（20）千伏' COLLATE 'utf8_bin',
	`stage` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '阶梯阈值，如深圳' COLLATE 'utf8_bin',
	`fund` FLOAT NOT NULL DEFAULT '0' COMMENT '政府性基金及附加，力调电费需排除',
	`sharp` FLOAT NOT NULL DEFAULT '0' COMMENT '尖峰电价',
	`sharp_date` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '尖峰日期' COLLATE 'utf8_bin',
	`sharp_hour` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '尖峰时段，如1100-1130,2200-0700' COLLATE 'utf8_bin',
	`peak` FLOAT NOT NULL DEFAULT '0' COMMENT '高峰电价',
	`peak_date` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '高峰日期' COLLATE 'utf8_bin',
	`peak_hour` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '高峰时段' COLLATE 'utf8_bin',
	`flat` FLOAT NOT NULL DEFAULT '0' COMMENT '平段电价',
	`flat_date` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '平段日期' COLLATE 'utf8_bin',
	`flat_hour` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '平段时段' COLLATE 'utf8_bin',
	`valley` FLOAT NOT NULL DEFAULT '0' COMMENT '低谷电价',
	`valley_date` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '低谷日期' COLLATE 'utf8_bin',
	`valley_hour` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '低谷时段' COLLATE 'utf8_bin',
	`deep` FLOAT NOT NULL DEFAULT '0' COMMENT '深谷电价',
	`deep_date` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '深谷日期，如weekend;holiday:元旦,春节;temp:35' COLLATE 'utf8_bin',
	`deep_hour` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '深谷时段' COLLATE 'utf8_bin',
	`demand` FLOAT NOT NULL DEFAULT '0' COMMENT '需量电价',
	`capacity` FLOAT NOT NULL DEFAULT '0' COMMENT '容量电价',
	`doc_no` VARCHAR(100) NOT NULL COMMENT '电价政策文号' COLLATE 'utf8_bin',
	`create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
	`update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (`id`) USING BTREE,
	UNIQUE INDEX `area_start_time_category_voltage_stage` (`area`, `start_time`, `category`, `voltage`, `stage`) USING BTREE
)
COMMENT='代理购电'
COLLATE='utf8_bin'
ENGINE=InnoDB
;