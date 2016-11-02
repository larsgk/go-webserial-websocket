'use strict';
/*jshint esversion: 6 */

class SerialPort {
    constructor(devicePath, modeOptions) {
        console.log("Creating a serial connection to ", devicePath);
        this.devicePath = devicePath;
        this.modeString = "";
        if(modeOptions) {
            for(var key in modeOptions) {
                this.modeString += "&" + key + "=" + modeOptions[key];
            }
        }

        if (window.location.protocol === "https:") {
            this.wsbaseuri = "wss:";
        } else {
            this.wsbaseuri = "ws:";
        }
        this.wsbaseuri += "//" + window.location.host;
    }

    static requestPorts(filters) {
        return new Promise(
            function(resolve, reject) {
                var oReq = new XMLHttpRequest();
                oReq.addEventListener("load", function() {
                    var res = JSON.parse(oReq.responseText);
                    if (res && res.Type === "CommPorts" && res.Data) {
                        if(filters) {
                            resolve(SerialPort.filterPorts(res.Data, filters));
                        } else {
                            resolve(res.Data);
                        }
                    } else {
                        reject("Error in request to get comm port list!");
                    }
                });
                oReq.open("GET", "/commports");
                oReq.send();
            });
    }

    static filterPorts(portList, filters) {
        var result = [];

        portList.forEach( function(port) {
            var idx, filter;
            for(idx = 0; idx < filters.length; idx++) {
                filter = filters[idx];
                for(var key in filter) {
                    if(port[key] !== filter[key]) {
                        return;
                    }
                }
            }
            result.push(port);
        });

        return result;
    }

    set onClose(callback) {
        this.onCloseCallback = callback;
    }

    get onClose() {
        return this.onCloseCallback;
    }

    connect(callback) {
        return new Promise(
            function(resolve, reject) {

                //this.websoc = new WebSocket('ws://127.0.0.1:3000/wsconnect?path=' + this.devicePath);
                this.websoc = new WebSocket(this.wsbaseuri + '/wsconnect?path=' + this.devicePath + this.modeString);
                this.websoc.binaryType = 'arraybuffer';

                this.websoc.onopen = function () {
                    console.log("WebSocket opened...");
                    resolve();
                };

                this.websoc.onmessage = function(evt) {
                    callback(new Int8Array(evt.data));
                };

                this.websoc.onclose = function() {
                    this.websoc = undefined;
                    if(this.onCloseCallback) {
                        this.onCloseCallback();
                    }
                    console.log("WebSocket closed...");
                    reject();
                }.bind(this);

            }.bind(this));
    }

    disconnect() {
        if(this.websoc) {
            this.websoc.close();
        }
    }

    write(data) {
        if(this.websoc) {
            this.websoc.send(data);
        }
    }
}
