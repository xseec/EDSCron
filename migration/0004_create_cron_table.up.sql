CREATE TABLE IF NOT EXISTS `cron` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `category` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '类型：weather,dlgd,holiday,carbon,re-dlgd',
  `scheduler` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '定时(*s *m *h *d *M)',
  `task` json NOT NULL COMMENT '任务',
  `time` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '时间参数模板',
  `comment` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '备注',
  `start_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '生效时间',
  `delta_time` varchar(50) COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT '时间间隔(+y,+M,+d,+h,+m,+s)',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin COMMENT='定时任务表。下日气温，下月电价，下年假日等';