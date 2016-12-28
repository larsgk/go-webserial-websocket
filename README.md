# go-webserial-websocket
A Go based WebSocket proxy for (USB) Serial devices

The client side API is inspired by WebSerial, WebUSB and chrome.serial

Usage (e.g.):

* clone the repo
* make a folder under _static_ with your serial port web application
* In the app, include /go-webserial.js

Scan for devices:

```javascript
SerialPort.requestPorts([{"vendorId":0x0425,"productId":0x0408}]).then( function(reply) {
        console.log("Serial port list: ", reply);
        if(reply.length == 1) {
            thePort = new SerialPort(reply[0].path, {baudrate:57600});
            thePort.onClose = connectionClosed;
            thePort.connect(onReceiveCallback).then( () => { isConnected = true; send_init(); });
        }
    });
```

Write (in this case, JSON) to the port:

```javascript
var JsonToArray = function(json)
{
    var str = JSON.stringify(json, null, 0);
    var ret = new Uint8Array(str.length);
    for (var i = 0; i < str.length; i++) {
        ret[i] = str.charCodeAt(i);
    }
    return ret;
};

function send_json(json) {
    if(thePort) {
        var encoded = JsonToArray(json);
        thePort.write(encoded.buffer);
    }
}
```

See **main.js** under **static** for a more complete example.

NOTE: The **static** folder contains a minimal example, communicating with a freescale FRDM-KL25Z evaluation/sensor board with the empiriKit firmware on it from https://github.com/larsgk/mbed_applications