
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
            showGohHeader();
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
            showGohActors();
            break;
        case cReadImgSensor:
            // TODO check if data != '{"error":"....."}'
            fc.imgSensorList = $.parseJSON(data);
            showGohImgSensors();
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
            break;
        case cReadSensorVal:
            // TODO check if data != '{"error":"....."}'
            var sensorVal = $.parseJSON(data);
            var sensor = getObjById(fc.sensorList,cmde.objectid);
            sensor.Ts = sensorVal.Ts;
            sensor.Val = sensorVal.Val;
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
        }
        if ( forceRefresh ) {
            //vm.$forceUpdate();
        }
    });
}

// -----------------------------------------------------------------------------------------------------------------------------------------

function showGohHeader() {
    $("#goh-header").html("Welcome " + getObjVal(fc.currentUser,"FirstName") + ' (' + getObjVal(fc.currentUser,"Email") + ')');
}

// -----------------------------------------------------------------------------------------------------------------------------------------
// goh-actors

function showGohActors() {
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

function showGohImgSensors() {
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


$(document).ready(function(){



/*


    Vue.component('goh-isensor-show', {
        props: ['imgsensorsrc'],
        template: ' ',
        methods: {
            hideImg: function (event) { vm.imgSensorSrc = '';  }
            // TODO : add button to Save img
        }
    });





    Vue.component('goh-sensors-row', {
        props: {
            sensor:null,
            readvalue: '',
            readdate: ''
        },
        template: ' <tr @click="updateval">\
                    <td>{{sensor.Values[0].IdObject}}</td><td>{{getObjVal(sensor,"Name")}}</td><td>{{readvalue}}</td><td>{{showdate}}</td>\
                    </tr>',
        computed: {
            showdate:function () {
                if (this.readdate == 0) {
                    this.updateval();
                    return '/';
                }
                var now = new Date();
                var dt = new Date(this.readdate);
                if ( dt.getDate() != now.getDate() )
                    return dt.getFullYear() + '-' + dt.getMonth() + '-' + dt.getDate();
                return ('0'+dt.getHours()).slice(-2) + ':' + ('0'+dt.getMinutes()).slice(-2) + ':' + ('0'+dt.getSeconds()).slice(-2);
            }
        },
        methods: {
            updateval: function () {
                var objid=this.sensor.Values[0].IdObject;
                callServer(cReadSensorVal,true,{command:'ReadSensor', itemid:0, objectid:objid, startts:0, endts:0, jsonparam:''});
            }
        }
    });



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
            name:function () {
                var name = getObjVal(this.homeobj,"Name");
                if ( name == '' && this.homeobj.Values[0].IdObject == 0 ) {
                    name = 'New ' + this.itemidname ;
                }
                if ( name == '' && getObjVal(this.homeobj,"FirstName") != '' ) {
                    name = getObjVal(this.homeobj,"FirstName") + ' ' + getObjVal(this.homeobj,"LastName") + ' (' + getObjVal(this.homeobj,"Email") + ')';
                }
                if ( name == '' && getObjVal(this.homeobj,"idMasterObj") != '' ) {
                    name = getObjVal(getObjById(vm.sensorList,getObjVal(this.homeobj,"idMasterObj")),"Name") + " to " +
                        getObjVal(getObjById(vm.actorList,getObjVal(this.homeobj,"idActor")),"Name") + " on '" +
                        htmlEncode(getObjVal(this.homeobj,"Condition")) + "'";
                }
                if ( name == '' ) {
                    name = 'Object_' + this.homeobj.Values[0].IdObject ;
                }
                return name;
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




    Vue.component('goh-admin-tab', {
        props: {
            objlist:null
        },
        template: ' <div class="panel panel-default">\
                        <div class="panel-heading" data-toggle="collapse" data-parent="#admin_list" :data-target="\'#\'+divid">\
                            <h4 class="panel-title">{{itemidname}}</h4>\
                        </div>\
                        <div :id="divid" class="panel-collapse collapse">\
                            <div class="panel-body">\
                                <ul class="list-group">\
                                    <goh-admin-obj v-for="i in objlist" :homeobj="i" ></goh-admin-obj>\
                                    <goh-admin-obj :homeobj="homeobj" ></goh-admin-obj>\
                                </ul>\
                            </div>\
                        </div>\
                    </div>',
        computed: {
            itemidname: function () {
                if ( this.objlist == null ) return 'lst null';
                if ( this.objlist[0].Fields == null ) return 'fields null';
                return getItemNameById( vm.itemList, this.objlist[0].Fields[0].IdItem );
            },
            divid: function() {
                return 'admin_' + this.itemidname.replace(/ /g,"_");
            },
            homeobj: function() {
                if ( this.objlist == null ) return null;
                return newHomeObject(this.objlist[0])
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


