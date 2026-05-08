import { WSMessage } from "@/types";

export class WSClient {
  private ws: WebSocket | null = null;
  private url: string;
  private roomId: string;
  private userId: string;
  private clientId: string = "";
  private messageHandlers: Map<string, (msg: WSMessage) => void> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  constructor(url: string, roomId: string, userId: string) {
    this.url = url;
    this.roomId = roomId;
    this.userId = userId;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        const wsUrl = `${this.url}?roomId=${this.roomId}&userId=${this.userId}`;
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          console.log("[WSClient] Connected to signaling server");
          this.reconnectAttempts = 0;
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const msg: WSMessage = JSON.parse(event.data);
            this.clientId = msg.from;
            this.dispatchMessage(msg.type, msg);
          } catch (e) {
            console.error("[WSClient] Failed to parse message:", e);
          }
        };

        this.ws.onerror = (error) => {
          console.error("[WSClient] WebSocket error:", error);
          reject(error);
        };

        this.ws.onclose = () => {
          console.log("[WSClient] Disconnected from signaling server");
          this.attemptReconnect();
        };
      } catch (e) {
        reject(e);
      }
    });
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      console.log(`[WSClient] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
      setTimeout(() => this.connect().catch(console.error), delay);
    }
  }

  send(msg: WSMessage) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error("[WSClient] WebSocket not connected");
      return;
    }
    this.ws.send(JSON.stringify(msg));
  }

  on(type: string, handler: (msg: WSMessage) => void) {
    this.messageHandlers.set(type, handler);
  }

  off(type: string) {
    this.messageHandlers.delete(type);
  }

  private dispatchMessage(type: string, msg: WSMessage) {
    const handler = this.messageHandlers.get(type);
    if (handler) {
      handler(msg);
    }
  }

  getClientId(): string {
    return this.clientId;
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}
