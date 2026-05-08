import { WSClient } from "./ws";
import { IcePayload, Peer, WSMessage } from "@/types";

export class WebRTCManager {
  private wsClient: WSClient;
  private localStream: MediaStream | null = null;
  private peers: Map<string, Peer> = new Map();
  private iceServers: RTCIceServer[] = [
    { urls: "stun:stun.l.google.com:19302" },
    { urls: "stun:stun1.l.google.com:19302" },
  ];

  private onRemoteStreamCallback?: (clientId: string, stream: MediaStream) => void;
  private onPeerLeftCallback?: (clientId: string) => void;

  constructor(wsClient: WSClient) {
    this.wsClient = wsClient;
    this.setupSignalingHandlers();
  }

  async initLocalStream(audio = true, video = true): Promise<MediaStream> {
    try {
      this.localStream = await navigator.mediaDevices.getUserMedia({ audio, video });
      console.log("[WebRTC] Local stream initialized");
      return this.localStream;
    } catch (error) {
      console.error("[WebRTC] Failed to get local stream:", error);
      throw error;
    }
  }

  getLocalStream(): MediaStream | null {
    return this.localStream;
  }

  private setupSignalingHandlers() {
    this.wsClient.on("join", (msg: WSMessage) => {
      console.log("[WebRTC] Peer joined:", msg.from);
      if (msg.from !== this.wsClient.getClientId()) {
        this.createPeerConnection(msg.from, true);
      }
    });

    this.wsClient.on("offer", async (msg: WSMessage) => {
      console.log("[WebRTC] Received offer from:", msg.from);
      const pc = this.createPeerConnection(msg.from, false);
      const payload = msg.payload as { type: string; sdp: string };
      await pc.setRemoteDescription(new RTCSessionDescription(payload));
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);
      this.wsClient.send({
        type: "answer",
        room_id: "",
        from: this.wsClient.getClientId(),
        to: msg.from,
        payload: answer,
      });
    });

    this.wsClient.on("answer", async (msg: WSMessage) => {
      console.log("[WebRTC] Received answer from:", msg.from);
      const peer = this.peers.get(msg.from);
      if (peer) {
        const payload = msg.payload as { type: string; sdp: string };
        await peer.pc.setRemoteDescription(new RTCSessionDescription(payload));
      }
    });

    this.wsClient.on("ice", async (msg: WSMessage) => {
      const peer = this.peers.get(msg.from);
      if (peer) {
        const payload = msg.payload as IcePayload;
        const candidate = new RTCIceCandidate({
          candidate: payload.candidate,
          sdpMLineIndex: payload.sdpMLineIndex ?? undefined,
          sdpMid: payload.sdpMid ?? undefined,
        });
        await peer.pc.addIceCandidate(candidate);
      }
    });

    this.wsClient.on("leave", (msg: WSMessage) => {
      console.log("[WebRTC] Peer left:", msg.from);
      this.removePeer(msg.from);
    });
  }

  private createPeerConnection(clientId: string, initiator: boolean): RTCPeerConnection {
    if (this.peers.has(clientId)) {
      return this.peers.get(clientId)!.pc;
    }

    const pc = new RTCPeerConnection({ iceServers: this.iceServers });

    // Add local tracks
    if (this.localStream) {
      this.localStream.getTracks().forEach((track) => {
        pc.addTrack(track, this.localStream!);
      });
    }

    // Handle remote stream
    pc.ontrack = (event) => {
      console.log("[WebRTC] Remote track received from:", clientId);
      const [stream] = event.streams;
      const peer = this.peers.get(clientId);
      if (peer) {
        peer.stream = stream;
        this.onRemoteStreamCallback?.(clientId, stream);
      }
    };

    // Handle ICE candidates
    pc.onicecandidate = (event) => {
      if (event.candidate) {
        this.wsClient.send({
          type: "ice",
          room_id: "",
          from: this.wsClient.getClientId(),
          to: clientId,
          payload: {
            candidate: event.candidate.candidate,
            sdpMLineIndex: event.candidate.sdpMLineIndex,
            sdpMid: event.candidate.sdpMid,
          },
        });
      }
    };

    pc.onconnectionstatechange = () => {
      console.log(`[WebRTC] Connection state with ${clientId}:`, pc.connectionState);
    };

    this.peers.set(clientId, { clientId, pc });

    // If initiator, create and send offer
    if (initiator) {
      pc.createOffer()
        .then((offer) => pc.setLocalDescription(offer))
        .then(() => {
          this.wsClient.send({
            type: "offer",
            room_id: "",
            from: this.wsClient.getClientId(),
            to: clientId,
            payload: pc.localDescription,
          });
        })
        .catch((err) => console.error("[WebRTC] Failed to create offer:", err));
    }

    return pc;
  }

  private removePeer(clientId: string) {
    const peer = this.peers.get(clientId);
    if (peer) {
      peer.pc.close();
      this.peers.delete(clientId);
      this.onPeerLeftCallback?.(clientId);
    }
  }

  onRemoteStream(callback: (clientId: string, stream: MediaStream) => void) {
    this.onRemoteStreamCallback = callback;
  }

  onPeerLeft(callback: (clientId: string) => void) {
    this.onPeerLeftCallback = callback;
  }

  cleanup() {
    this.peers.forEach((peer) => peer.pc.close());
    this.peers.clear();
    if (this.localStream) {
      this.localStream.getTracks().forEach((track) => track.stop());
      this.localStream = null;
    }
  }
}
