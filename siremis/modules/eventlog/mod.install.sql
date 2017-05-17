/*Table structure for table `event_log` */

DROP TABLE IF EXISTS `event_log`;

CREATE TABLE `event_log` (                                                                                                                    
     `id` int(11) NOT NULL auto_increment,                                                                                    
     `user_id` int(11) NOT NULL default '0',                                                                                  
     `ipaddr` varchar(50) NOT NULL,                                                                                     
     `event` varchar(255) NOT NULL,                                                                                       
     `message` varchar(255) NOT NULL,                                                                               
     `comment` text NOT NULL,                                                                                       
     `timestamp` timestamp NOT NULL default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP,                          
     PRIMARY KEY  (`id`),                                                                                                                        
     KEY `UserID` (`user_id`,`ipaddr`,`event`),                                                                                                  
     KEY `Message` (`message`)                                                                                                                   
   ) ENGINE=MyISAM DEFAULT CHARSET=utf8;

