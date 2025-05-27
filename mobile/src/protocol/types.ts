/**
 * PPA Protocol Type Definitions
 * Based on the Go implementation in lib/protocol/ppa-protocol.go
 */

export enum MessageType {
  Ping = 0,
  LiveCmd = 1,
  DeviceData = 2,
  PresetRecall = 4,
  PresetSave = 5,
}

export enum StatusType {
  RequestClient = 0x0101,
  RequestServer = 0x0102,
  CommandClient = 0x0201,
  CommandServer = 0x0202,
  ResponseClient = 0x0301,
  ResponseServer = 0x0302,
}

export enum LevelType {
  Input = 1,
  Output = 2,
  Eq = 3,
  Gain = 4,
  EqType = 5,
  Quality = 7,
  Active = 8,
  Mute = 9,
  Delay = 10,
  PhaseInversion = 11,
}

export enum RecallType {
  RecallByPresetIndex = 0,
  RecallByPresetPosition = 1,
}

export interface BasicHeader {
  messageType: MessageType;
  protocolId: number;
  status: StatusType;
  deviceUniqueId: Uint8Array;
  sequenceNumber: number;
  componentId: number;
  reserved: number;
}

export interface PresetRecall {
  recallType: RecallType;
  option: number;
  indexPosition: number;
}

export interface DeviceInfo {
  address: string;
  port: number;
  uniqueId: Uint8Array;
  name: string;
  lastSeen: Date;
  isConnected: boolean;
}

export interface ReceivedMessage {
  buffer: ArrayBuffer;
  remoteAddress: string;
  remotePort: number;
  header?: BasicHeader;
}
