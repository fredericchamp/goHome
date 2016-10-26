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
create table ItemField (idField integer, idItem integer, nOrder integer, Name text, idDataType, Helper text, Regexp text );
create table ItemFieldVal ( idObject integer, idField integer, intVal integer, floatVal float, textVal text, byteVal blob );
create table HistoSensor (ts datetime, idObject integer, intVal integer, floatVal float, textVal text);
create table HistoActor (ts datetime, idObject integer, Param text);

insert into goHome values    ( 0, 'goHome', 'InterfaceVersion', '1');
insert into goHome values    ( 1, 'goHome', 'port', '6000');
insert into goHome values    ( 2, 'goHome', 'email', 'admin@goHomeDomain.com');

insert into Item values      ( 1, 'User', 1, 1, 0, null);
insert into ItemField values ( 1, 1, 1, 'FirstName', 4, '', '');
insert into ItemField values ( 2, 1, 2, 'LastName', 4, '', '');
insert into ItemField values ( 3, 1, 3, 'Email', 4, '', '');
insert into ItemField values ( 4, 1, 4, 'Phone', 4, '', '');
insert into ItemField values ( 5, 1, 5, 'IdProfil', 2, '{"Administrator":1,"User":2}', '');
insert into ItemField values ( 6, 1, 6, 'Cert', 6, '', '');

insert into ItemFieldVal values (  1,  1, 0, 0, 'Main', null ) ;
insert into ItemFieldVal values (  1,  2, 0, 0, 'Administrator', null ) ;
insert into ItemFieldVal values (  1,  3, 0, 0, 'main.admin@goHomeDomain.com', null ) ;
insert into ItemFieldVal values (  1,  4, 0, 0, '+1234567890', null);
insert into ItemFieldVal values (  1,  5, 1, 0, '', null);
insert into ItemFieldVal values (  1,  6, 0, 0, '', null);

insert into ItemFieldVal values (  2,  1, 0, 0, 'Frederic', null ) ;
insert into ItemFieldVal values (  2,  2, 0, 0, 'Champ', null ) ;
insert into ItemFieldVal values (  2,  3, 0, 0, 'fredchamp@goHomeDomain.com', null ) ;
insert into ItemFieldVal values (  2,  4, 0, 0, '1234567890', null);
insert into ItemFieldVal values (  2,  5, 1, 0, '', null);
insert into ItemFieldVal values (  2,  6, 0, 0, '', null);

insert into ItemFieldVal values (  3,  1, 0, 0, 'Frederic', null ) ;
insert into ItemFieldVal values (  3,  2, 0, 0, 'Champ', null ) ;
insert into ItemFieldVal values (  3,  3, 0, 0, 'fredchamp@goHomeDomain.com', null ) ;
insert into ItemFieldVal values (  3,  4, 0, 0, '1234567890', null);
insert into ItemFieldVal values (  3,  5, 1, 0, '', null);
insert into ItemFieldVal values (  3,  6, 0, 0, '', null);

insert into Item values      ( 2, 'Sensor', 1, 2, 0, null);
insert into ItemField values ( 7, 2, 1, 'Name', 4, '', '');
insert into ItemField values ( 8, 2, 2, 'IdProfil', 2, '{"Administrator":1,"User":2}', '');
insert into ItemField values ( 9, 2, 3, 'Record', 1, '{"Yes":1,"No":0}', '');
insert into ItemField values (10, 2, 4, 'IsInternal', 1, '{"Yes":1,"No":0}', '');
insert into ItemField values (11, 2, 5, 'ReadCmd', 4, '', '');
insert into ItemField values (12, 2, 6, 'ReadParam', 4, '', '');
insert into ItemField values (13, 2, 7, 'Interval', 4, '', '');
insert into ItemField values (14, 2, 8, 'IdDataType', 2, '{"Bool":1,"Int":2,"Float":3,"Text":4,"DateTime":4}', '');
insert into ItemField values (15, 2, 9, 'IsActive', 1, '{"Yes":1,"No":0}', '');

insert into ItemFieldVal values ( 10,  7, 0, 0, '%CPU', null ) ;
insert into ItemFieldVal values ( 10,  8, 2, 0, '', null ) ;
insert into ItemFieldVal values ( 10,  9, 0, 0, '', null ) ;
insert into ItemFieldVal values ( 10, 10, 1, 0, '', null ) ;
insert into ItemFieldVal values ( 10, 11, 0, 0, 'CpuUsage', null ) ;
insert into ItemFieldVal values ( 10, 12, 0, 0, '', null ) ;
insert into ItemFieldVal values ( 10, 13, 0, 0, '1m', null ) ;
insert into ItemFieldVal values ( 10, 14, 2, 0, '', null ) ;
insert into ItemFieldVal values ( 10, 15, 1, 0, '', null ) ;

insert into ItemFieldVal values ( 11,  7, 0, 0, 'Alarm', null ) ;
insert into ItemFieldVal values ( 11,  8, 2, 0, '', null ) ;
insert into ItemFieldVal values ( 11,  9, 0, 0, '', null ) ;
insert into ItemFieldVal values ( 11, 10, 1, 0, '', null ) ;
insert into ItemFieldVal values ( 11, 11, 0, 0, 'GPIO', null ) ;
insert into ItemFieldVal values ( 11, 12, 0, 0, '{"pin":16,"operation":"read","repeat":5,"interval":"50ms","result":"min"}', null ) ;
insert into ItemFieldVal values ( 11, 13, 0, 0, '1s', null ) ;
insert into ItemFieldVal values ( 11, 14, 2, 0, '', null ) ;
insert into ItemFieldVal values ( 11, 15, 1, 0, '', null ) ;

insert into ItemFieldVal values ( 12,  7, 0, 0, 'XMR', null ) ;
insert into ItemFieldVal values ( 12,  8, 1, 0, '', null ) ;
insert into ItemFieldVal values ( 12,  9, 1, 0, '', null ) ;
insert into ItemFieldVal values ( 12, 10, 1, 0, '', null ) ;
insert into ItemFieldVal values ( 12, 11, 0, 0, 'PoloTicker', null ) ;
insert into ItemFieldVal values ( 12, 12, 0, 0, '{"cmd":"get_ticker,"key":"SQDFQSD","secret":"QSERS54356TZE,"market":"XMR_BTC"}', null ) ;
insert into ItemFieldVal values ( 12, 13, 0, 0, '2m', null ) ;
insert into ItemFieldVal values ( 12, 14, 3, 0, '', null ) ;
insert into ItemFieldVal values ( 12, 15, 1, 0, '', null ) ;

insert into Item values      ( 3, 'Actor', 1, 3, 0, null);
insert into ItemField values (16, 3, 1, 'Name', 4, '', '');
insert into ItemField values (17, 3, 2, 'IdProfil', 2, '{"Administrator":1,"User":2}', '');
insert into ItemField values (18, 3, 3, 'IsInternal', 1, '{"Yes":1,"No":0}', '');
insert into ItemField values (19, 3, 4, 'ActCmd', 4, '', '');
insert into ItemField values (20, 3, 5, 'ActParam', 4, '', '');
insert into ItemField values (21, 3, 6, 'IsActive', 1, '{"Yes":1,"No":0}', '');

insert into ItemFieldVal values ( 20, 16, 0, 0, 'Portal', null ) ;
insert into ItemFieldVal values ( 20, 17, 2, 0, '', null ) ;
insert into ItemFieldVal values ( 20, 18, 1, 0, '', null ) ;
insert into ItemFieldVal values ( 20, 19, 1, 0, 'GPIO', null ) ;
insert into ItemFieldVal values ( 20, 20, 0, 0, '{"pin":21,"operation":"write","value":1,"duration":"2s"}', null ) ;
insert into ItemFieldVal values ( 20, 21, 1, 0, '', null ) ;

insert into ItemFieldVal values ( 21, 16, 0, 0, 'Garage', null ) ;
insert into ItemFieldVal values ( 21, 17, 2, 0, '', null ) ;
insert into ItemFieldVal values ( 21, 18, 1, 0, '', null ) ;
insert into ItemFieldVal values ( 21, 19, 1, 0, 'GPIO', null ) ;
insert into ItemFieldVal values ( 21, 20, 0, 0, '{"pin":14,"operation":"write","value":1,"duration":"2s"}', null ) ;
insert into ItemFieldVal values ( 21, 21, 1, 0, '', null ) ;

insert into ItemFieldVal values ( 22, 16, 0, 0, 'GsmReset', null ) ;
insert into ItemFieldVal values ( 22, 17, 1, 0, '', null ) ;
insert into ItemFieldVal values ( 22, 18, 1, 0, '', null ) ;
insert into ItemFieldVal values ( 22, 19, 1, 0, 'GPIO', null ) ;
insert into ItemFieldVal values ( 22, 20, 0, 0, '{"pin":6,"operation":"write","value":1,"duration":"100ms"}', null ) ;
insert into ItemFieldVal values ( 22, 21, 1, 0, '', null ) ;

insert into ItemFieldVal values ( 23, 16, 0, 0, 'SendSMS', null ) ;
insert into ItemFieldVal values ( 23, 17, 1, 0, '', null ) ;
insert into ItemFieldVal values ( 23, 18, 1, 0, '', null ) ;
insert into ItemFieldVal values ( 23, 19, 1, 0, 'SerialATSMS', null ) ;
insert into ItemFieldVal values ( 23, 20, 0, 0, '/dev/ttyAMA0', null ) ;
insert into ItemFieldVal values ( 23, 21, 1, 0, '', null ) ;

insert into Item values      ( 4, 'SensorAct', 1, 4, 2, null);
insert into ItemField values (22, 4, 1, 'idSensor', 2, 'sensor list', '');
insert into ItemField values (23, 4, 2, 'idActor', 2, 'actor list', '');
insert into ItemField values (24, 4, 3, 'Condition', 5, '', '');
insert into ItemField values (25, 4, 4, 'ActorParam', 4, '', '');
insert into ItemField values (26, 4, 5, 'IsActive', 1, '{"Yes":1,"No":0}', '');

insert into ItemFieldVal values ( 30, 22, 4, 0, '', null ) ;
insert into ItemFieldVal values ( 30, 23, 9, 0, '', null ) ;
insert into ItemFieldVal values ( 30, 24, 0, 0, '@value@ != @PrevValue@', null ) ;
insert into ItemFieldVal values ( 30, 25, 0, 0, '{"phone":"+123123456789","message":"Alarm @value@"}', null ) ;
insert into ItemFieldVal values ( 30, 26, 1, 0, '', null ) ;
