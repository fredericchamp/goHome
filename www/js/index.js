
const cReadRefList     =   1;
const cReadItem        =   2;
const cReadUsers       =  10;
const cReadCurrentUser =  11;
const cReadActors      =  20;
const cReadImgSensor   =  30;
const cReadSensor      =  40;
const cReadSensorVal   =  41;
const cReadSensorAct   =  50;
const cTriggerActor    = 100;
const cSaveObject      = 200;

const DBTypeNone       = 0;
const DBTypeBool       = 1;
const DBTypeInt        = 2;
const DBTypeFloat      = 3;
const DBTypeText       = 4;
const DBTypeDateTime   = 5;
const DBTypeURL        = 6;


var fc = new Object({
    refList: null,
    itemList: null,
    currentUser: null,
    userList: null,
    actorList: null,
    sensorList: null,
    sensorActList : null,
    imgSensorList: null,
    imgSensorSrc: '',
});

// TODO create methods on fc to handle xxxList update with relevant GUI update

// -----------------------------------------------------------------------------------------------------------------------------------------

function getReadForItemId( itemId ) {
    switch (itemId) {
    case 1: return cReadUsers;
    case 2: return cReadSensor;
    case 3: return cReadActors;
    case 4: return cReadSensorAct;
    case 5: return cReadImgSensor;
    default:return 0;
    }
}

function getItemIdForRead( action ) {
    switch (action) {
    case cReadUsers:     return 1;
    case cReadSensor:    return 2;
    case cReadActors:    return 3;
    case cReadSensorAct: return 4;
    case cReadImgSensor: return 5;
    default :            return 0;
    }
}

function objectListForItemId(itemId) {
    switch (itemId) {
    case 1: return fc.userList;
    case 2: return fc.sensorList;
    case 3: return fc.actorList;
    case 4: return fc.sensorActList;
    case 5: return fc.imgSensorList;
    default:return [];
    }
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function htmlEncode(str) {
    var result = "";
    var str = (arguments.length===1) ? str : this;
    for(var i=0; i<str.length; i++) {
        var chrcode = str.charCodeAt(i);
        result+=(chrcode>128) ? "&#"+chrcode+";" : str.substr(i,1)
    }
    return result;
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function formatUnixTs(unixts) {
    if ( unixts == 0 ) {
        return '/';
    }
    var now = new Date();
    var dt = new Date(unixts);
    if ( dt.getDate() != now.getDate() )
        return dt.getFullYear() + '-' + dt.getMonth() + '-' + dt.getDate();
    return ('0'+dt.getHours()).slice(-2) + ':' + ('0'+dt.getMinutes()).slice(-2) + ':' + ('0'+dt.getSeconds()).slice(-2);
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function refListGetVal( refList, lstname, code ) {
    if ( refList == null ) return '';
    var i = 0;
    for (i = 0; i < refList[lstname].length; i++) {
        if ( refList[lstname][i].Code == code ) {
            return refList[lstname][i].Label;
        }
    }
    return '';
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function refListFromNames( listName, objList ) {
    var refList = new Array();
    if ( objList == null ) return refList;
    var nameIdx = 0;
    var i = 0;
    for (i = 0; i < objList[0].Fields.length; i++) {
        if ( objList[0].Fields[i].Name == 'Name' ) {
            nameIdx = i;
            break;
        }
    }
    for (i = 0; i < objList.length; i++) {
        refList.push( {Name: listName, Code:objList[i].Values[0].IdObject, Label:objList[i].Values[nameIdx].Val} );
    }
    return refList;
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function getItemById( itemList, itemId ) {
    if ( itemList == null ) return null;
    var i = 0;
    for (i = 0; i < itemList.length; i++) {
        if ( itemList[i].IdItem == itemId ) {
            return itemList[i];
        }
    }
    return null;
}

function getItemNameById( itemList, itemId ) {
    var item = getItemById( itemList, itemId )
    if ( item == null ) {
        return "IdItem_" + itemId;
    }
    return item.Name;
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function newHomeObject(fields) {
    if ( fields == null ) return null;
    var newObj = new Object();
    newObj.Fields = fields;
    newObj.Values = new Array();
    var i;
    for (i = 0; i < newObj.Fields.length; i++) {
        newObj.Values.push( {IdObject:0, IdField:newObj.Fields[i].IdField, Val:''} );
    }
    return newObj;
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function getObjById(objs,id){
    if ( objs == null ) return null;
    var i = 0;
    for (i = 0; i < objs.length; i++) {
        if (objs[i].Values[0].IdObject == id) {
            return objs[i];
        }
    }
    return null;
}

function getObjVal(obj,key){
    if ( obj == null ) return '';
    var i = 0;
    var fieldId=0;
    for (i = 0; i < obj.Fields.length; i++) {
        if (obj.Fields[i].Name == key) {
            fieldId = obj.Fields[i].IdField;
            break;
        }
    }
    if ( fieldId==0 ) {
        return '';
    }
    for (i = 0; i < obj.Values.length; i++) {
        if (obj.Values[i].IdField == fieldId) {
            return obj.Values[i].Val;
        }
    }
    return '';
}

function searchObjByVal(objLst,itemId,objId,key,val){
    if ( objLst == null ) return null;
    var x=0;
    for (x = 0; x < objLst.length; x++) {
        if ( objLst[x].Fields[0].IdItem != itemId ) {
            continue;
        }
        if ( objLst[x].Values[0].IdObject == objId ) {
            continue;
        }
        if ( getObjVal(objLst[x],key) == val) {
            return objLst[x];
        }
    }
    return null;
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function getObjectName(obj) {
    if ( obj == null ) return 'null';
    switch (obj.Fields[0].IdItem) {
    case 1: // User
        return getObjVal(obj,"FirstName") + ' ' + getObjVal(obj,"LastName") + ' (' + getObjVal(obj,"Email") + ')';
    case 2: // Sensor;
    case 3: // Actors;
    case 5: // ImgSensor;
        return getObjVal(obj,"Name");
    case 4: // SensorAct
        return getObjVal(getObjById(fc.sensorList,getObjVal(obj,"idMasterObj")),"Name") + " to " +
               getObjVal(getObjById(fc.actorList,getObjVal(obj,"idActor")),"Name") + " on '" +
               htmlEncode(getObjVal(obj,"Condition")) + "'";
    default:
        return 'Object_' + obj.Values[0].IdObject ;
    }
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function readObjectLst( action ) {
   var itemId = getItemIdForRead( action );
    if ( itemId <= 0 ) {
        // TODO ERROR unknown itemId or action
        return;
    }
    callServer(action, { command:'ReadObject', itemid:itemId, objectid:0, startts:0, endts:0, jsonparam:'' });
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function callServer(action,cmde){
    $.post("/api", { command:$.toJSON(cmde) }, function(data, status){
        switch (action) {
        case cReadRefList:
            // TODO check if data != '{"error":"....."}'
            if ( fc.refList == null ) {
                fc.refList = new Object();
            }
            var lstAll = $.parseJSON(data);
            var curName = lstAll[0].Name;
            var oneList = new Array();
            for ( i = 0; i < lstAll.length; i++) {
                if ( curName == lstAll[i].Name ) {
                    oneList.push(lstAll[i]);
                } else {
                    fc.refList[curName] = oneList;
                    curName = lstAll[i].Name
                    oneList = new Array();
                    oneList.push(lstAll[i]);
                }
            }
            fc.refList[curName] = oneList;
            break;
        case cReadItem:
            // TODO check if data != '{"error":"....."}'
            fc.itemList = $.parseJSON(data);
            break;
        case cReadUsers:
            // TODO check if data != '{"error":"....."}'
            fc.userList = $.parseJSON(data);
            gohAdminTab();
            break;
        case cReadCurrentUser:
            // TODO check if data != '{"error":"....."}'
            fc.currentUser = $.parseJSON(data);
            gohHeader();
            break;
        case cReadActors:
            // TODO check if data != '{"error":"....."}'
            fc.actorList = $.parseJSON(data);
            if ( fc.refList == null ) {
                fc.refList = new Object();
            }
            fc.refList['ActorList'] = refListFromNames('ActorList', fc.actorList);
            gohActors();
            gohAdminTab();
            break;
        case cReadImgSensor:
            // TODO check if data != '{"error":"....."}'
            fc.imgSensorList = $.parseJSON(data);
            gohImgSensors();
            gohAdminTab();
            break;
        case cReadSensor:
            // TODO check if data != '{"error":"....."}'
            fc.sensorList = $.parseJSON(data);
            for (i = 0; i < fc.sensorList.length; i++) {
                fc.sensorList[i].Ts='';
                fc.sensorList[i].Val='';
            }
            if ( fc.refList == null ) {
                fc.refList = new Object();
            }
            fc.refList['SensorList'] = refListFromNames('SensorList', fc.sensorList);
            gohSensorTr();
            gohAdminTab();
            break;
        case cReadSensorVal:
            // TODO check if data != '{"error":"....."}'
            var sensorVal = $.parseJSON(data);
            var sensor = getObjById(fc.sensorList,cmde.objectid);
            sensor.Ts = sensorVal.Ts;
            sensor.Val = sensorVal.Val;
            gohSensorTd(sensor.Values[0].IdObject, getObjVal(sensor,"Name"), sensor.Ts, sensor.Val, true);
            break;
        case cReadSensorAct:
            // TODO check if data != '{"error":"....."}'
            fc.sensorActList = $.parseJSON(data);
            gohAdminTab();
            break;
        case cSaveObject:
            // TODO check if data != '{"error":"....."}'

            // Reload Objects from server, this will update GUI as well
            setTimeout(function() { readObjectLst( getReadForItemId(cmde.itemid) ); }, 100);
            break;
        case cTriggerActor:
            // TODO check if data != '{"error":"....."}'
            gohMessage(cmde.command + '(' + data + ')' ,'success',1500);
            break;
        default:
            gohMessage('callServer : action inconnue (' + action + ')', 'danger',3000);
            break;
        }
    });
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function gohMessage(message,msgtype,delay){
    curClass = 'panel panel-info goh-msg';
    if ( msgtype != null && msgtype.length > 0) {
        curClass = 'panel panel-' + msgtype + " goh-msg" ;
    }
    if (delay!=null && delay >0) {
        setTimeout(function() { $('#goh-message').attr("class", "hide" ); }, delay);
    }
    $('#goh-message').attr("class", curClass ) ;
    $('#goh-message').html('<div class="panel-heading">' + message + '</div>');
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function gohHeader() {
    $("#goh-header").html("Welcome " + getObjVal(fc.currentUser,"FirstName") + ' (' + getObjVal(fc.currentUser,"Email") + ')');
}

// -----------------------------------------------------------------------------------------------------------------------------------------
// goh-actors

function gohActors() {
    var html = '';
    var i = 0;
    for (i = 0; i < fc.actorList.length; i++) {
        html = html + '<button type="button" class="btn btn-link" style="background:none; width:100px; color:black;" data-toggle="modal" data-target="#actormodal_' + i + '">';
        html = html + getObjVal(fc.actorList[i],"Name") + '<br><img class="icone" src="';
        html = html + getObjVal(fc.actorList[i],"ImgFileName") + '"></img></button>';

        html = html + '<div id="actormodal_' + i + '" class="modal fade" role="dialog">';
        html = html + '<div class="modal-dialog"><div class="modal-content">';
        html = html + '<div class="modal-header"><h4 class="modal-title">Confirmation</h4></div>';
        html = html + '<div class="modal-body">Actionner ' + getObjVal(fc.actorList[i],"Name") ;
        if ( getObjVal(fc.actorList[i],"DynParamType") != '0' ) {
            html = html + '<input id="actorparam_' + i + '" type="text" class="form-control"></span>';
        }
        html = html + '</div><div class="modal-footer">';
        html = html + '<button type="button" class="btn btn-default" data-dismiss="modal" onclick="actionner(' + i + ')">OK</button>';
        html = html + '<button type="button" class="btn btn-default" data-dismiss="modal">Cancel</button>';
        html = html + '</div>';
        html = html + '</div></div></div>';
    }
    $("#goh-actors").html(html);
}

function actionner(idx) {
    var jparam = '' ;
    if ( getObjVal(fc.actorList[idx],"DynParamType") != '0' ) {
        jparam = $("#actorparam_"+idx).val();
    }
console.log( fc.actorList[idx].Values[0].IdObject + '-' + jparam );
    callServer(cTriggerActor,{ command:'TriggerActor', itemid:0, objectid:fc.actorList[idx].Values[0].IdObject, startts:0, endts:0, jsonparam:jparam });
}

// -----------------------------------------------------------------------------------------------------------------------------------------
// goh-isensor

function gohImgSensors() {
    var html = '';
    var i = 0;
    for (i = 0; i < fc.imgSensorList.length; i++) {
        html = html + '<button type="button" class="btn btn-link" style="background:none; width:100px; color:black;" onclick="showImgSensorReading(' + i + ');">';
        html = html + getObjVal(fc.imgSensorList[i],"Name") + '<br><img class="icone" src="';
        html = html + getObjVal(fc.imgSensorList[i],"ImgFileName") + '"></img></button>';
    }
    $("#goh-isensor").html(html);
}

function showImgSensorReading(idx) {
    var src = getObjVal(fc.imgSensorList[idx],"Param");
    $('#imgsensorsrc').attr('src', src );
    $('#imgsensortitle').html(src);
    $('#imgsensor').show();
}

function hideImgSensorReading() {
    $('#imgsensorsrc').attr('src', '' );
    $('#imgsensor').hide();
}

// -----------------------------------------------------------------------------------------------------------------------------------------
// goh-sensor-row

function gohSensorTd(sensorId,sensorName,readTs,readVal,update) {
    var html = '';
    html = html + '<td>' + sensorId + '</td>';
    html = html + '<td>' + sensorName + '</td>';
    html = html + '<td>' + readVal + '</td>';
    html = html + '<td>' + formatUnixTs(readTs) + '</td>';
    if ( update ) {
        $("#gohsensorrow_"+sensorId).html(html);
    }
    return html;
}

function gohSensorTr() {
    var html = '';
    var i = 0;
    for (i = 0; i < fc.sensorList.length; i++) {
        // TODO ignore inactive sensor
        var objid = fc.sensorList[i].Values[0].IdObject;
        html = html + '<tr id="gohsensorrow_' + objid + '" onclick="readSensorVal(' + objid + ')" >';
        html = html + gohSensorTd(objid, getObjVal(fc.sensorList[i],"Name"), fc.sensorList[i].Ts, fc.sensorList[i].Val,false);
        html = html + '</tr>';
    }
    for (i = 0; i < fc.sensorList.length; i++) {
        readSensorVal(fc.sensorList[i].Values[0].IdObject);
    }
    $("#goh-sensor-row").html(html);
}


function readSensorVal(sensorId) {
    callServer(cReadSensorVal,{command:'ReadSensor', itemid:0, objectid:sensorId, startts:0, endts:0, jsonparam:''});
}

// -----------------------------------------------------------------------------------------------------------------------------------------
// goh-admin-tab



function gohAdminTab() {
    var html = '';
    var i = 0;
    for (i = 0; i < fc.itemList.length; i++) {
        var itemDivId = 'item_' + fc.itemList[i].IdItem;
        html = html + '<div class="panel panel-default">';
        html = html + '<div class="panel-heading" data-toggle="collapse" data-parent="#goh-admin-tab" data-target="#' + itemDivId + '">';
        html = html + '<h4 class="panel-title">' + fc.itemList[i].Name + '</h4>';
        html = html + '</div>';
        html = html + '<div id="' + itemDivId + '" class="panel-collapse collapse">';
        html = html + '<div class="panel-body">';
        html = html + '<ul class="list-group">';
        html = html + gohAdminObjLst(fc.itemList[i].IdItem,fc.itemList[i].Name);
        html = html + '</ul></div></div></div>';
    }
    $("#goh-admin-tab").html(html);
}


function gohAdminObjLst(itemId,itemName) {
    var html = '';
    var i = 0;
    var objLst = objectListForItemId(itemId);
    if ( objLst == null ) {
        html = html + 'ERROR gohAdminObjLst : objLst=null ';
    } else {
        for (i = 0; i < objLst.length; i++) {
            html = html + '<li class="list-group-item" style="padding:0px;">';
            html = html + '<div onclick="gohAdminEditObj(\'' + itemName + '\',' + itemId + ',' + objLst[i].Values[0].IdObject + ');" style="padding:5px;">';
            html = html + getObjectName(objLst[i]);
            html = html + '</div></li>';
        }
    }
    if ( itemName != null && itemName.length > 0 ) {
        html = html + '<li class="list-group-item" style="padding:0px;">';
        html = html + '<div onclick="gohAdminEditObj(\'' + itemName + '\',' + itemId + ',0);" style="padding:5px;">';
        html = html + '<span class="glyphicon glyphicon-plus"></span> New ' + itemName;
        html = html + '</div></li>';
    }
    return html;
}

// ---------------------------
// global var for object edit

var curObjAdminEdit = null;

// ---------------------------


function gohAdminEditObj(itemName,itemId,objectId) {
    var html = '';
    var objLst = objectListForItemId(itemId);
    if ( objLst == null ) {
        html = html + 'gohAdminEditObj : ERROR objLst=null ';
    } else {
        if ( objectId == 0 ) {
            // TODO : use Fields from Idem definition rather than an existing HomeObject ... in case we need to create the first obj
            curObjAdminEdit = newHomeObject(objLst[0].Fields);
        } else {
            curObjAdminEdit = getObjById(objLst, objectId );
        }
    }

    html = html + '<div class="form-panel"><div class="panel panel-default">';
    html = html + '<div class="panel-heading">' + itemName + '</div>';
    html = html + '<div class="panel-body"><form class="form-horizontal">';
    if ( curObjAdminEdit == null ) {
        html = html + 'gohAdminEditObj : ERROR curObjAdminEdit=null ';
    } else {
        var i = 0;
        for (i = 0; i < curObjAdminEdit.Fields.length; i++) {
            html = html + '<span style="font-size: 0.7em;">';
            html = html + gohAdminEditField(i);
            html = html + '</span>';
        }
    }
    html = html + '</form></div>';
    html = html + '<div class="panel-footer">';
    html = html + '<button type="button" class="btn btn-default" onclick="adminSaveObj()">Save</button> ';
    html = html + '<button type="button" class="btn btn-default" onclick="adminCancelEdit()">Cancel</button>';
    html = html + '</div>';
    html = html + '</div></div>';

    $("#goh-admin-edit-object").html(html);
    adminHasError();
    $("#goh-admin-edit-object").attr("class", 'gray-out-page');

}

function gohAdminEditField(idx) {
    var field = curObjAdminEdit.Fields[idx];
    var val = curObjAdminEdit.Values[idx];
    var html = '';
    html = html + '<div id="admin_div_field_' + field.IdField + '" class="form-group has-feedback">';
    html = html + '<label class="col-sm-2 control-label" for="admin_field_' + field.IdField + '">' + field.Label + '</label>';
    html = html + '<div class="col-sm-10">';
    switch ( htmlfieldtype(field) ) {
    case 'input':
        html = html + '<input id="admin_field_' + field.IdField + '" type="' + getinputtype(field) + '" placeholder="' + field.Helper + '" ';
        html = html + 'oninput="isValidFieldVal(' + idx + ');" class="form-control input-sm" value="' + val.Val + '">';
        html = html + '<span id="admin_field_ico_' + field.IdField + '" class="glyphicon form-control-feedback"></span>';
        break;
    case 'select':
        html = html + '<select id="admin_field_' + field.IdField + '" placeholder="' + field.Helper + '" onchange="' + isValidFieldVal(idx) + '" ';
        html = html + 'value="' + val.Val + '" class="form-control">';
        var i = 0;
        for (i = 0; i < fc.refList[field.RefList].length; i++) {
            html = html + '<option value="' + fc.refList[field.RefList][i].Code + '" >' + fc.refList[field.RefList][i].Label + '</option>';
        }
        html = html + '</select>';
        break;
    default:
        html = html + 'Unknonw field type : ' + field.IdField + ' / ' + field.Name + '.';
        break;
    }
    html = html + '</div></div>';
    return html;
}

function htmlfieldtype(field) {
    switch (field.IdDataType){
    case DBTypeInt:
        if ( field.RefList=='' ) return 'input';
        return 'select';
    case DBTypeFloat:
    case DBTypeText:
        return 'input';
    default :
        return '';
    }
}

function getinputtype(field) {
    switch (field.IdDataType) {
    case DBTypeInt:
    case DBTypeFloat:
        return 'number';
    default:
        switch (field.Regexp) {
        case 'email':
        case 'tel':
        case 'password':
            return field.Regexp;
        default:
            return 'text';
        }
    }
}

function isValidFieldVal(idx) {
    var field = curObjAdminEdit.Fields[idx];
    var curVal = $('#admin_field_' + field.IdField).val();
    if ( field.Required  != '0' && curVal == '' ) {
        $("#admin_div_field_"+field.IdField).attr("class", 'form-group has-feedback has-error');
        $("#admin_field_ico_"+field.IdField).attr("class", 'glyphicon form-control-feedback glyphicon-remove');
        curObjAdminEdit.Values[idx]["_v_"] = false;
console.log('isValidFieldVal('+idx+') a= false ' + field.Required + '-' + curVal );
        return false;
    }
    if ( field.Regexp != '' ) {
        var pattern = new RegExp(fc.refList[field.Regexp][0].Label,"g");
        if ( !pattern.test(curVal) ) {
            $("#admin_div_field_"+field.IdField).attr("class", 'form-group has-feedback has-error');
            $("#admin_field_ico_"+field.IdField).attr("class", 'glyphicon form-control-feedback glyphicon-remove');
            curObjAdminEdit.Values[idx]["_v_"] = false;
console.log('isValidFieldVal('+idx+') b= false ' + field.Regexp + '-' + fc.refList[field.Regexp][0].Label + '-' + curVal );
            return false;
        }
    }
    if ( field.UniqKey != '0' ) {
        if ( curVal == '' ) {
            $("#admin_div_field_"+field.IdField).attr("class", 'form-group has-feedback has-error');
            $("#admin_field_ico_"+field.IdField).attr("class", 'glyphicon form-control-feedback glyphicon-remove');
            curObjAdminEdit.Values[idx]["_v_"] = false;
console.log('isValidFieldVal('+idx+') c= false');
            return false;
        } else {
            if ( searchObjByVal(objectListForItemId(field.IdItem),field.IdItem,curObjAdminEdit.Values[0].IdObject,field.Name,curVal) != null ) {
                $("#admin_div_field_"+field.IdField).attr("class", 'form-group has-feedback has-error');
                $("#admin_field_ico_"+field.IdField).attr("class", 'glyphicon form-control-feedback glyphicon-remove');
                curObjAdminEdit.Values[idx]["_v_"] = false;
console.log('isValidFieldVal('+idx+') d= false');
                return false;
            }
        }
    }
    $("#admin_div_field_"+field.IdField).attr("class", 'form-group has-feedback has-success');
    $("#admin_field_ico_"+field.IdField).attr("class", 'glyphicon form-control-feedback glyphicon-ok');
    curObjAdminEdit.Values[idx]["_v_"] = true;
console.log('isValidFieldVal('+idx+') = true');
    return true;
}

function adminHasError() {
    var hasError = false;
    var i = 0;
    for (i = 0; i < curObjAdminEdit.Values.length; i++) {
        if ( isValidFieldVal(i) != true ) hasError = true;
    }
    return hasError;
}

function adminCancelEdit() {
    curObjAdminEdit = null;
    $("#goh-admin-edit-object").attr("class", 'hide');
}

function adminSaveObj() {
    if ( adminHasError() ) {
        gohMessage('<center><br><br>Bad value(s)</br></br></br></center>', 'danger',3000);
        return;
    }
    var i = 0;
    for (i = 0; i < curObjAdminEdit.Values.length; i++) {
        curObjAdminEdit.Values[i].Val = $('#admin_field_' + curObjAdminEdit.Fields[i].IdField).val();
    }

    callServer(cSaveObject,
        { command:'SaveObject', itemid:curObjAdminEdit.Fields[0].IdItem, objectid:0, startts:0, endts:0, jsonparam: $.toJSON(curObjAdminEdit) });

    curObjAdminEdit = null;
    $("#goh-admin-edit-object").attr("class", 'hide');
}

// -----------------------------------------------------------------------------------------------------------------------------------------

$(document).ready(function(){

    // Read Reference lists
    callServer(cReadRefList, { command:'ReadRefList', itemid:0, objectid:0, startts:0, endts:0, jsonparam:'%' });
    // Read Item definition
    callServer(cReadItem, { command:'ReadItem', itemid:0, objectid:0, startts:0, endts:0, jsonparam:'' });
    // Read current user
    callServer(cReadCurrentUser, { command:'ReadCurrentUser', itemid:0, objectid:0, startts:0, endts:0, jsonparam:'' });

    // Read actors
    readObjectLst(cReadActors);

    // Read img sensors
    readObjectLst(cReadImgSensor);

    // Read users
    readObjectLst(cReadUsers);

    // Read sensors
    readObjectLst(cReadSensor);

    // Read sensorAct i.e. actors trigger by sensor reading
    readObjectLst(cReadSensorAct);

});

// -----------------------------------------------------------------------------------------------------------------------------------------


