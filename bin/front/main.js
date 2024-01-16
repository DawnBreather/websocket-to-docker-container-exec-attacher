// import {FitAddon} from "xterm-addon-fit";
// import {AttachAddon} from "xterm-addon-attach";
// import {WebLinksAddon} from "xterm-addon-web-links/src/WebLinksAddon";
// import {Unicode11Addon} from "xterm-addon-unicode11/src/Unicode11Addon";
// import {SerializeAddon} from "xterm-addon-serialize/src/SerializeAddon";

const terminal = new Terminal({
  screenKeys: true,
  useStyle: true,
  cursorBlink: true,
  fullscreenWin: true,
  maximizeWin: true,
  screenReaderMode: true,
  cols: 128,
  allowProposedApi: true,
});

terminal.open(document.getElementById('terminal'));

const ws = new WebSocket('ws://localhost:4242/ws');
const attachAddon = new AttachAddon.AttachAddon(ws);
const fitAddon = new FitAddon.FitAddon();
terminal.loadAddon(fitAddon);
const webLinksAddon = new WebLinksAddon.WebLinksAddon();
terminal.loadAddon(webLinksAddon);
const unicode11Addon = new Unicode11Addon.Unicode11Addon();
terminal.loadAddon(unicode11Addon);
const serializeAddon = new SerializeAddon.SerializeAddon();
terminal.loadAddon(serializeAddon);
ws.onclose = function(event) {
  console.log(event);
  terminal.write('\r\n\nconnection has been terminated from the server-side (hit refresh to restart)\n')
};
ws.onopen = function() {
  terminal.loadAddon(attachAddon);
  terminal._initialized = true;
  terminal.focus();
  setTimeout(function() {fitAddon.fit()});
  terminal.onResize(function(event) {
    const rows = event.rows;
    const cols = event.cols;
    const size = JSON.stringify({cols: cols, rows: rows + 1});
    const send = new TextEncoder().encode("\x01" + size);
    console.log('resizing to', size);
    ws.send(send);
  });
  terminal.onTitleChange(function(event) {
    console.log(event);
  });
  window.onresize = function() {
    fitAddon.fit();
  };
};

// ws.onmessage = function (event) {
//   const message = new Uint8Array(event.data);
//   const str = new TextDecoder("utf-8").decode(message);
//   const filteredStr = filterUnsupportedSequences(str);
//
//   console.log("Filtered String:", filteredStr);
//   terminal.write(filteredStr);
// };
//
// function filterUnsupportedSequences(str) {
//   // Regex to match ESC[?2004h and ESC[?2004l sequences
//   const regex = /\x1b\[\?2004[hl]/g;
//   return str.replace(regex, '');
// }