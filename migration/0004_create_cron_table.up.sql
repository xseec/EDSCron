CREATE TABLE `cron` (
	`id` INT(11) NOT NULL AUTO_INCREMENT,
	`category` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '类型：weather,dlgd,holiday,carbon,re-dlgd' COLLATE 'utf8_bin',
	`task` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '任务' COLLATE 'utf8_bin',
	`scheduler` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '定时(*s *m *h *d *M)' COLLATE 'utf8_bin',
	`time` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '时间参数模板' COLLATE 'utf8_bin',
	`comment` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '备注' COLLATE 'utf8_bin',
	`start_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '生效时间',
	`delta_time` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '时间间隔(+y,+M,+d,+h,+m,+s)' COLLATE 'utf8_bin',
	`create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (`id`) USING BTREE
)
COMMENT='定时任务表。下日气温，下月电价，下年假日等'
COLLATE='utf8_bin'
ENGINE=InnoDB
;