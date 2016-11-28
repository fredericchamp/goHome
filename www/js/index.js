
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


var gStartup            = { startup:true, loadAsk:0, loadGot:0 } ;

var fc = new Object({
    initLoadDone: false,
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

function showMessage(message,msgtype,delay){
    curClass = 'panel panel-info goh-msg';
    if ( msgtype != null && msgtype.length > 0) {
        curClass = 'panel panel-' + msgtype + " goh-msg" ;
    }
    if (delay!=null && delay >0) {
        setTimeout(function() { $('#gohMessage').attr("class", "hide" ); }, delay);
    }
    $('#gohMessage').attr("class", curClass ) ;
    $('#gohMessageText').html(message);
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
            nameIdx = i; // TODO should use IdField rather than index
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

function newHomeObject(obj) {
    if ( obj == null ) return null;
    var newObj = new Object();
    newObj.Fields = obj.Fields;
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

function readObjectLst( action, forceRefresh ) {
   var itemId = getItemIdForRead( action );
    if ( itemId <= 0 ) {
        // TODO ERROR unknown itemId or action
        return;
    }
    callServer(action,forceRefresh,{ command:'ReadObject', itemid:itemId, objectid:0, startts:0, endts:0, jsonparam:'' });
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function callServer(action,forceRefresh,cmde){
    if ( gStartup.startup ) {
        gStartup.loadAsk++;
    }
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
            gohHeader();
            break;
        case cReadCurrentUser:
            // TODO check if data != '{"error":"....."}'
            fc.currentUser = $.parseJSON(data);
            break;
        case cReadActors:
            // TODO check if data != '{"error":"....."}'
            fc.actorList = $.parseJSON(data);
            if ( fc.refList == null ) {
                fc.refList = new Object();
            }
            fc.refList['ActorList'] = refListFromNames('ActorList', fc.actorList);
            gohActors();
            break;
        case cReadImgSensor:
            // TODO check if data != '{"error":"....."}'
            fc.imgSensorList = $.parseJSON(data);
            gohImgSensors();
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
            break;
        case cSaveObject:
            // TODO check if data != '{"error":"....."}'
            // TODO if save fail, read original values from server and update corresponding fc.xxxxList to restore valid values ... or reload page :-)
            if ( forceRefresh ) {
                setTimeout(function() { readObjectLst( getReadForItemId(cmde.itemid), forceRefresh ); }, 1500);
                forceRefresh = false;
            }
            break;
        case cTriggerActor:
            // TODO check if data != '{"error":"....."}'
            showMessage(cmde.command + '(' + data + ')' ,'success',1500);
            break;
        default:
            showMessage('callServer : action inconnue (' + action + ')', 'danger',3000);
            break;
        }
        gStartup.loadGot++;
        if (gStartup.startup==false && gStartup.loadAsk <= gStartup.loadGot) {
            fc.initLoadDone=true;
            gohAdminTab();
        }
        if ( forceRefresh ) {
            //vm.$forceUpdate();
        }
    });
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
    callServer(cTriggerActor,false,{ command:'TriggerActor', itemid:0, objectid:fc.actorList[idx].Values[0].IdObject, startts:0, endts:0, jsonparam:jparam });
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
    callServer(cReadSensorVal,true,{command:'ReadSensor', itemid:0, objectid:sensorId, startts:0, endts:0, jsonparam:''});
}

// -----------------------------------------------------------------------------------------------------------------------------------------
// goh-admin-tab



function gohAdminTab() {
    var html = '';
    var i = 0;
    for (i = 0; i < fc.itemList.length; i++) {
        var itemDivId = 'item_' + fc.itemList[i].IdItem;
        html = html + '<div class="panel panel-default">';
        html = html + '<div class="panel-heading" data-toggle="collapse" data-parent="#admin_list" data-target="#' + itemDivId + '">';
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
    html = html + '<div class="form-panel"><div class="panel panel-default">';
    html = html + '<div class="panel-heading">' + itemName + '</div>';
    html = html + '<div class="panel-body"><form class="form-horizontal">';

    var objLst = objectListForItemId(itemId);
    if ( objLst == null ) {
        html = html + 'gohAdminEditObj : ERROR objLst=null ';
    } else {
        if ( objectId == 0 ) {
            curObjAdminEdit = newHomeObject(objLst[0])
        }
        curObjAdminEdit = getObjById(objLst, objectId );
    }
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
    html = html + '<button type="button" class="btn btn-default" onclick="adminSaveObj()">Save</button>';
    html = html + '<button type="button" class="btn btn-default" onclick="adminCancelEdit()">Cancel</button>';
    html = html + '</div>';
    html = html + '</div></div>';

    $("#goh-admin-edit-object").html(html);
    $("#goh-admin-edit-object").attr("class", 'gray-out-page');

}

function adminCancelEdit() {
    curObjAdminEdit = null;
    $("#goh-admin-edit-object").attr("class", 'hide');
}

function adminSaveObj() {
    curObjAdminEdit = null;
    $("#goh-admin-edit-object").attr("class", 'hide');
}

function gohAdminEditField(idx) {
    var html = '';
    html = html + curObjAdminEdit.Fields[idx].Name + '<br>';
    return html;
}


// -----------------------------------------------------------------------------------------------------------------------------------------

$(document).ready(function(){



/*

    Vue.component('goh-obj-edit-field', {
        props: {
            field:null,
            vals:null
        },
        template: ' <div :id="\'div_\'+field.Name" class="form-group has-feedback">\
                        <label class="col-sm-2 control-label" :for="field.Name"> {{field.Label}} </label>\
                        <div class="col-sm-10">\
                            <span v-if="(htmlfieldtype(field)==\'input\')">\
                                <input v-model="vals[field.Name]" :type="getinputtype(field)" :id="field.Name" :placeholder="field.Helper" \
                                @input="isvalid(field,$event.target.value)" class="form-control input-sm">\
                                <span :id="\'ico_\'+field.Name" class="glyphicon form-control-feedback"></span>\
                            </span>\
                            <span v-if="(htmlfieldtype(field)==\'select\')">\
                                <select v-model="vals[field.Name]" :id="field.Name" @input="isvalid(field,$event.target.value)" :placeholder="field.Helper" class="form-control">\
                                    <option v-for="oneVal in vm.refList[field.RefList]" :value="oneVal.Code" >{{oneVal.Label}}</option>\
                                </select>\
                            </span>\
                        </div>\
                    </div>',
        mounted: function () {
            this.isvalid(this.field,this.vals[this.field.Name]);
        },
        methods: {
            isvalid: function (field,curVal) {
                if ( field.Required  != '0' && curVal == '' ) {
                    $("#div_"+field.Name).attr("class", 'form-group has-feedback has-error');
                    $("#ico_"+field.Name).attr("class", 'glyphicon form-control-feedback glyphicon-remove');
                    this.vals["_v_"+field.Name] = false;
                    return false;
                }
                if ( field.Regexp != '' ) {
                    var pattern = new RegExp(vm.refList[field.Regexp][0].Label,"g");
                    if ( !pattern.test(curVal) ) {
                        $("#div_"+field.Name).attr("class", 'form-group has-feedback has-error');
                        $("#ico_"+field.Name).attr("class", 'glyphicon form-control-feedback glyphicon-remove');
                        this.vals["_v_"+field.Name] = false;
                        return false;
                    }
                }
                if ( field.UniqKey != '0' ) {
                    if ( curVal == '' ) {
                        $("#div_"+field.Name).attr("class", 'form-group has-feedback has-error');
                        $("#ico_"+field.Name).attr("class", 'glyphicon form-control-feedback glyphicon-remove');
                        this.vals["_v_"+field.Name] = false;
                        return false;
                    } else {
                        var find = false;
                        switch (field.IdItem) {
                        case vm.userList[0].Fields[0].IdItem:
                            find = (searchObjByVal(vm.userList,field.IdItem,this.vals.IdObject,field.Name,curVal) != null) ;
                            break;
                        case vm.actorList[0].Fields[0].IdItem:
                            find = (searchObjByVal(vm.actorList,field.IdItem,this.vals.IdObject,field.Name,curVal) != null) ;
                            break;
                        case vm.sensorList[0].Fields[0].IdItem:
                            find = (searchObjByVal(vm.sensorList,field.IdItem,this.vals.IdObject,field.Name,curVal) != null) ;
                            break;
                        case vm.sensorActList[0].Fields[0].IdItem:
                            find = (searchObjByVal(vm.sensorActList,field.IdItem,this.vals.IdObject,field.Name,curVal) != null) ;
                            break;
                        case vm.imgSensorList[0].Fields[0].IdItem:
                            find = (searchObjByVal(vm.imgSensorList,field.IdItem,this.vals.IdObject,field.Name,curVal) != null) ;
                            break;
                        default:
                            find = false ;
                        }
                        if ( find ) {
                            $("#div_"+field.Name).attr("class", 'form-group has-feedback has-error');
                            $("#ico_"+field.Name).attr("class", 'glyphicon form-control-feedback glyphicon-remove');
                            this.vals["_v_"+field.Name] = false;
                            return false;
                        }
                    }
                }
                $("#div_"+field.Name).attr("class", 'form-group has-feedback has-success');
                $("#ico_"+field.Name).attr("class", 'glyphicon form-control-feedback glyphicon-ok');
                this.vals["_v_"+field.Name] = true;
                return true;
            },
            htmlfieldtype: function (field) {
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
            },
            getinputtype: function (field) {
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
        }
    });



    Vue.component('goh-admin-obj', {
        props: {
            showmodalform:false,
            editobj:null,
            homeobj:null
        },
        template: ' <li class="list-group-item" style="padding:0px;">\
                        <div @click="showmodalform=true;" style="padding:5px;">\
                            <span v-if="homeobj.Values[0].IdObject == 0" class="glyphicon glyphicon-plus"></span> {{name}}\
                        </div>\
                        <div class="gray-out-page" v-if="showmodalform==true"><div class="form-panel">\
                            <div class="panel panel-default">\
                                <div class="panel-heading">{{itemidname}}</div>\
                                <div class="panel-body">\
                                    <form class="form-horizontal"><span v-for="(field, idx) in homeobj.Fields"  style="font-size: 0.7em;">\
                                        <goh-obj-edit-field :field="field" :vals="getEditobj"></goh-obj-edit-field>\
                                    </span></form>\
                                </div>\
                                <div class="panel-footer">\
                                    <button type="button" class="btn btn-default" @click="this.saveObj">Save</button>\
                                    <button type="button" class="btn btn-default" @click="this.cancelEdit">Cancel</button>\
                                </div>\
                            </div>\
                        </div></div>\
                    </li>',
        computed: {
            itemidname: function () {
                return getItemNameById( vm.itemList, this.homeobj.Fields[0].IdItem );
            },
            getEditobj:function () {
                if ( this.editobj == null ) {
                    var obj = new Object;
                    for ( idx = 0; idx < this.homeobj.Fields.length; idx++ ) {
                        obj[this.homeobj.Fields[idx].Name] = this.homeobj.Values[idx].Val;
                    }
                    obj.IdObject = this.homeobj.Values[0].IdObject;
                    this.editobj = obj;
                }
                return this.editobj;
            }
        },
        methods: {
            cancelEdit: function() {
                this.editobj = null;
				this.showmodalform=false;
            },
            saveObj: function() {
                for ( idx = 0; idx < this.homeobj.Fields.length; idx++ ) {
                    if ( this.editobj["_v_"+this.homeobj.Fields[idx].Name] == false ) {
                        showMessage('<center><br><br>Bad value(s)</br></br></br></center>', 'danger',3000);
                        return;
                    }
                }
                for ( idx = 0; idx < this.homeobj.Fields.length; idx++ ) {
                    this.homeobj.Values[idx].Val = this.editobj[this.homeobj.Fields[idx].Name]
                }
                // SaveObject
                callServer(cSaveObject, (this.homeobj.Values[0].IdObject == '0'),
                    { command:'SaveObject', itemid:this.homeobj.Fields[0].IdItem, objectid:0, startts:0, endts:0,
                        jsonparam: $.toJSON(this.homeobj) });

				this.showmodalform=false;
            }
        }
    });

*/
// -----------------------------------------------------------------------------------------------------------------------------------------


    // Read Reference lists
    callServer(cReadRefList, false, { command:'ReadRefList', itemid:0, objectid:0, startts:0, endts:0, jsonparam:'%' });
    // Read Item definition
    callServer(cReadItem, false, { command:'ReadItem', itemid:0, objectid:0, startts:0, endts:0, jsonparam:'' });
    // Read current user
    callServer(cReadCurrentUser, false, { command:'ReadCurrentUser', itemid:0, objectid:0, startts:0, endts:0, jsonparam:'' });

    // Read actors
    readObjectLst(cReadActors, false);

    // Read img sensors
    readObjectLst(cReadImgSensor, false);

    gStartup.startup = false;

    // Read users
    readObjectLst(cReadUsers, false);

    // Read sensors
    readObjectLst(cReadSensor, false);

    // Read sensorAct i.e. actors trigger by sensor reading
    readObjectLst(cReadSensorAct, false);



// -----------------------------------------------------------------------------------------------------------------------------------------

});


