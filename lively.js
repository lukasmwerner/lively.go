import morphdom from "https://cdn.jsdelivr.net/npm/morphdom@2.7.0/+esm";
console.log("lively.go");

const uri = `${
  window.location.protocol.replace("http", "ws")
}//${window.location.host}${window.location.pathname}`;

const bindingMap = new Map();
const pusherMap = new Map();
const socket = new WebSocket(uri);
socket.addEventListener("open", (_) => {
  const bindings = document.querySelectorAll("[lively-bind]");
  const bindNames = new Array(bindings.length);
  for (let i = 0; i < bindings.length; i++) {
    bindingMap.set(bindings[i].getAttribute("lively-bind"), bindings[i]);
    bindNames[i] = bindings[i].getAttribute("lively-bind");
  }
  const pushers = document.querySelectorAll("[lively-push]");
  const pushNames = new Array(pushers.length);
  for (let i = 0; i < pushers.length; i++) {
    pusherMap.set(pushers[i].getAttribute("lively-bind"), pushers[i]);
    pushNames[i] = bindings[i].getAttribute("lively-push");
    pushers[i].addEventListener(
      bindings[i].getAttribute("lively-push-type"),
      (e) => {
        socket.send(JSON.stringify({
          kind: "push",
          name: pushNames[i],
          event: e,
        }));
      },
    );
  }
  socket.send(JSON.stringify({
    kind: "init",
    bindings: bindNames, // things that come to client side from server
    pushers: pushNames, // things that send back to server side from client
  }));
});
socket.addEventListener("message", (e) => {
  const msg = JSON.parse(e.data);
  if (msg.kind == "setCmd") {
    morphdom(bindingMap.get(msg.key), msg.value, {
      childrenOnly: true,
    });
  }
});
