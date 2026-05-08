export interface WSMessage {
  type: "offer" | "answer" | "ice" | "join" | "leave" | "chat";
  room_id: string;
  from: string;
  to?: string;
  payload?: unknown;
}

export interface IcePayload {
  candidate: string;
  sdpMLineIndex: number | null;
  sdpMid: string | null;
}

export interface ChatPayload {
  message: string;
  sender: string;
}

export interface ChatEntry extends ChatPayload {
  from: string;
  timestamp: Date;
}

export interface Peer {
  clientId: string;
  pc: RTCPeerConnection;
  stream?: MediaStream;
}
