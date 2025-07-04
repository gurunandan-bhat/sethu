DROP TABLE IF EXISTS `orders`;
CREATE TABLE `orders` (
  `iDonationID` int unsigned NOT NULL AUTO_INCREMENT,
  `vRzpOrderID` varchar(64) DEFAULT NULL,
  `vRzpPaymentID` varchar(64) DEFAULT NULL,
  `vRcptID` varchar(40) NOT NULL,
  `vName` varchar(128) NOT NULL,
  `vEmail` varchar(128) NOT NULL,
  `vTelephone` varchar(24) DEFAULT NULL,
  `vAddr1` varchar(128) NOT NULL,
  `vAddr2` varchar(128) DEFAULT NULL,
  `vCity` varchar(64) NOT NULL,
  `vPin` varchar(16) NOT NULL,  
  `vState` varchar(16) NOT NULL,
  `vPAN` varchar(64) DEFAULT NULL,
  `iAmount` int NOT NULL,
  `vProject` varchar(64) NOT NULL,
  `vStatus` varchar(256) DEFAULT NULL,
  `vReturnStatus` text,
  `dtCreatedAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `dtUpdatedAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`iDonationID`)
) ENGINE=InnoDB AUTO_INCREMENT=73 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci