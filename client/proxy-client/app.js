const dgram = require("dgram")

const WebSocket = require('ws')

const wss = new WebSocket.Server({ port: 8000 })

//Initialize a UDP server to listen for json payloads
var srv = dgram.createSocket({ type: 'udp4', reuseAddr: true });
var flag = false
var wes;

srv.on("listening", function () {
    var address = srv.address();
    console.log("server listening " + address.address + ":" + address.port);
  });
  
  srv.bind(2222)



wss.on('connection', ws => {
    flag = true
    wes = ws
})

wss.on('error', err => {
    flag = false
    console.log(err)
    wss = new WebSocket.Server({ port: 8000 })
})

wss.on('close', () => {
    flag = false;
    wss = new WebSocket.Server({ port: 8000 })
})

srv.on("message", function (msg, rinfo) {
    console.log(`server got: ${msg}`)
    // " from " + rinfo.address + ":" + rinfo.port
   if (flag) wes.send("" + msg)
});
