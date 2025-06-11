DROP TABLE IF EXISTS `orders`;
CREATE TABLE `orders` (
  `iOrderID` int unsigned NOT NULL AUTO_INCREMENT,
  `vRzpOrderID` varchar(64) DEFAULT NULL,
  `vRcptID` varchar(40) NOT NULL,
  `vName` varchar(128) NOT NULL,
  `vEmail` varchar(128) NOT NULL,
  `iAmount` int NOT NULL,
  `vProject` varchar(64) NOT NULL,
  `vStatus` varchar(256) DEFAULT NULL,
  `vReturnStatus` text,
  `dtCreatedAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `dtUpdatedAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`iOrderID`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci