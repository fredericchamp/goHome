--
-- Default database initialisation
--
-- One sql stmt per line
-- No multi-line stmt
-- A line stating with -- is ignore
-- An empty line (or only white spaces)  is ignore
-- Any other pattern won't work
--
create table goHome (Perimeter text, Name text, Val text);
create unique index goHome_Uniq on goHome (Perimeter, Name);

create table Item (idItem integer not null primary key, Name text, idProfil integer not null, idMasterItem integer not null, iconeFileName text);

create table ItemField (idField integer not null primary key, idItem integer not null, nOrder integer not null, Name text, idDataType text not null, Label text, Helper text, UniqKey integer, Required integer, RefList text, Regexp text );
create unique index ItemField_Uniq on ItemField (idItem, nOrder);
insert into ItemField values ( 0, 0, 0, 'fieldNone', 0, '', '', 0, 0, '', '' );

create table ItemFieldVal ( idObject integer not null, idField integer not null, Val text);
create unique index ItemFieldVal_PK on ItemFieldVal (idObject, idField);
insert into ItemFieldVal values ( 0, 0, 'objectNone' );

create table HistoSensor (ts datetime not null, idObject integer not null, Val text);
create unique index HistoSensor_PK on HistoSensor (ts, idObject);

create table HistoActor (ts datetime not null, idObject integer not null, idUser int not null, Param text, Res text);
create unique index HistoActor_PK on HistoActor (ts, idObject, idUser);

create table RefValues (name text not null, code text not null, label text);
create unique index RefValues_PK1 on RefValues (name, code);
create unique index RefValues_PK2 on RefValues (name, label);

-- Basic global parameters
insert into goHome values    ( 'Global', 'Version',         '0.2');
insert into goHome values    ( 'Global', 'Email',           'admin@mydomain.net');
insert into goHome values    ( 'Global', 'ServerName',      '-SrvName-');
-- Change to the real server IP
insert into goHome values    ( 'Http',   'https_port',      '443');
insert into goHome values    ( 'Http',   'server_crt',      '/var/goHome/certificats/server.crt.pem');
insert into goHome values    ( 'Http',   'server_key',      '/var/goHome/certificats/server.key.pem');
insert into goHome values    ( 'Http',   'ca_crt',          '/var/goHome/certificats/goHomeCAcert.pem');
insert into goHome values    ( 'Http',   'fileserver_root', '/var/goHome/www');
insert into goHome values    ( 'Http',   'simple_port',     '442');
-- Backup parameters
insert into goHome values    ( 'Backup', 'date/time',       '0 2 * *');
insert into goHome values    ( 'Backup', 'dir',             '/var/goHome/backup');
insert into goHome values    ( 'Backup', 'files_1',         '/usr/bin/rsync -axv /var/goHome/certificats @backupDir@');
insert into goHome values    ( 'Backup', 'files_2',         '/usr/bin/rsync -axv /var/goHome/www @backupDir@');
insert into goHome values    ( 'Backup', 'files_3',         '/usr/bin/rsync -axv /var/goHome/private.sql @backupDir@');
insert into goHome values    ( 'Backup', 'archive',         '/bin/tar cvfah @archiveName@ -C @backupDir@ .');
insert into goHome values    ( 'Backup', 'externalize',     '/usr/bin/curl -s --disable-epsv -T"@archiveName@" -u"$USERNAME:$PASSWORD" "ftp://$SERVER$DESTDIR/"');
insert into goHome values    ( 'Backup', 'cleanup',         '/bin/rm -f @archiveName@');
-- UPnP parameters to allow access to the server if behind a router with NAT (UPnP must be enable on the router) -- update with desire port number
insert into goHome select 'UPnP', '8080', '@localhost@:' || gp.val from goHome gp where gp.perimeter = 'Http' and gp.name = 'https_port';
insert into goHome select 'UPnP', '8079', '@localhost@:' || gp.val from goHome gp where gp.perimeter = 'Http' and gp.name = 'simple_port';
-- Proxy for IP webcam
insert into goHome values ( 'Proxy', '/mini/', 'http://10.0.0.53:8080' );
insert into goHome values ( 'Proxy', '/spica/', 'http://10.0.0.54:8080' );
insert into goHome values ( 'Proxy', '/oneplus3/', 'http://10.0.0.55:8080' );
insert into goHome values ( 'Proxy', '/stairway/', 'http://10.0.0.56:8080' );
insert into goHome values ( 'Proxy', '/oneplus1/', 'http://10.0.0.57:8080' );



-- Proxy for USB local webcam (working with motion running)
insert into goHome values ( 'Proxy', '/sous-sol/', 'http://127.0.0.1:8081' ); -- 8081=mjpeg ; 8080=controls
-- GSM Device reference
-- Disabled : insert into goHome values    ( 'GSM',    'device',          '/dev/ttyAMA0');
--insert into goHome values ( 'Proxy', '/rpi/', 'http://10.0.0.2');


-- YN
insert into RefValues values ('YN', '0', 'No');
insert into RefValues values ('YN', '1', 'Yes');
-- UserProfil
insert into RefValues values ('UserProfil', '1', 'Administrator');
insert into RefValues values ('UserProfil', '2', 'User');
-- DataType / ajouter un type JSON ?
insert into RefValues values ('DataType', '1', 'Bool');
insert into RefValues values ('DataType', '2', 'Int');
insert into RefValues values ('DataType', '3', 'Float');
insert into RefValues values ('DataType', '4', 'Text');
insert into RefValues values ('DataType', '5', 'DateTime');
-- ImgSensorT
insert into RefValues values ('ImgSensorT', '1', 'USB');
insert into RefValues values ('ImgSensorT', '2', 'URL');
-- ImgFormat
insert into RefValues values ('ImgFormat', '1', 'JPEG');
insert into RefValues values ('ImgFormat', '2', 'MJPEG');
insert into RefValues values ('ImgFormat', '3', 'Video');
-- DynParamT / ajouter un type JSON ?
insert into RefValues values ('DynParamT', '0', 'None');
insert into RefValues values ('DynParamT', '1', 'Bool');
insert into RefValues values ('DynParamT', '2', 'Int');
insert into RefValues values ('DynParamT', '3', 'Float');
insert into RefValues values ('DynParamT', '4', 'Text');
insert into RefValues values ('DynParamT', '5', 'DateTime');
insert into RefValues values ('DynParamT', '6', 'URL');
insert into RefValues values ('DynParamT', '7', 'Email');
insert into RefValues values ('DynParamT', '8', 'Tel');
-- email
insert into RefValues values ('email', '-1', '^[a-zA-Z0-9.\-_]*(@)[a-zA-Z0-9.\-_]*(\.)[a-zA-Z]{2,}$');
-- url
insert into RefValues values ('url', '-1', '^[a-zA-Z0-9.\-_/:]*$');
-- phone number
insert into RefValues values ('tel', '-1', '^[0-9]*$');
-- duration
insert into RefValues values ('Duration', '-1', '^[0-9]+(h|m|s|ms)$');


insert into Item values ( 1, 'User',         1, 0, '' );
insert into Item values ( 2, 'Sensor',       1, 0, '' );
insert into Item values ( 3, 'Actor',        1, 0, '' );
insert into Item values ( 4, 'SensorAct',    1, 0, '' );
insert into Item values ( 5, 'Image Sensor', 1, 0, '' );



-- HomeObj definition : User
insert into ItemField select max(f.idField)+1, i.idItem, 1,               'ImgFileName', 4, 'Avatar picture', 'URL for avatar',  0, 0, '',           'url'   from ItemField f, Item i where i.name='User'                         group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'FirstName',   4, 'First name',     'user first name', 0, 1, '',           ''      from ItemField f, Item i where i.name='User' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'LastName',    4, 'Last name',      'user last  name', 0, 1, '',           ''      from ItemField f, Item i where i.name='User' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Email',       4, 'Email',          'email address',   1, 1, '',           'email' from ItemField f, Item i where i.name='User' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Phone',       4, 'Phone Num.',     'user phone num.', 0, 0, '',           'tel'   from ItemField f, Item i where i.name='User' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IdProfil',    2, 'User profil',    'profil for user', 0, 1, 'UserProfil', ''      from ItemField f, Item i where i.name='User' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsActive',    2, 'Active',         'status',          0, 1, 'YN',         ''      from ItemField f, Item i where i.name='User' and f.idItem = i.idItem group by i.idItem;

-- HomeObj definition : Sensor
insert into ItemField select max(f.idField)+1, i.idItem, 1,               'ImgFileName', 4, 'Icone for sensor', 'URL for icone',              0, 1, '',           'url'      from ItemField f, Item i where i.name='Sensor'                         group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Name',        4, 'Name',             'sensor name (unique)',       1, 1, '',           ''         from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IdProfil',    2, 'User profil',      'profil for access',          0, 1, 'UserProfil', ''         from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Record',      2, 'Record readings',  'record readings',            0, 1, 'YN',         ''         from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsInternal',  2, 'Internal',         'is internal fuction',        0, 1, 'YN',         ''         from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'ReadCmd',     4, 'Command',          'read command',               0, 1, '',           ''         from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'ReadParam',   4, 'Parameter',        'read parameters',            0, 0, '',           ''         from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Interval',    4, 'Interval',         'duration between mesure',    0, 0, '',           'Duration' from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IdDataType',  2, 'Data Type',        'data type return by sensor', 0, 1, 'DataType',   ''         from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsVisible',   2, 'Show in GUI',      'Show sensor in GUI',         0, 1, 'YN',         ''         from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsActive',    2, 'Active',           'status',                     0, 1, 'YN',         ''         from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;

-- HomeObj definition : ImageSensor
insert into ItemField select max(f.idField)+1, i.idItem, 1,               'ImgFileName', 4, 'Icone for sensor', 'URL for icone',              0, 1, '',           'url' from ItemField f, Item i where i.name='Image Sensor'                         group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Name',        4, 'Name',             'sensor name (unique)',       1, 1, '',           ''    from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IdProfil',    2, 'User profil',      'profil for access',          0, 1, 'UserProfil', ''    from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Type',        2, 'Sensor type',      'image sensor type',          0, 1, 'ImgSensorT', ''    from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Output',      2, 'Output format',    'stream format',              0, 1, 'ImgFormat',  ''    from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Param',       4, 'Parameter',        'read parameters',            0, 0, '',           ''    from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsVisible',   2, 'Show in GUI',      'Show image sensor in GUI',   0, 1, 'YN',         ''    from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsActive',    2, 'Active',           'status',                     0, 1, 'YN',         ''    from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;

-- HomeObj definition : Actor
insert into ItemField select max(f.idField)+1, i.idItem, 1,               'ImgFileName',  4, 'Icone for actor',     'URL for icone',           0, 1, '',           'url' from ItemField f, Item i where i.name='Actor'                         group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Name',         4, 'Name',                'actor name (unique)',     1, 1, '',           ''    from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IdProfil',     2, 'User profil',         'profil for access',       0, 1, 'UserProfil', ''    from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsInternal',   2, 'Internal',            'is internal fucntion',    0, 1, 'YN',         ''    from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'ActCmd',       4, 'Command ',            'action command',          0, 1, '',           ''    from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'ActParam',     4, 'Parameter',           'action parameters',       0, 0, '',           ''    from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'DynParamType', 2, 'Runtime param. type', 'run time parameter type', 0, 1, 'DynParamT',  ''    from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsVisible',    2, 'Show in GUI',         'Show actor in GUI',       0, 1, 'YN',         ''    from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsActive',     2, 'Active',              'status',                  0, 1, 'YN',         ''    from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;

-- HomeObj definition : SensorAct
insert into ItemField select max(f.idField)+1, i.idItem, 1,               'idMasterObj', 2, 'Master',    'linked sensor',     0, 1, 'SensorList', '' from ItemField f, Item i where i.name='SensorAct'                         group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'idActor',     2, 'Actor',     'trigger actor',     0, 1, 'ActorList',  '' from ItemField f, Item i where i.name='SensorAct' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Condition',   4, 'Condition', 'trigger condition', 0, 0, '',           '' from ItemField f, Item i where i.name='SensorAct' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'ActorParam',  4, 'Parameter', 'action parameters', 0, 0, '',           '' from ItemField f, Item i where i.name='SensorAct' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsActive',    2, 'Active',    'status',            0, 1, 'YN',         '' from ItemField f, Item i where i.name='SensorAct' and f.idItem = i.idItem group by i.idItem;




-- User : Default admin user
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/person.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='User' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'System'            from ItemFieldVal v, ItemField f, Item i where f.name='FirstName'   and i.name='User' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Administrator'     from ItemFieldVal v, ItemField f, Item i where f.name='LastName'    and i.name='User' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'un@goHome.goHome'  from ItemFieldVal v, ItemField f, Item i where f.name='Email'       and i.name='User' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1234567890'        from ItemFieldVal v, ItemField f, Item i where f.name='Phone'       and i.name='User' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='User' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='User' and f.idItem = i.idItem group by f.nOrder;




-- Actor : Exit from server exec (if correctly setup with systemd, server will restart)
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/shutdown.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName'  and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Exit'                from ItemFieldVal v, ItemField f, Item i where f.name='Name'         and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'   and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GoHomeExit'          from ItemFieldVal v, ItemField f, Item i where f.name='ActCmd'       and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, ''                    from ItemFieldVal v, ItemField f, Item i where f.name='ActParam'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '4'                   from ItemFieldVal v, ItemField f, Item i where f.name='DynParamType' and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'    and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;

-- Actor : Portal : gpio write pin 22
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/portail.jpg'  from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName'  and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Portal'              from ItemFieldVal v, ItemField f, Item i where f.name='Name'         and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                   from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'   and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GPIO'                from ItemFieldVal v, ItemField f, Item i where f.name='ActCmd'       and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '{"pin":22,"do":"write","value":"high","duration":2000}' from ItemFieldVal v, ItemField f, Item i where f.name='ActParam'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                   from ItemFieldVal v, ItemField f, Item i where f.name='DynParamType' and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'    and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;

-- Actor : Garage : gpio write pin 23
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/garage.jpg'   from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName'  and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Garage'              from ItemFieldVal v, ItemField f, Item i where f.name='Name'         and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                   from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'   and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GPIO'                from ItemFieldVal v, ItemField f, Item i where f.name='ActCmd'       and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '{"pin":23,"do":"write","value":"high","duration":2000}' from ItemFieldVal v, ItemField f, Item i where f.name='ActParam'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                   from ItemFieldVal v, ItemField f, Item i where f.name='DynParamType' and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'    and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;

-- Disabled : -- Actor : Hard reset Gsm module : gpio write pin 18
-- Disabled : insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/gsmreset.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName'  and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GsmReset'            from ItemFieldVal v, ItemField f, Item i where f.name='Name'         and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'   and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GPIO'                from ItemFieldVal v, ItemField f, Item i where f.name='ActCmd'       and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '{"pin":18,"do":"write","value":"high","duration":1000}' from ItemFieldVal v, ItemField f, Item i where f.name='ActParam'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                   from ItemFieldVal v, ItemField f, Item i where f.name='DynParamType' and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'    and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : -- GSM actor reference
-- Disabled : insert into goHome select 'GSM', 'resetActorId', max(v.idObject) from ItemFieldVal v;

-- Disabled : -- Actor : Switch On/Off Gsm module : gpio write pin 17
-- Disabled : insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/gsmonoff.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName'  and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GsmOnOff'            from ItemFieldVal v, ItemField f, Item i where f.name='Name'         and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'   and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GPIO'                from ItemFieldVal v, ItemField f, Item i where f.name='ActCmd'       and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '{"pin":17,"do":"write","value":"high","duration":1000}' from ItemFieldVal v, ItemField f, Item i where f.name='ActParam'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                   from ItemFieldVal v, ItemField f, Item i where f.name='DynParamType' and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'    and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : -- GSM actor reference
-- Disabled : insert into goHome select 'GSM', 'onOffActorId', max(v.idObject) from ItemFieldVal v;

-- Disabled : -- Actor : Restart Gsm module : full reinit
-- Disabled : insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/gsmreset.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName'  and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GsmRestart'          from ItemFieldVal v, ItemField f, Item i where f.name='Name'         and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'   and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GsmRestart'          from ItemFieldVal v, ItemField f, Item i where f.name='ActCmd'       and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, ''                    from ItemFieldVal v, ItemField f, Item i where f.name='ActParam'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                   from ItemFieldVal v, ItemField f, Item i where f.name='DynParamType' and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'    and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;

-- Disabled : -- Actor : SendSMS using Gsm module
-- Disabled : insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/gsmsms.png'   from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName'  and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'SendSMS'             from ItemFieldVal v, ItemField f, Item i where f.name='Name'         and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                   from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'   and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'SendSMS'             from ItemFieldVal v, ItemField f, Item i where f.name='ActCmd'       and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '/dev/ttyAMA0'        from ItemFieldVal v, ItemField f, Item i where f.name='ActParam'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '4'                   from ItemFieldVal v, ItemField f, Item i where f.name='DynParamType' and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'    and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
-- Actor : SendSMS using shell script (calling HTTP gateway)
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/sms-blue.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName'  and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'SendSMS'             from ItemFieldVal v, ItemField f, Item i where f.name='Name'         and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                   from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'   and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'SendSMS'             from ItemFieldVal v, ItemField f, Item i where f.name='ActCmd'       and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '192.168.43.1:1116'   from ItemFieldVal v, ItemField f, Item i where f.name='ActParam'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '4'                   from ItemFieldVal v, ItemField f, Item i where f.name='DynParamType' and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'    and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                   from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;






-- Sensor : % memory usage
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/perf.png'   from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '%Memory'           from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                 from ItemFieldVal v, ItemField f, Item i where f.name='Record'      and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'MemoryUsage'       from ItemFieldVal v, ItemField f, Item i where f.name='ReadCmd'     and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, ''                  from ItemFieldVal v, ItemField f, Item i where f.name='ReadParam'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '5m'                from ItemFieldVal v, ItemField f, Item i where f.name='Interval'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdDataType'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;

-- Sensor : % cpu usage
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/perf.png'   from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '%CPU'              from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                 from ItemFieldVal v, ItemField f, Item i where f.name='Record'      and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'CpuUsage'          from ItemFieldVal v, ItemField f, Item i where f.name='ReadCmd'     and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, ''                  from ItemFieldVal v, ItemField f, Item i where f.name='ReadParam'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '5m'                from ItemFieldVal v, ItemField f, Item i where f.name='Interval'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdDataType'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;

-- Sensor : Is Alarm on ? : read gpio pin 27
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/alarm.png'  from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Alarm'             from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                 from ItemFieldVal v, ItemField f, Item i where f.name='Record'      and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GPIO'              from ItemFieldVal v, ItemField f, Item i where f.name='ReadCmd'     and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '{"pin":27,"do":"read","repeat":5,"interval":50,"op":"min"}' from ItemFieldVal v, ItemField f, Item i where f.name='ReadParam'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1s'                from ItemFieldVal v, ItemField f, Item i where f.name='Interval'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdDataType'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- SensorAct : is Alarm in On (read 0) and alarm was off (@lastVal@ < @prevVal@) => sens SMS "Alarm"
insert into ItemFieldVal select max(v.idObject)+1, f.idField, mv.idObject      from ItemFieldVal mv, ItemField mf, Item mi, ItemFieldVal v, ItemField f, Item i where f.name='idMasterObj' and i.name='SensorAct' and f.idItem = i.idItem and mv.idfield = mv.idfield and mv.val='Alarm'   and mf.name='Name' and mf.idItem = mi.idItem and mi.name = 'Sensor' group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, av.idObject      from ItemFieldVal av, ItemField af, Item ai, ItemFieldVal v, ItemField f, Item i where f.name='idActor'     and i.name='SensorAct' and f.idItem = i.idItem and av.idfield = av.idfield and av.val='SendSMS' and af.name='Name' and af.idItem = ai.idItem and ai.name = 'Actor'  group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '@lastVal@ < @prevVal@' from ItemFieldVal v, ItemField f, Item i where f.name='Condition'   and i.name='SensorAct' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0123123456789 "Alarm maison (@lastVal@)"' from ItemFieldVal v, ItemField f, Item i where f.name='ActorParam'  and i.name='SensorAct' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'              from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='SensorAct' and f.idItem = i.idItem group by f.nOrder;
-- SensorAct : is Alarm in Off (read 1) and alarm was on (@lastVal@ > @prevVal@) => sens SMS "Alarm end"
insert into ItemFieldVal select max(v.idObject)+1, f.idField, mv.idObject      from ItemFieldVal mv, ItemField mf, Item mi, ItemFieldVal v, ItemField f, Item i where f.name='idMasterObj' and i.name='SensorAct' and f.idItem = i.idItem and mv.idfield = mv.idfield and mv.val='Alarm'   and mf.name='Name' and mf.idItem = mi.idItem and mi.name = 'Sensor' group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, av.idObject      from ItemFieldVal av, ItemField af, Item ai, ItemFieldVal v, ItemField f, Item i where f.name='idActor'     and i.name='SensorAct' and f.idItem = i.idItem and av.idfield = av.idfield and av.val='SendSMS' and af.name='Name' and af.idItem = ai.idItem and ai.name = 'Actor'  group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '@lastVal@ > @prevVal@' from ItemFieldVal v, ItemField f, Item i where f.name='Condition'   and i.name='SensorAct' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0123123456789 "Fin alarm (@lastVal@)"' from ItemFieldVal v, ItemField f, Item i where f.name='ActorParam'  and i.name='SensorAct' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'              from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='SensorAct' and f.idItem = i.idItem group by f.nOrder;



-- Image Sensor : sensor IP webcam Entree
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/video.png'  from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Entree'            from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='Type'        and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='Output'      and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '/spica/video'     from ItemFieldVal v, ItemField f, Item i where f.name='Param'       and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'   and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;


-- Sensor : Get external IP : UPnP read from router
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/atsign.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'ExternalIP'        from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                 from ItemFieldVal v, ItemField f, Item i where f.name='Record'      and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GetExternalIP'     from ItemFieldVal v, ItemField f, Item i where f.name='ReadCmd'     and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, ''                  from ItemFieldVal v, ItemField f, Item i where f.name='ReadParam'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '60m'               from ItemFieldVal v, ItemField f, Item i where f.name='Interval'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '4'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdDataType'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;


-- Disabled : -- Sensor : Is Gsm module on ? : send AT\r to module 
-- Disabled : insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/gsmisup.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GsmIsUp'            from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                  from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                  from ItemFieldVal v, ItemField f, Item i where f.name='Record'      and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                  from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GsmIsUp'            from ItemFieldVal v, ItemField f, Item i where f.name='ReadCmd'     and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, ''                   from ItemFieldVal v, ItemField f, Item i where f.name='ReadParam'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '60m'                from ItemFieldVal v, ItemField f, Item i where f.name='Interval'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                  from ItemFieldVal v, ItemField f, Item i where f.name='IdDataType'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                  from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                  from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : -- SensorAct : Gsm module not responding "OK\r" => restart module
-- Disabled : insert into ItemFieldVal select max(v.idObject)+1, f.idField, mv.idObject      from ItemFieldVal mv, ItemField mf, Item mi, ItemFieldVal v, ItemField f, Item i where f.name='idMasterObj' and i.name='SensorAct' and f.idItem = i.idItem and mv.idfield = mv.idfield and mv.val='GsmIsUp'    and mf.name='Name' and mf.idItem = mi.idItem and mi.name = 'Sensor' group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, av.idObject      from ItemFieldVal av, ItemField af, Item ai, ItemFieldVal v, ItemField f, Item i where f.name='idActor'     and i.name='SensorAct' and f.idItem = i.idItem and av.idfield = av.idfield and av.val='GsmRestart' and af.name='Name' and af.idItem = ai.idItem and ai.name = 'Actor'  group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '@lastVal@ != 1' from ItemFieldVal v, ItemField f, Item i where f.name='Condition'   and i.name='SensorAct' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, ''               from ItemFieldVal v, ItemField f, Item i where f.name='ActorParam'  and i.name='SensorAct' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'              from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='SensorAct' and f.idItem = i.idItem group by f.nOrder;


-- Sensor : Take snapshot from USB webcam using motion 
-- Disabled : insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/camera.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'AlarmSnap'         from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                 from ItemFieldVal v, ItemField f, Item i where f.name='Record'      and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '/usr/local/bin/newSnap.sh' from ItemFieldVal v, ItemField f, Item i where f.name='ReadCmd'     and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'localhost:8080 0 alarm.jpg -rotate 88 -crop 160x100+560+410' from ItemFieldVal v, ItemField f, Item i where f.name='ReadParam'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '30s'                from ItemFieldVal v, ItemField f, Item i where f.name='Interval'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '4'                  from ItemFieldVal v, ItemField f, Item i where f.name='IdDataType'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                  from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                  from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;

-- Image Sensor : Add USB webcam
-- Disabled : insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/video.png'  from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'AlarmVideo'        from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='Type'        and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='Output'      and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '/sous-sol/video0'  from ItemFieldVal v, ItemField f, Item i where f.name='Param'       and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'   and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;

-- Image Sensor : Add USB webcam : view snapshot
-- Disabled : insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/alarm.png'  from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Alarm'             from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='Type'        and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='Output'      and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '/capture/alarm.jpg' from ItemFieldVal v, ItemField f, Item i where f.name='Param'       and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsVisible'   and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
-- Disabled : insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;


--List param
--P1 tel
--P2 text
--P3 mail
--
--Link param actor
--idActor - idParam - iOrder
--
--
--insert into Item values ( 6, 'Parameter', 1, 0, '' );
--insert into Item values ( 7, 'ParamLink', 1, 0, '' );
--
--
---- HomeObj definition : Parameter
--insert into ItemField select max(f.idField)+1, i.idItem, 1,               'Label',        4, 'Display label',       'label for input form',    0, 1, '',           '' from ItemField f, Item i where i.name='Parameter'                         group by i.idItem;
--insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Name',         4, 'Name',                'parameter name (unique)', 1, 1, '',           ''    from ItemField f, Item i where i.name='Parameter' and f.idItem = i.idItem group by i.idItem;
--insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IdProfil',     2, 'User profil',         'profil for access',       0, 1, 'UserProfil', ''    from ItemField f, Item i where i.name='Parameter' and f.idItem = i.idItem group by i.idItem;
--insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'DynParamType', 2, 'Runtime param. type', 'run time parameter type', 0, 1, 'DynParamT',  ''    from ItemField f, Item i where i.name='Parameter' and f.idItem = i.idItem group by i.idItem;
--
---- HomeObj definition : ParamLink
--insert into ItemField select max(f.idField)+1, i.idItem, 1,               'idMasterObj', 2, 'Actor',    'linked actor', 0, 1, 'ActorList', '' from ItemField f, Item i where i.name='ParamLink'                         group by i.idItem;
--insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'idParam',     2, 'Paramter', 'linked param', 0, 1, 'ParamList', '' from ItemField f, Item i where i.name='ParamLink' and f.idItem = i.idItem group by i.idItem;
--insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'nOrder',      2, 'Position', 'order',        0, 0, '',          '' from ItemField f, Item i where i.name='ParamLink' and f.idItem = i.idItem group by i.idItem;
--insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsActive',    2, 'Active',   'status',       0, 1, 'YN',        '' from ItemField f, Item i where i.name='ParamLink' and f.idItem = i.idItem group by i.idItem;
