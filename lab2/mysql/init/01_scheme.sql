use `lab2`;

DROP TABLE IF EXISTS `coupon_master`;
CREATE TABLE `coupon_master` (
  `coupon_type` varchar(100) NOT NULL PRIMARY KEY,
  `amount` int unsigned NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
