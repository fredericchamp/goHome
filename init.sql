--
-- Default database initialisation
--
-- One sql stmt per line
-- No multi-line stmt
-- A line stating with -- is ignore
-- An empty line (or only white spaces)  is ignore
-- Any other pattern wont work
--
create table goHome (idParam integer not null primary key, perimeter text, name text, value text);
create unique index goHome_Uniq on goHome (perimeter, name);

create table Item (idItem integer not null primary key, Name text, idProfil integer not null, idItemType integer not null, idMasterItem integer not null, iconeFileName text);

create table ItemField (idField integer not null primary key, idItem integer not null, nOrder integer not null, Name text, idDataType not null, Helper text, Rules text );
create unique index ItemField_Uniq on ItemField (idItem, nOrder);

create table ItemFieldVal ( idObject integer not null, idField integer not null, intVal integer not null, floatVal float not null, textVal text);
create unique index ItemFieldVal_PK on ItemFieldVal (idObject, idField);

create table HistoSensor (ts datetime not null, idObject integer not null, intVal integer not null, floatVal float not null, textVal text);
create unique index HistoSensor_PK on HistoSensor (ts, idObject);

create table HistoActor (ts datetime not null, idObject integer not null, idUser int not null, Param text, Result text);
create unique index HistoActor_PK on HistoActor (ts, idObject, idUser);


insert into goHome values    ( 0, 'Global', 'InterfaceVersion', '1');
insert into goHome values    ( 1, 'Global', 'email', 'admin@goHomeDomain.net');
insert into goHome values    ( 2, 'Global', 'UserItemId', '1');
insert into goHome values    ( 3, 'Http', 'server_name', 'localhost');
insert into goHome values    ( 4, 'Http', 'https_port', '5100');
insert into goHome values    ( 5, 'Http', 'server_crt', '/var/goHome/certificats/server.crt.pem');
insert into goHome values    ( 6, 'Http', 'server_key', '/var/goHome/certificats/server.key.pem');
insert into goHome values    ( 7, 'Http', 'ca_crt', '/var/goHome/certificats/goHomeCAcert.pem');
insert into goHome values    ( 8, 'Http', 'fileserver_root', '/var/goHome/www');

insert into Item select g.value, 'User', 1, 1, 0, '' from goHome g where g.perimeter='Global' and g.name ='UserItemId';
insert into ItemField values ( 1, 1, 1, 'FirstName', 4, '', '');
insert into ItemField values ( 2, 1, 2, 'LastName', 4, '', '');
insert into ItemField values ( 3, 1, 3, 'Email', 4, '', '{"uniq":1,"regexp":"^[[:alnum:].\-_]*@[[:alnum:].\-_]*[.][[:alpha:]]{2,}$"}');
insert into ItemField values ( 4, 1, 4, 'Phone', 4, '', '');
insert into ItemField values ( 5, 1, 5, 'IdProfil', 2, '{"Administrator":1,"User":2}', '');
insert into ItemField values ( 6, 1, 6, 'IsActive', 1, '{"Yes":1,"No":0}', '');

insert into ItemFieldVal values (  1,  1, 0, 0, 'System' );
insert into ItemFieldVal values (  1,  2, 0, 0, 'Administrator' );
insert into ItemFieldVal values (  1,  3, 0, 0, 'main.admin@goHomeDomain.com' );
insert into ItemFieldVal values (  1,  4, 0, 0, '1234567890' );
insert into ItemFieldVal values (  1,  5, 1, 0, '' );
insert into ItemFieldVal values (  1,  6, 1, 0, '' );

insert into Item values      ( 2, 'Sensor', 1, 2, 0, '');
insert into ItemField values ( 8, 2, 1, 'Name', 4, '', '');
insert into ItemField values ( 9, 2, 2, 'IdProfil', 2, '{"Administrator":1,"User":2}', '');
insert into ItemField values (10, 2, 3, 'Record', 1, '{"Yes":1,"No":0}', '');
insert into ItemField values (11, 2, 4, 'IsInternal', 1, '{"Yes":1,"No":0}', '');
insert into ItemField values (12, 2, 5, 'ReadCmd', 4, '', '');
insert into ItemField values (13, 2, 6, 'ReadParam', 4, '', '');
insert into ItemField values (14, 2, 7, 'Interval', 4, '', '');
insert into ItemField values (15, 2, 8, 'IdDataType', 2, '{"Bool":1,"Int":2,"Float":3,"Text":4,"DateTime":5,"FileName":6}', '');
insert into ItemField values (16, 2, 9, 'IsActive', 1, '{"Yes":1,"No":0}', '');

insert into ItemFieldVal values ( 10,  8, 0, 0, '%CPU' );
insert into ItemFieldVal values ( 10,  9, 2, 0, '' );
insert into ItemFieldVal values ( 10, 10, 0, 0, '' );
insert into ItemFieldVal values ( 10, 11, 1, 0, '' );
insert into ItemFieldVal values ( 10, 12, 0, 0, 'CpuUsage' );
insert into ItemFieldVal values ( 10, 13, 0, 0, '' );
insert into ItemFieldVal values ( 10, 14, 0, 0, '1m' );
insert into ItemFieldVal values ( 10, 15, 2, 0, '' );
insert into ItemFieldVal values ( 10, 16, 1, 0, '' );

insert into ItemFieldVal values ( 11,  8, 0, 0, '%Memory' );
insert into ItemFieldVal values ( 11,  9, 2, 0, '' );
insert into ItemFieldVal values ( 11, 10, 0, 0, '' );
insert into ItemFieldVal values ( 11, 11, 1, 0, '' );
insert into ItemFieldVal values ( 11, 12, 0, 0, 'MemoryUsage' );
insert into ItemFieldVal values ( 11, 13, 0, 0, '' );
insert into ItemFieldVal values ( 11, 14, 0, 0, '1m' );
insert into ItemFieldVal values ( 11, 15, 2, 0, '' );
insert into ItemFieldVal values ( 11, 16, 1, 0, '' );

insert into ItemFieldVal values ( 12,  8, 0, 0, 'Alarm' );
insert into ItemFieldVal values ( 12,  9, 2, 0, '' );
insert into ItemFieldVal values ( 12, 10, 0, 0, '' );
insert into ItemFieldVal values ( 12, 11, 1, 0, '' );
insert into ItemFieldVal values ( 12, 12, 0, 0, 'GPIO' );
insert into ItemFieldVal values ( 12, 13, 0, 0, '{"pin":16,"operation":"read","repeat":5,"interval":"50ms","result":"min"}' );
insert into ItemFieldVal values ( 12, 14, 0, 0, '1s' );
insert into ItemFieldVal values ( 12, 15, 2, 0, '' );
insert into ItemFieldVal values ( 12, 16, 1, 0, '' );

insert into Item values      ( 3, 'Actor', 1, 3, 0, '');
insert into ItemField values (17, 3, 1, 'Name', 4, '', '');
insert into ItemField values (18, 3, 2, 'IdProfil', 2, '{"Administrator":1,"User":2}', '');
insert into ItemField values (19, 3, 3, 'IsInternal', 1, '{"Yes":1,"No":0}', '');
insert into ItemField values (20, 3, 4, 'ActCmd', 4, '', '');
insert into ItemField values (21, 3, 5, 'ActParam', 4, '', '');
insert into ItemField values (22, 3, 6, 'IsActive', 1, '{"Yes":1,"No":0}', '');

insert into ItemFieldVal values ( 20, 17, 0, 0, 'Portal' );
insert into ItemFieldVal values ( 20, 18, 2, 0, '' );
insert into ItemFieldVal values ( 20, 19, 1, 0, '' );
insert into ItemFieldVal values ( 20, 20, 1, 0, 'GPIO' );
insert into ItemFieldVal values ( 20, 21, 0, 0, '{"pin":21,"operation":"write","value":1,"duration":"2s"}' );
insert into ItemFieldVal values ( 20, 22, 1, 0, '' );

insert into ItemFieldVal values ( 21, 17, 0, 0, 'Garage' );
insert into ItemFieldVal values ( 21, 18, 2, 0, '' );
insert into ItemFieldVal values ( 21, 19, 1, 0, '' );
insert into ItemFieldVal values ( 21, 20, 1, 0, 'GPIO' );
insert into ItemFieldVal values ( 21, 21, 0, 0, '{"pin":14,"operation":"write","value":1,"duration":"2s"}' );
insert into ItemFieldVal values ( 21, 22, 1, 0, '' );

insert into ItemFieldVal values ( 22, 17, 0, 0, 'GsmReset' );
insert into ItemFieldVal values ( 22, 18, 1, 0, '' );
insert into ItemFieldVal values ( 22, 19, 1, 0, '' );
insert into ItemFieldVal values ( 22, 20, 1, 0, 'GPIO' );
insert into ItemFieldVal values ( 22, 21, 0, 0, '{"pin":6,"operation":"write","value":1,"duration":"100ms"}' );
insert into ItemFieldVal values ( 22, 22, 1, 0, '' );

insert into ItemFieldVal values ( 23, 17, 0, 0, 'SendSMS' );
insert into ItemFieldVal values ( 23, 18, 1, 0, '' );
insert into ItemFieldVal values ( 23, 19, 1, 0, '' );
insert into ItemFieldVal values ( 23, 20, 1, 0, 'SerialATSMS' );
insert into ItemFieldVal values ( 23, 21, 0, 0, '/dev/ttyAMA0' );
insert into ItemFieldVal values ( 23, 22, 1, 0, '' );

insert into Item values      ( 4, 'SensorAct', 1, 4, 2, '');
-- linked Items (if idMasterItem > 0) MUST have a field 'idMasterObj'  : needed to handle the link at object level
insert into ItemField values (23, 4, 1, 'idMasterObj', 2, 'sensor list', '');
insert into ItemField values (24, 4, 2, 'idActor', 2, 'actor list', '');
insert into ItemField values (25, 4, 3, 'Condition', 4, '', '');
insert into ItemField values (26, 4, 4, 'ActorParam', 4, '', '');
insert into ItemField values (27, 4, 5, 'IsActive', 1, '{"Yes":1,"No":0}', '');

insert into ItemFieldVal values ( 30, 23, 12, 0, '' );
insert into ItemFieldVal values ( 30, 24, 23, 0, '' );
insert into ItemFieldVal values ( 30, 25, 0, 0, '@lastVal@ != @prevVal@' );
insert into ItemFieldVal values ( 30, 26, 0, 0, '{"phone":"+123123456789","message":"Alarm @lastVal@"}' );
insert into ItemFieldVal values ( 30, 27, 1, 0, '' );
