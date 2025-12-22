// Minimal WebSocket client wrapper compatible with the backend's raw websocket API

class WSClient {
  constructor(url) {
    this.url = url;
    this.ws = null;
    this.handlers = new Map(); // event -> Set of handlers
    this.connected = false;
    this.reconnectInterval = 2000;
    this._shouldReconnect = false;
  }

  connect() {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    this.ws = new WebSocket(this.url);

    this.ws.addEventListener("open", () => {
      this.connected = true;
      this._emitLocal("connect");
    });

    this.ws.addEventListener("close", () => {
      this.connected = false;
      this._emitLocal("disconnect");
      if (this._shouldReconnect) {
        setTimeout(() => this.connect(), this.reconnectInterval);
      }
    });

    this.ws.addEventListener("message", (ev) => this._onMessage(ev));

    this.ws.addEventListener("error", (err) => {
      // bubble up as generic error event
      this._emitLocal("error", err);
    });
  }

  close() {
    this._shouldReconnect = false;
    if (this.ws) this.ws.close();
  }

  sendEvent(event, data) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return false;
    const envelope = { event, data };
    try {
      this.ws.send(JSON.stringify(envelope));
      return true;
    } catch (err) {
      console.error("ws send failed", err);
      return false;
    }
  }

  on(event, handler) {
    const set = this.handlers.get(event) || new Set();
    set.add(handler);
    this.handlers.set(event, set);
  }

  off(event, handler) {
    const set = this.handlers.get(event);
    if (!set) return;
    set.delete(handler);
    if (set.size === 0) this.handlers.delete(event);
  }

  _emitLocal(event, ...args) {
    const set = this.handlers.get(event);
    if (!set) return;
    for (const h of Array.from(set)) {
      try {
        h(...args);
      } catch (err) {
        console.error("ws handler error", err);
      }
    }
  }

  _onMessage(ev) {
    let payload = ev.data;
    // payload may be a string or already JSON
    try {
      payload = JSON.parse(payload);
      // some server writes JSON inside a JSON string (double-encoded) -> try to parse again
      if (typeof payload === "string") {
        try {
          payload = JSON.parse(payload);
        } catch (e) {
          // leave as string
        }
      }
    } catch (e) {
      // not JSON -> ignore
      return;
    }

    // handle server-side broadcast envelope produced by backend
    // Backend sends BroadcastMessage JSON like: { Type: "direct", UserID: "...", Payload: { "type": "message:new", "message": {...} } }

    // normalize key casing (in case server uses uppercase keys)
    const normalized = {};
    for (const k in payload) {
      normalized[k.toLowerCase()] = payload[k];
    }

    // If we have a nested payload (hub broadcast), extract it
    const inner = normalized.payload || normalized.Payload || payload.payload || payload.Payload || payload;

    // If inner has a `type` field, emit that type event with inner payload
    if (inner && typeof inner === "object" && (inner.type || inner.Type)) {
      const type = inner.type || inner.Type;
      // prefer to pass the `message` or the whole payload depending on shape
      const msg = inner.message !== undefined ? inner.message : inner;
      this._emitLocal(type, msg);
      return;
    }

    // If the top level has `type` (older messages), emit it
    if (payload.type) {
      this._emitLocal(payload.type, payload);
      return;
    }

    // If server sends a welcome server message shape { sender, message }
    if (payload.sender && payload.message) {
      this._emitLocal("server:message", payload);
      return;
    }

    // Fallback: emit generic message
    this._emitLocal("message", payload);
  }
}

// create a singleton instance (use ws:// or wss:// depending on page origin)
const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
const host = window.__API_HOST__ || window.location.hostname + ":8080"; // fallback
export const socket = new WSClient(`${protocol}//${host}/api/v1/ws`);
