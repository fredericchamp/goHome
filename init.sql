--
-- Default database initialisation
--
-- One sql stmt per line
-- No multi-line stmt
-- A line stating with -- is ignore
-- An empty line (or only white spaces)  is ignore
-- Any other pattern wont work
--
create table goHome (id integer not null primary key, scope text, name text, value text);
create table Item (id integer not null primary key, Name text, idProfil integer, idItemType integer, idMasterItem integer, icone blob);
create table ItemField (id integer, idItem integer, nOrder integer, Name text, idDataType, Helper text, Rules text );
create table ItemFieldVal ( idObject integer, idField integer, intVal integer, floatVal float, textVal text, byteVal blob );
create table HistoSensor (ts datetime, idObject integer, intVal integer, floatVal float, textVal text);
create table HistoActor (ts datetime, idObject integer, Param text, Result text);

insert into goHome values    ( 0, 'goHome', 'InterfaceVersion', '1');
insert into goHome values    ( 1, 'goHome', 'server_name', 'localhost');
insert into goHome values    ( 2, 'goHome', 'https_port', '5100');
insert into goHome values    ( 3, 'goHome', 'email', 'admin@goHomeDomain.net');
insert into goHome values    ( 4, 'goHome', 'server_crt', '/var/goHome/certificats/server.crt.pem');
insert into goHome values    ( 5, 'goHome', 'server_key', '/var/goHome/certificats/server.key.pem');
insert into goHome values    ( 6, 'goHome', 'ca_crt', '/var/goHome/certificats/goHomeCAcert.pem');
insert into goHome values    ( 7, 'goHome', 'UserItemId', '1');

-- Note that userItemId param value in goHome table is directly link to the above insert statement
insert into Item values      ( 1, 'User', 1, 1, 0, null);
insert into ItemField values ( 1, 1, 1, 'FirstName', 4, '', '');
insert into ItemField values ( 2, 1, 2, 'LastName', 4, '', '');
insert into ItemField values ( 3, 1, 3, 'Email', 4, '', '{"uniq":1,"regexp":"^[[:alnum:].\-_]*@[[:alnum:].\-_]*[.][[:alpha:]]{2,}$"}');
insert into ItemField values ( 4, 1, 4, 'Phone', 4, '', '');
insert into ItemField values ( 5, 1, 5, 'IdProfil', 2, '{"Administrator":1,"User":2}', '');
insert into ItemField values ( 6, 1, 6, 'IsActive', 1, '{"Yes":1,"No":0}', '');

insert into ItemFieldVal values (  1,  1, 0, 0, 'Main', null );
insert into ItemFieldVal values (  1,  2, 0, 0, 'Administrator', null );
insert into ItemFieldVal values (  1,  3, 0, 0, 'main.admin@goHomeDomain.com', null );
insert into ItemFieldVal values (  1,  4, 0, 0, '1234567890', null);
insert into ItemFieldVal values (  1,  5, 1, 0, '', null);
insert into ItemFieldVal values (  1,  6, 1, 0, '', null);

insert into Item values      ( 2, 'Sensor', 1, 2, 0, null);
insert into ItemField values ( 8, 2, 1, 'Name', 4, '', '');
insert into ItemField values ( 9, 2, 2, 'IdProfil', 2, '{"Administrator":1,"User":2}', '');
insert into ItemField values (10, 2, 3, 'Record', 1, '{"Yes":1,"No":0}', '');
insert into ItemField values (11, 2, 4, 'IsInternal', 1, '{"Yes":1,"No":0}', '');
insert into ItemField values (12, 2, 5, 'ReadCmd', 4, '', '');
insert into ItemField values (13, 2, 6, 'ReadParam', 4, '', '');
insert into ItemField values (14, 2, 7, 'Interval', 4, '', '');
insert into ItemField values (15, 2, 8, 'IdDataType', 2, '{"Bool":1,"Int":2,"Float":3,"Text":4,"DateTime":5}', '');
insert into ItemField values (16, 2, 9, 'IsActive', 1, '{"Yes":1,"No":0}', '');

insert into ItemFieldVal values ( 10,  8, 0, 0, '%CPU', null );
insert into ItemFieldVal values ( 10,  9, 2, 0, '', null );
insert into ItemFieldVal values ( 10, 10, 0, 0, '', null );
insert into ItemFieldVal values ( 10, 11, 1, 0, '', null );
insert into ItemFieldVal values ( 10, 12, 0, 0, 'CpuUsage', null );
insert into ItemFieldVal values ( 10, 13, 0, 0, '', null );
insert into ItemFieldVal values ( 10, 14, 0, 0, '1m', null );
insert into ItemFieldVal values ( 10, 15, 2, 0, '', null );
insert into ItemFieldVal values ( 10, 16, 1, 0, '', null );

insert into ItemFieldVal values ( 11,  8, 0, 0, '%Memory', null );
insert into ItemFieldVal values ( 11,  9, 2, 0, '', null );
insert into ItemFieldVal values ( 11, 10, 0, 0, '', null );
insert into ItemFieldVal values ( 11, 11, 1, 0, '', null );
insert into ItemFieldVal values ( 11, 12, 0, 0, 'MemoryUsage', null );
insert into ItemFieldVal values ( 11, 13, 0, 0, '', null );
insert into ItemFieldVal values ( 11, 14, 0, 0, '1m', null );
insert into ItemFieldVal values ( 11, 15, 2, 0, '', null );
insert into ItemFieldVal values ( 11, 16, 1, 0, '', null );

insert into ItemFieldVal values ( 12,  8, 0, 0, 'Alarm', null );
insert into ItemFieldVal values ( 12,  9, 2, 0, '', null );
insert into ItemFieldVal values ( 12, 10, 0, 0, '', null );
insert into ItemFieldVal values ( 12, 11, 1, 0, '', null );
insert into ItemFieldVal values ( 12, 12, 0, 0, 'GPIO', null );
insert into ItemFieldVal values ( 12, 13, 0, 0, '{"pin":16,"operation":"read","repeat":5,"interval":"50ms","result":"min"}', null );
insert into ItemFieldVal values ( 12, 14, 0, 0, '1s', null );
insert into ItemFieldVal values ( 12, 15, 2, 0, '', null );
insert into ItemFieldVal values ( 12, 16, 1, 0, '', null );

insert into Item values      ( 3, 'Actor', 1, 3, 0, null);
insert into ItemField values (17, 3, 1, 'Name', 4, '', '');
insert into ItemField values (18, 3, 2, 'IdProfil', 2, '{"Administrator":1,"User":2}', '');
insert into ItemField values (19, 3, 3, 'IsInternal', 1, '{"Yes":1,"No":0}', '');
insert into ItemField values (20, 3, 4, 'ActCmd', 4, '', '');
insert into ItemField values (21, 3, 5, 'ActParam', 4, '', '');
insert into ItemField values (22, 3, 6, 'IsActive', 1, '{"Yes":1,"No":0}', '');

insert into ItemFieldVal values ( 20, 17, 0, 0, 'Portal', null );
insert into ItemFieldVal values ( 20, 18, 2, 0, '', null );
insert into ItemFieldVal values ( 20, 19, 1, 0, '', null );
insert into ItemFieldVal values ( 20, 20, 1, 0, 'GPIO', null );
insert into ItemFieldVal values ( 20, 21, 0, 0, '{"pin":21,"operation":"write","value":1,"duration":"2s"}', null );
insert into ItemFieldVal values ( 20, 22, 1, 0, '', null );

insert into ItemFieldVal values ( 21, 17, 0, 0, 'Garage', null );
insert into ItemFieldVal values ( 21, 18, 2, 0, '', null );
insert into ItemFieldVal values ( 21, 19, 1, 0, '', null );
insert into ItemFieldVal values ( 21, 20, 1, 0, 'GPIO', null );
insert into ItemFieldVal values ( 21, 21, 0, 0, '{"pin":14,"operation":"write","value":1,"duration":"2s"}', null );
insert into ItemFieldVal values ( 21, 22, 1, 0, '', null );

insert into ItemFieldVal values ( 22, 17, 0, 0, 'GsmReset', null );
insert into ItemFieldVal values ( 22, 18, 1, 0, '', null );
insert into ItemFieldVal values ( 22, 19, 1, 0, '', null );
insert into ItemFieldVal values ( 22, 20, 1, 0, 'GPIO', null );
insert into ItemFieldVal values ( 22, 21, 0, 0, '{"pin":6,"operation":"write","value":1,"duration":"100ms"}', null );
insert into ItemFieldVal values ( 22, 22, 1, 0, '', null );

insert into ItemFieldVal values ( 23, 17, 0, 0, 'SendSMS', null );
insert into ItemFieldVal values ( 23, 18, 1, 0, '', null );
insert into ItemFieldVal values ( 23, 19, 1, 0, '', null );
insert into ItemFieldVal values ( 23, 20, 1, 0, 'SerialATSMS', null );
insert into ItemFieldVal values ( 23, 21, 0, 0, '/dev/ttyAMA0', null );
insert into ItemFieldVal values ( 23, 22, 1, 0, '', null );

insert into Item values      ( 4, 'SensorAct', 1, 4, 2, null);
insert into ItemField values (23, 4, 1, 'idSensor', 2, 'sensor list', '');
insert into ItemField values (24, 4, 2, 'idActor', 2, 'actor list', '');
insert into ItemField values (25, 4, 3, 'Condition', 4, '', '');
insert into ItemField values (26, 4, 4, 'ActorParam', 4, '', '');
insert into ItemField values (27, 4, 5, 'IsActive', 1, '{"Yes":1,"No":0}', '');

insert into ItemFieldVal values ( 30, 23, 12, 0, '', null );
insert into ItemFieldVal values ( 30, 24, 23, 0, '', null );
insert into ItemFieldVal values ( 30, 25, 0, 0, '@lastVal@ != @prevVal@', null );
insert into ItemFieldVal values ( 30, 26, 0, 0, '{"phone":"+123123456789","message":"Alarm @lastVal@"}', null );
insert into ItemFieldVal values ( 30, 27, 1, 0, '', null );
