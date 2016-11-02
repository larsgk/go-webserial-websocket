var JsonToArray = function(json)
{
    var str = JSON.stringify(json, null, 0);
    var ret = new Uint8Array(str.length);
    for (var i = 0; i < str.length; i++) {
        ret[i] = str.charCodeAt(i);
    }
    return ret
};

var binArrayToJson = function(binArray)
{
    var str = "";
    for (var i = 0; i < binArray.length; i++) {
        str += String.fromCharCode(parseInt(binArray[i]));
    }
    return JSON.parse(str)
}

var binArrayToString = function(binArray)
{
    var str = "";
    for (var i = 0; i < binArray.length; i++) {
        str += String.fromCharCode(parseInt(binArray[i]));
    }
    return str;
}

function handleMessage(message) {
    console.log("HANDLE JSON RESPONSE: ", message);
    try {
        var json = JSON.parse(message)
        console.log("It's JSON!", json);

        if(json.datatype === "StreamData" && json.accelerometerdata) {
            var accData = json.accelerometerdata;
            send_set_rgb(Math.abs(accData[0]&0xff), Math.abs(accData[1]&0xff), Math.abs(accData[2]&0xff));
        }
    } catch (e) {
        console.log(e)
    }
}

var receiveStr = "";

var onReceiveCallback = function(rdata) {
  var rstr = binArrayToString(rdata);
  // do hackish handling ;)
  receiveStr += rstr;
  if(rstr.indexOf("}") != -1) {
    handleMessage(receiveStr);
    receiveStr="";
  }
};

var thePort = null;

function send_CMD(json) {
    if(thePort) {
        var encoded = JsonToArray(json)
        thePort.write(encoded.buffer);
    }
}

function send_GETINF() {
    send_CMD({GETINF:1});
}

function send_start_streaming() {
    send_CMD({STRACC:1});
}

function send_set_rgb(r,g,b) {
    send_CMD({SETRGB:[r,g,b]});
}

function send_stop_streaming() {
    send_CMD({STRACC:0});
}

function send_init() {
    setTimeout(send_GETINF, 200);
}

var empiriKitPnPFilter = {"vendorId":0x0425,"productId":0x0408};

var isConnected = false;
var isScanning = false;

function connectionClosed() {
    isConnected = false;
}

function scan() {
    isScanning = true;
    SerialPort.requestPorts([empiriKitPnPFilter]).then( function(reply) {
      console.log("Serial port list: ", reply);

      if(reply.length == 1) {
        thePort = new SerialPort(reply[0].path, {baudrate:57600});
        thePort.onClose = connectionClosed;
        thePort.connect(onReceiveCallback).then( () => { isConnected = true; send_init(); });
      }
      isScanning = false;
    });
}

function setStatus(msg) {
    var el = document.getElementById("status");
    if(el) {
        el.innerHTML = msg;
    }
}

function init() {
    setStatus("Initalizing...");
    // buttons
    document.getElementById("start").onclick = send_start_streaming;
    document.getElementById("stop").onclick = send_stop_streaming;

    setInterval(function() {
        if(!isConnected && !isScanning) {
            scan();
        }
    }, 1000);
}

window.addEventListener("load", init);