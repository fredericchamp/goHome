--
-- Default database initialisation
--
-- One sql stmt per line
-- No multi-line stmt
-- A line stating with -- is ignore
-- An empty line (or only white spaces)  is ignore
-- Any other pattern wont work
--
create table goHome (Perimeter text, Name text, Val text);
create unique index goHome_Uniq on goHome (Perimeter, Name);

create table Item (idItem integer not null primary key, Name text, idProfil integer not null, idItemType integer not null, idMasterItem integer not null, iconeFileName text);

create table ItemField (idField integer not null primary key, idItem integer not null, nOrder integer not null, Name text, idDataType not null, Rules text );
create unique index ItemField_Uniq on ItemField (idItem, nOrder);
insert into ItemField values ( 0, 0, 0, 'fieldNone', 0, '');

create table ItemFieldVal ( idObject integer not null, idField integer not null, Val text);
create unique index ItemFieldVal_PK on ItemFieldVal (idObject, idField);
insert into ItemFieldVal values ( 0, 0, 'objectNone' );

create table HistoSensor (ts datetime not null, idObject integer not null, Val text);
create unique index HistoSensor_PK on HistoSensor (ts, idObject);

create table HistoActor (ts datetime not null, idObject integer not null, idUser int not null, Param text, Res text);
create unique index HistoActor_PK on HistoActor (ts, idObject, idUser);

create table RefValues (name text not null, code text not null, label text);
create unique index RefValues_PK on RefValues (name, code);
create unique index RefValues_PK on RefValues (name, label);

-- YN				[{id:1,label:"Yes"},{id:0,label:"No"}]
insert into RefValues values ('YN', '0', 'No');
insert into RefValues values ('YN', '1', 'Yes');
-- UserProfil		[{id:1,label:"Administrator"},{id:2,label:"User"}]
insert into RefValues values ('UserProfil', '0', 'Profil None');
insert into RefValues values ('UserProfil', '1', 'Administrator');
insert into RefValues values ('UserProfil', '2', 'User');
-- DataType		[{id:1,label:"Bool"},{id:2,label:"Int"},{id:3,label:"Float"},{id:4,label:"Text"},{id:5,label:"DateTime"},{id:6,label:"URL"}], ............... ajouter un type JSON ?
insert into RefValues values ('DataType', '1', 'Bool');
insert into RefValues values ('DataType', '2', 'Int');
insert into RefValues values ('DataType', '3', 'Float');
insert into RefValues values ('DataType', '4', 'Text');
insert into RefValues values ('DataType', '5', 'DateTime');
insert into RefValues values ('DataType', '6', 'URL');
-- email			^[[:alnum:].\-_]*@[[:alnum:].\-_]*[.][[:alpha:]]{2,}$
insert into RefValues values ('Email', '-1', '^[[:alnum:].\-_]*@[[:alnum:].\-_]*[.][[:alpha:]]{2,}$');
-- url				juste exchude les car speciaux ?
insert into RefValues values ('URL', '-1', '^[[:alnum:].\-_/]$');
-- duration			^[[:num:]]+[dhms][s]{0,1}$
insert into RefValues values ('Duration', '-1', '^[[:num:]]+[dhms][s]{0,1}$');
-- SensorList		select ...
insert into RefValues values ('SensorList', '-2', 'select idObject, Val from ItemFieldVal where ...');
-- ActorList			select ...
insert into RefValues values ('ActorList', '-2', 'select idObject, Val from ItemFieldVal where ...');
-- ImgSensorT		[{id:1,label:"USB attach"},{id:2,label:"URL"}]
insert into RefValues values ('ImgSensorT', '1', 'USB');
insert into RefValues values ('ImgSensorT', '2', 'URL');
-- ImgFormat 		[{id:1,label:"JPEG"},{id:2,label:"MJPEG"},{id:2,label:"Video"}]
insert into RefValues values ('ImgFormat', '1', 'JPEG');
insert into RefValues values ('ImgFormat', '2', 'MJPEG');
insert into RefValues values ('ImgFormat', '3', 'VIDEO');
-- DynParamT		[{id:0,label:"None"},{id:1,label:"Bool"},{id:2,label:"Int"},{id:3,label:"Float"},{id:4,label:"Text"},{id:5,label:"DateTime"},{id:6,label:"URL"}] ............... ajouter un type JSON ?
insert into RefValues values ('DynParamT', '0', 'None');
insert into RefValues values ('DynParamT', '1', 'Bool');
insert into RefValues values ('DynParamT', '2', 'Int');
insert into RefValues values ('DynParamT', '3', 'Float');
insert into RefValues values ('DynParamT', '4', 'Text');
insert into RefValues values ('DynParamT', '5', 'DateTime');
insert into RefValues values ('DynParamT', '6', 'URL');


insert into goHome values    ( 'Global', 'Version',         '0.1');
insert into goHome values    ( 'Global', 'ApiVersion',      '0.1');
insert into goHome values    ( 'Global', 'Email',           'admin@goHomeDomain.net');
insert into goHome values    ( 'Global', 'UserItemId',      '1');
insert into goHome values    ( 'Http',   'server_name',     'localhost');
insert into goHome values    ( 'Http',   'https_port',      '5100');
insert into goHome values    ( 'Http',   'server_crt',      '/var/goHome/certificats/server.crt.pem');
insert into goHome values    ( 'Http',   'server_key',      '/var/goHome/certificats/server.key.pem');
insert into goHome values    ( 'Http',   'ca_crt',          '/var/goHome/certificats/goHomeCAcert.pem');
insert into goHome values    ( 'Http',   'fileserver_root', '/var/goHome/www');


insert into Item select g.val, 'User', 1, 1, 0, '' from goHome g where g.perimeter='Global' and g.name ='UserItemId';
insert into ItemField select max(f.idField)+1, i.idItem, 1,               'ImgFileName', 4, 'regexp:url'            from ItemField f, Item i where i.name='User'                         group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'FirstName',   4, ''                      from ItemField f, Item i where i.name='User' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'LastName',    4, ''                      from ItemField f, Item i where i.name='User' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Email',       4, 'uniq:1,regexp:email'   from ItemField f, Item i where i.name='User' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Phone',       4, ''                      from ItemField f, Item i where i.name='User' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IdProfil',    2, 'selectList:UserProfil' from ItemField f, Item i where i.name='User' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsActive',    1, 'selectList:YN'         from ItemField f, Item i where i.name='User' and f.idItem = i.idItem group by i.idItem;

insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/cross.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='User' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'System'           from ItemFieldVal v, ItemField f, Item i where f.name='FirstName'   and i.name='User' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Administrator'    from ItemFieldVal v, ItemField f, Item i where f.name='LastName'    and i.name='User' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'admin@goHome.com' from ItemFieldVal v, ItemField f, Item i where f.name='Email'       and i.name='User' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1234567890'       from ItemFieldVal v, ItemField f, Item i where f.name='Phone'       and i.name='User' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='User' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='User' and f.idItem = i.idItem group by f.nOrder;


insert into Item select max(i.idItem)+1, 'Sensor', 1, 2, 0, '' from Item i;
insert into ItemField select max(f.idField)+1, i.idItem, 1,               'ImgFileName', 4, 'regexp:url'            from ItemField f, Item i where i.name='Sensor'                         group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Name',        4, 'uniq:1'                from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IdProfil',    2, 'selectList:UserProfil' from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Record',      1, 'selectList:YN'         from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsInternal',  1, 'selectList:YN'         from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'ReadCmd',     4, ''                      from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'ReadParam',   4, ''                      from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Interval',    4, 'regexp:duration'       from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IdDataType',  1, 'selectList:DataType'   from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsActive',    1, 'selectList:YN'         from ItemField f, Item i where i.name='Sensor' and f.idItem = i.idItem group by i.idItem;

insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/cross.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '%CPU'             from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                from ItemFieldVal v, ItemField f, Item i where f.name='Record'      and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'CpuUsage'         from ItemFieldVal v, ItemField f, Item i where f.name='ReadCmd'     and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, ''                 from ItemFieldVal v, ItemField f, Item i where f.name='ReadParam'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1m'               from ItemFieldVal v, ItemField f, Item i where f.name='Interval'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                from ItemFieldVal v, ItemField f, Item i where f.name='IdDataType'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;

insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/cross.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '%Memory'          from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                from ItemFieldVal v, ItemField f, Item i where f.name='Record'      and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'MemoryUsage'      from ItemFieldVal v, ItemField f, Item i where f.name='ReadCmd'     and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, ''                 from ItemFieldVal v, ItemField f, Item i where f.name='ReadParam'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1m'               from ItemFieldVal v, ItemField f, Item i where f.name='Interval'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                from ItemFieldVal v, ItemField f, Item i where f.name='IdDataType'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;

insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/cross.png'                             from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Alarm'                                        from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                                            from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                                            from ItemFieldVal v, ItemField f, Item i where f.name='Record'      and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                                            from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GPIO'                                         from ItemFieldVal v, ItemField f, Item i where f.name='ReadCmd'     and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'pin:16,do:read,repeat:5,interval:50ms,op:min' from ItemFieldVal v, ItemField f, Item i where f.name='ReadParam'   and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1s'                                           from ItemFieldVal v, ItemField f, Item i where f.name='Interval'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                                            from ItemFieldVal v, ItemField f, Item i where f.name='IdDataType'  and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                                            from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Sensor' and f.idItem = i.idItem group by f.nOrder;


insert into Item select max(i.idItem)+1, 'Actor', 1, 3, 0, '' from Item i;
insert into ItemField select max(f.idField)+1, i.idItem, 1,               'ImgFileName',  4, 'regexp:url'            from ItemField f, Item i where i.name='Actor'                         group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Name',         4, 'uniq:1'                from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IdProfil',     2, 'selectList:UserProfil' from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsInternal',   1, 'selectList:YN'         from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'ActCmd',       4, ''                      from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'ActParam',     4, ''                      from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'DynParamType', 2, 'selectList:DynParamT'  from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsActive',     1, 'selectList:YN'         from ItemField f, Item i where i.name='Actor' and f.idItem = i.idItem group by i.idItem;

insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/portail.jpg'                  from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName'  and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Portal'                              from ItemFieldVal v, ItemField f, Item i where f.name='Name'         and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                                   from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                                   from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'   and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GPIO'                                from ItemFieldVal v, ItemField f, Item i where f.name='ActCmd'       and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'pin:21,do:write,value:1,duration:2s' from ItemFieldVal v, ItemField f, Item i where f.name='ActParam'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                                   from ItemFieldVal v, ItemField f, Item i where f.name='DynParamType' and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                                   from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;

insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/garage.jpg'                   from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName'  and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Garage'                              from ItemFieldVal v, ItemField f, Item i where f.name='Name'         and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                                   from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                                   from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'   and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GPIO'                                from ItemFieldVal v, ItemField f, Item i where f.name='ActCmd'       and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'pin:14,do:write,value:1,duration:2s' from ItemFieldVal v, ItemField f, Item i where f.name='ActParam'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                                   from ItemFieldVal v, ItemField f, Item i where f.name='DynParamType' and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                                   from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;

insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/gsmreset.png'                   from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName'  and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GsmReset'                              from ItemFieldVal v, ItemField f, Item i where f.name='Name'         and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                                     from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                                     from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'   and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'GPIO'                                  from ItemFieldVal v, ItemField f, Item i where f.name='ActCmd'       and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'pin:6,do:write,value:1,duration:100ms' from ItemFieldVal v, ItemField f, Item i where f.name='ActParam'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '0'                                     from ItemFieldVal v, ItemField f, Item i where f.name='DynParamType' and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                                     from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;

insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/gsmsms.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName'  and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'SendSMS'           from ItemFieldVal v, ItemField f, Item i where f.name='Name'         and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsInternal'   and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'SerialATSMS'       from ItemFieldVal v, ItemField f, Item i where f.name='ActCmd'       and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '/dev/ttyAMA0'      from ItemFieldVal v, ItemField f, Item i where f.name='ActParam'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '4'                 from ItemFieldVal v, ItemField f, Item i where f.name='DynParamType' and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'     and i.name='Actor' and f.idItem = i.idItem group by f.nOrder;


insert into Item select max(i.idItem)+1, 'SensorAct', 1, 4, 0, '' from Item i;
-- linked Items (if idMasterItem > 0) MUST have a field 'idMasterObj'  : needed to handle the link at object level
insert into ItemField select max(f.idField)+1, i.idItem, 1,               'idMasterObj', 2, 'selectList:SensorList' from ItemField f, Item i where i.name='SensorAct'                         group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'idActor',     2, 'selectList:ActorList'  from ItemField f, Item i where i.name='SensorAct' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Condition',   4, ''                      from ItemField f, Item i where i.name='SensorAct' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'ActorParam',  4, ''                      from ItemField f, Item i where i.name='SensorAct' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsActive',    1, 'selectList:YN'         from ItemField f, Item i where i.name='SensorAct' and f.idItem = i.idItem group by i.idItem;

insert into ItemFieldVal select max(v.idObject)+1, f.idField, mv.idObject                                   from ItemFieldVal mv, ItemField mf, Item mi, ItemFieldVal v, ItemField f, Item i where f.name='idMasterObj' and i.name='SensorAct' and f.idItem = i.idItem and mv.idfield = mv.idfield and mv.val='Alarm' and mf.name='Name' and mf.idItem = mi.idItem and mi.idItemType=2 group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, av.idObject                                   from ItemFieldVal av, ItemField af, Item ai, ItemFieldVal v, ItemField f, Item i where f.name='idActor' and i.name='SensorAct' and f.idItem = i.idItem and av.idfield = av.idfield and av.val='SendSMS' and af.name='Name' and af.idItem = ai.idItem and ai.idItemType=3 group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '@lastVal@ != @prevVal@'                      from ItemFieldVal v, ItemField f, Item i where f.name='Condition'   and i.name='SensorAct' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'phone:+123123456789,message:Alarm @lastVal@' from ItemFieldVal v, ItemField f, Item i where f.name='ActorParam'  and i.name='SensorAct' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                                           from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='SensorAct' and f.idItem = i.idItem group by f.nOrder;


insert into Item select max(i.idItem)+1, 'Image Sensor', 1, 5, 0, '' from Item i;
insert into ItemField select max(f.idField)+1, i.idItem, 1,               'ImgFileName', 4, 'regexp:url'            from ItemField f, Item i where i.name='Image Sensor'                         group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Name',        4, 'uniq:1'                from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IdProfil',    2, 'selectList:UserProfil' from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Type',        2, 'selectList:ImgSensorT' from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Output',      2, 'selectList:ImgFormat'  from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'Param',       4, ''                      from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;
insert into ItemField select max(f.idField)+1, i.idItem, max(f.nOrder)+1, 'IsActive',    1, 'selectList:YN'         from ItemField f, Item i where i.name='Image Sensor' and f.idItem = i.idItem group by i.idItem;

-- Add sensor IP webcam Nexus
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/video.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Entree'        from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='Type'        and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='Output'      and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '/nexus/video'      from ItemFieldVal v, ItemField f, Item i where f.name='Param'       and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;

-- Add sensor photo Nexus
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/camera.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'EntreeX'           from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='Type'        and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='Output'      and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '/nexus/photo.jpg'  from ItemFieldVal v, ItemField f, Item i where f.name='Param'       and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;

-- Add USB webcam
insert into ItemFieldVal select max(v.idObject)+1, f.idField, 'images/video.png' from ItemFieldVal v, ItemField f, Item i where f.name='ImgFileName' and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, 'Sous-sol'          from ItemFieldVal v, ItemField f, Item i where f.name='Name'        and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='IdProfil'    and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='Type'        and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '2'                 from ItemFieldVal v, ItemField f, Item i where f.name='Output'      and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '/dev/video0'       from ItemFieldVal v, ItemField f, Item i where f.name='Param'       and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;
insert into ItemFieldVal select max(v.idObject)  , f.idField, '1'                 from ItemFieldVal v, ItemField f, Item i where f.name='IsActive'    and i.name='Image Sensor' and f.idItem = i.idItem group by f.nOrder;

