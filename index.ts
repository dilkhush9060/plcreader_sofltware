import { io } from "socket.io-client";

const SERVER_URL = "http://localhost:5000";

const socket = io(SERVER_URL, {
  transports: ["websocket"], // Use only WebSocket
  reconnection: true, // Auto Reconnect
  reconnectionDelay: 5000, // Reconnect every 5 sec
});

socket.on("connect", () => {
  console.log("✅ Connected to Server:", socket.id);
});

socket.on("disconnect", () => {
  console.log("❌ Disconnected");
});

socket.on("realtime", (data) => {
  console.log("📡 Realtime Data:", data);
});

// Send Data Every 20 Sec
setInterval(() => {
  socket.emit("realtime", {
    id: socket.id,
    value: Math.random() * 100,
    timestamp: new Date(),
  });
  console.log("📤 Sent Data");
}, 2000);
