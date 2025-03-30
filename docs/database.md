# Database Structure

The new `Siremis` written in Go does not use a database for its own needs. Instead it uses JSON configuration files to get the details for user access, schemas of database tables and the menu.

It is required to have the Kamailio database created, either with `kamctl` or
`kamcli`:

```shell
kamdbctl create

# or

kamcli db create
```

See more details in the Kamailio installation guidelines:

  - [https://kamailio.org/docs/tutorials/6.0.x/kamailio-install-guide-git/](https://kamailio.org/docs/tutorials/6.0.x/kamailio-install-guide-git/)

## Kamailio Database

For accounting, CDRs and statistics, the Kamailio database has to be extended. It
also adds a record to the domain table with `127.0.0.1` as local domain, just to
have a select value for the domain fields in the input forms.

Next are the sql statements needed for this:

```sql
DROP TABLE IF EXISTS acc;

CREATE TABLE `acc` (
  `id` int(10) unsigned NOT NULL auto_increment,
  `method` varchar(16) NOT NULL default '',
  `from_tag` varchar(64) NOT NULL default '',
  `to_tag` varchar(64) NOT NULL default '',
  `callid` varchar(128) NOT NULL default '',
  `sip_code` char(3) NOT NULL default '',
  `sip_reason` varchar(32) NOT NULL default '',
  `time` datetime NOT NULL default '2000-01-01 00:00:00',
  `src_ip` varchar(64) NOT NULL default '',
  `dst_ouser` VARCHAR(64) NOT NULL DEFAULT '',
  `dst_user` varchar(64) NOT NULL default '',
  `dst_domain` varchar(128) NOT NULL default '',
  `src_user` varchar(64) NOT NULL default '',
  `src_domain` varchar(128) NOT NULL default '',
  `cdr_id` integer NOT NULL default '0',
  INDEX acc_callid (`callid`),
  PRIMARY KEY  (`id`)
);

DROP TABLE IF EXISTS missed_calls;

CREATE TABLE `missed_calls` (
  `id` int(10) unsigned NOT NULL auto_increment,
  `method` varchar(16) NOT NULL default '',
  `from_tag` varchar(64) NOT NULL default '',
  `to_tag` varchar(64) NOT NULL default '',
  `callid` varchar(128) NOT NULL default '',
  `sip_code` char(3) NOT NULL default '',
  `sip_reason` varchar(32) NOT NULL default '',
  `time` datetime NOT NULL default '2000-01-01 00:00:00',
  `src_ip` varchar(64) NOT NULL default '',
  `dst_ouser` VARCHAR(64) NOT NULL DEFAULT '',
  `dst_user` varchar(64) NOT NULL default '',
  `dst_domain` varchar(128) NOT NULL default '',
  `src_user` varchar(64) NOT NULL default '',
  `src_domain` varchar(128) NOT NULL default '',
  `cdr_id` integer NOT NULL default '0',
  INDEX mc_callid (`callid`),
  PRIMARY KEY  (`id`)
);

DROP TABLE IF EXISTS `cdrs`;

CREATE TABLE `cdrs` (
  `cdr_id` bigint(20) NOT NULL auto_increment,
  `src_username` varchar(64) NOT NULL default '',
  `src_domain` varchar(128) NOT NULL default '',
  `dst_username` varchar(64) NOT NULL default '',
  `dst_domain` varchar(128) NOT NULL default '',
  `dst_ousername` varchar(64) NOT NULL default '',
  `call_start_time` datetime NOT NULL default '2000-01-01 00:00:00',
  `duration` int(10) unsigned NOT NULL default '0',
  `sip_call_id` varchar(128) NOT NULL default '',
  `sip_from_tag` varchar(128) NOT NULL default '',
  `sip_to_tag` varchar(128) NOT NULL default '',
  `src_ip` varchar(64) NOT NULL default '',
  `cost` integer NOT NULL default '0',
  `rated` integer NOT NULL default '0',
  `created` datetime NOT NULL,
  PRIMARY KEY  (`cdr_id`),
  UNIQUE KEY `uk_cft` (`sip_call_id`,`sip_from_tag`,`sip_to_tag`)
);

DROP TABLE IF EXISTS `billing_rates`;

CREATE TABLE `billing_rates` (
  `rate_id` bigint(20) NOT NULL auto_increment,
  `rate_group` varchar(64) NOT NULL default 'default',
  `prefix` varchar(64) NOT NULL default '',
  `rate_unit` integer NOT NULL default '0',
  `time_unit` integer NOT NULL default '60',
  PRIMARY KEY  (`rate_id`),
  UNIQUE KEY `uk_rp` (`rate_group`,`prefix`)
);

DROP PROCEDURE IF EXISTS `kamailio_cdrs`;
DROP PROCEDURE IF EXISTS `kamailio_rating`;

DELIMITER %%

CREATE PROCEDURE `kamailio_cdrs`()
BEGIN
  DECLARE done INT DEFAULT 0;
  DECLARE bye_record INT DEFAULT 0;
  DECLARE v_src_user,v_src_domain,v_dst_user,v_dst_domain,v_dst_ouser,v_callid,
     v_from_tag,v_to_tag,v_src_ip VARCHAR(64);
  DECLARE v_inv_time, v_bye_time DATETIME;
  DECLARE inv_cursor CURSOR FOR SELECT src_user, src_domain, dst_user,
     dst_domain, dst_ouser, time, callid,from_tag, to_tag, src_ip
     FROM acc
     where method='INVITE' and cdr_id='0';
  DECLARE CONTINUE HANDLER FOR SQLSTATE '02000' SET done = 1;
  OPEN inv_cursor;
  REPEAT
    FETCH inv_cursor INTO v_src_user, v_src_domain, v_dst_user, v_dst_domain,
            v_dst_ouser, v_inv_time, v_callid, v_from_tag, v_to_tag, v_src_ip;
    IF NOT done THEN
      SET bye_record = 0;
      SELECT 1, time INTO bye_record, v_bye_time FROM acc WHERE
                 method='BYE' AND callid=v_callid AND ((from_tag=v_from_tag
                 AND to_tag=v_to_tag)
                 OR (from_tag=v_to_tag AND to_tag=v_from_tag))
                 ORDER BY time ASC LIMIT 1;
      IF bye_record = 1 THEN
        INSERT INTO cdrs (src_username,src_domain,dst_username,
                 dst_domain,dst_ousername,call_start_time,duration,sip_call_id,
                 sip_from_tag,sip_to_tag,src_ip,created) VALUES (v_src_user,
                 v_src_domain,v_dst_user,v_dst_domain,v_dst_ouser,v_inv_time,
                 UNIX_TIMESTAMP(v_bye_time)-UNIX_TIMESTAMP(v_inv_time),
                 v_callid,v_from_tag,v_to_tag,v_src_ip,NOW());
        UPDATE acc SET cdr_id=last_insert_id() WHERE callid=v_callid
                 AND from_tag=v_from_tag AND to_tag=v_to_tag;
      END IF;
      SET done = 0;
    END IF;
  UNTIL done END REPEAT;
END

%%

CREATE PROCEDURE `kamailio_rating`(`rgroup` varchar(64))
BEGIN
  DECLARE done, rate_record, vx_cost INT DEFAULT 0;
  DECLARE v_cdr_id BIGINT DEFAULT 0;
  DECLARE v_duration, v_rate_unit, v_time_unit INT DEFAULT 0;
  DECLARE v_dst_username VARCHAR(64);
  DECLARE cdrs_cursor CURSOR FOR SELECT cdr_id, dst_username, duration
     FROM cdrs WHERE rated=0;
  DECLARE CONTINUE HANDLER FOR SQLSTATE '02000' SET done = 1;
  OPEN cdrs_cursor;
  REPEAT
    FETCH cdrs_cursor INTO v_cdr_id, v_dst_username, v_duration;
    IF NOT done THEN
      SET rate_record = 0;
      SELECT 1, rate_unit, time_unit INTO rate_record, v_rate_unit, v_time_unit
             FROM billing_rates
             WHERE rate_group=rgroup AND v_dst_username LIKE concat(prefix, '%')
             ORDER BY prefix DESC LIMIT 1;
      IF rate_record = 1 THEN
        SET vx_cost = v_rate_unit * CEIL(v_duration/v_time_unit);
        UPDATE cdrs SET rated=1, cost=vx_cost WHERE cdr_id=v_cdr_id;
      END IF;
      SET done = 0;
    END IF;
  UNTIL done END REPEAT;
END

%%

DELIMITER ;

DROP TABLE IF EXISTS `statistics`;

CREATE TABLE `statistics` (
  `id` int(10) unsigned NOT NULL auto_increment,
  `time_stamp` int(10) unsigned NOT NULL default '0',
  `shm_used_size` int(10) unsigned NOT NULL default '0',
  `shm_real_used_size` int(10) unsigned NOT NULL default '0',
  `shm_max_used_size` int(10) unsigned NOT NULL default '0',
  `shm_free_used_size` int(10) unsigned NOT NULL default '0',
  `ul_users` int(10) unsigned NOT NULL default '0',
  `ul_contacts` int(10) unsigned NOT NULL default '0',
  `tm_active` int(10) unsigned NOT NULL default '0',
  `rcv_req_diff` int(10) unsigned NOT NULL default '0',
  `fwd_req_diff` int(10) unsigned NOT NULL default '0',
  `2xx_trans_diff` int(10) unsigned NOT NULL default '0',
  PRIMARY KEY  (`id`)
);

-- insert a default domain
INSERT INTO domain (domain, did) VALUES ('127.0.0.1', 'default');
```

IMPORTANT: the above SQL statements **drop** the existing `acc` and `missed_calls` tables before creating them again. Do not use this script if you have data in those tables that you want to keep, instead you can use `ALTER TABLE` statements to add the new columns.

```sql
  ALTER TABLE acc ADD COLUMN src_user VARCHAR(64) NOT NULL DEFAULT '';
  ALTER TABLE acc ADD COLUMN src_domain VARCHAR(128) NOT NULL DEFAULT '';
  ALTER TABLE acc ADD COLUMN src_ip varchar(64) NOT NULL default '';
  ALTER TABLE acc ADD COLUMN dst_ouser VARCHAR(64) NOT NULL DEFAULT '';
  ALTER TABLE acc ADD COLUMN dst_user VARCHAR(64) NOT NULL DEFAULT '';
  ALTER TABLE acc ADD COLUMN dst_domain VARCHAR(128) NOT NULL DEFAULT '';
  ALTER TABLE acc ADD COLUMN cdr_id integer NOT NULL default '0';
  ALTER TABLE missed_calls ADD COLUMN src_user VARCHAR(64) NOT NULL DEFAULT '';
  ALTER TABLE missed_calls ADD COLUMN src_domain VARCHAR(128) NOT NULL DEFAULT '';
  ALTER TABLE missed_calls ADD COLUMN src_ip varchar(64) NOT NULL default '';
  ALTER TABLE missed_calls ADD COLUMN dst_ouser VARCHAR(64) NOT NULL DEFAULT '';
  ALTER TABLE missed_calls ADD COLUMN dst_user VARCHAR(64) NOT NULL DEFAULT '';
  ALTER TABLE missed_calls ADD COLUMN dst_domain VARCHAR(128) NOT NULL DEFAULT '';
  ALTER TABLE missed_calls ADD COLUMN cdr_id integer NOT NULL default '0';
```

The statistics table has to be filled by Kamailio via its configuration file. Next
is a snipped to do that.

```shell
loadmodule "rtimer.so"
loadmodule "sqlops.so"
loadmodule "htable.so"

...

modparam("rtimer", "timer", "name=tst;interval=300;mode=1;")
modparam("rtimer", "exec", "timer=tst;route=STATS")

modparam("sqlops","sqlcon",
         "ca=>mysql://kamailio:kamailiorw@localhost/kamailio")

modparam("htable", "htable", "stats=>size=6;")

...

route[STATS] {

	# clean very old records
	$var(tmc) = $var(tmc) + 1;
	$var(x) = $var(tmc) mod 144;
	if($var(x) == 0)
	    sql_query("ca",
			"delete from statistics where time_stamp<$Ts - 864000",
			"ra");

	# insert values for Kamailio internal statistics
	sql_query("ca",
		"insert into statistics (time_stamp,shm_used_size,"
		"shm_real_used_size,shm_max_used_size,shm_free_used_size,"
		"ul_users,ul_contacts) values ($Ts,$stat(used_size),"
		"$stat(real_used_size),$stat(max_used_size),$stat(free_size),"
		"$stat(location-users),$stat(location-contacts))",
		"ra");

	# init the values for first execution, compute the diff for the rest
	if($var(tmc)==1)
	{
		$var(rcv_req_diff) = $stat(rcv_requests);
		$var(fwd_req_diff) = $stat(fwd_requests);
		$var(2xx_trans_diff) = $stat(2xx_transactions);
	} else {
		$var(rcv_req_diff) = $stat(rcv_requests) - $sht(stats=>last_rcv_req);
		$var(fwd_req_diff) = $stat(fwd_requests) - $sht(stats=>last_fwd_req);
		$var(2xx_trans_diff) = $stat(2xx_transactions)
									- $sht(stats=>last_2xx_trans);
	}
	# update the values for stats stored in cache (htable)
	$sht(stats=>last_rcv_req) = $stat(rcv_requests);
	$sht(stats=>last_fwd_req) = $stat(fwd_requests);
	$sht(stats=>last_2xx_trans) = $stat(2xx_transactions);

	# insert values for stats computed in config
	sql_query("ca",
		"update statistics set tm_active=$stat(inuse_transactions),"
		"rcv_req_diff=$var(rcv_req_diff),fwd_req_diff=$var(fwd_req_diff),"
		"2xx_trans_diff=$var(2xx_trans_diff) where time_stamp=$Ts",
		"ra");
}
```