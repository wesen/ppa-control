/**
 * PPA Protocol Encoding/Decoding Functions
 * Based on the Go implementation in lib/protocol/ppa-protocol.go
 */

import {
  BasicHeader,
  MessageType,
  StatusType,
  PresetRecall,
  RecallType,
  LevelType,
} from './types';

/**
 * Create a new BasicHeader with default values
 */
export function createBasicHeader(
  messageType: MessageType,
  status: StatusType,
  deviceUniqueId: Uint8Array,
  sequenceNumber: number,
  componentId: number,
): BasicHeader {
  return {
    messageType,
    protocolId: 1, // Always 1 for PPA protocol
    status,
    deviceUniqueId,
    sequenceNumber,
    componentId,
    reserved: 0,
  };
}

/**
 * Encode BasicHeader to binary format (12 bytes)
 */
export function encodeHeader(header: BasicHeader): ArrayBuffer {
  const buffer = new ArrayBuffer(12);
  const view = new DataView(buffer);

  view.setUint8(0, header.messageType);
  view.setUint8(1, header.protocolId);
  view.setUint16(2, header.status, true); // Little endian
  
  // Device unique ID (4 bytes)
  for (let i = 0; i < 4; i++) {
    view.setUint8(4 + i, header.deviceUniqueId[i] || 0);
  }
  
  view.setUint16(8, header.sequenceNumber, true); // Little endian
  view.setUint8(10, header.componentId);
  view.setUint8(11, header.reserved);

  return buffer;
}

/**
 * Parse BasicHeader from binary data
 */
export function parseHeader(buffer: ArrayBuffer): BasicHeader {
  if (buffer.byteLength < 12) {
    throw new Error('Buffer too short for BasicHeader');
  }

  const view = new DataView(buffer);
  const deviceUniqueId = new Uint8Array(4);
  
  for (let i = 0; i < 4; i++) {
    deviceUniqueId[i] = view.getUint8(4 + i);
  }

  return {
    messageType: view.getUint8(0) as MessageType,
    protocolId: view.getUint8(1),
    status: view.getUint16(2, true) as StatusType, // Little endian
    deviceUniqueId,
    sequenceNumber: view.getUint16(8, true), // Little endian
    componentId: view.getUint8(10),
    reserved: view.getUint8(11),
  };
}

/**
 * Create a new PresetRecall structure
 */
export function createPresetRecall(
  recallType: RecallType,
  option: number,
  indexPosition: number,
): PresetRecall {
  return {
    recallType,
    option,
    indexPosition,
  };
}

/**
 * Encode PresetRecall to binary format (3 bytes)
 */
export function encodePresetRecall(presetRecall: PresetRecall): ArrayBuffer {
  const buffer = new ArrayBuffer(3);
  const view = new DataView(buffer);

  view.setUint8(0, presetRecall.recallType);
  view.setUint8(1, presetRecall.option);
  view.setUint8(2, presetRecall.indexPosition);

  return buffer;
}

/**
 * Parse PresetRecall from binary data
 */
export function parsePresetRecall(buffer: ArrayBuffer): PresetRecall {
  if (buffer.byteLength < 3) {
    throw new Error('Buffer too short for PresetRecall');
  }

  const view = new DataView(buffer);

  return {
    recallType: view.getUint8(0) as RecallType,
    option: view.getUint8(1),
    indexPosition: view.getUint8(2),
  };
}

/**
 * Encode volume value for LiveCmd
 * volume: 0.0 = -80dB, 1.0 = +20dB
 */
export function encodeVolume(volume: number): number {
  const twentyDB = 0x3e8; // +20dB encoded value
  const minusEightyDB = 0x00; // -80dB encoded value
  
  // Clamp volume between 0 and 1
  const clampedVolume = Math.max(0, Math.min(1, volume));
  
  return Math.round(clampedVolume * (twentyDB - minusEightyDB));
}

/**
 * Decode volume value from LiveCmd response
 */
export function decodeVolume(encodedGain: number): number {
  const twentyDB = 0x3e8;
  const minusEightyDB = 0x00;
  
  return encodedGain / (twentyDB - minusEightyDB);
}

/**
 * Create a complete message buffer with header and payload
 */
export function createMessage(header: BasicHeader, payload?: ArrayBuffer): ArrayBuffer {
  const headerBuffer = encodeHeader(header);
  
  if (!payload) {
    return headerBuffer;
  }
  
  const totalLength = headerBuffer.byteLength + payload.byteLength;
  const messageBuffer = new ArrayBuffer(totalLength);
  const messageView = new Uint8Array(messageBuffer);
  
  messageView.set(new Uint8Array(headerBuffer), 0);
  messageView.set(new Uint8Array(payload), headerBuffer.byteLength);
  
  return messageBuffer;
}

/**
 * Create a ping message
 */
export function createPingMessage(
  sequenceNumber: number,
  componentId: number = 0xFF,
): ArrayBuffer {
  const header = createBasicHeader(
    MessageType.Ping,
    StatusType.RequestServer,
    new Uint8Array([0, 0, 0, 0]), // Broadcast device ID
    sequenceNumber,
    componentId,
  );
  
  return createMessage(header);
}

/**
 * Create a volume control message
 */
export function createVolumeMessage(
  volume: number,
  sequenceNumber: number,
  componentId: number = 0xFF,
): ArrayBuffer {
  const header = createBasicHeader(
    MessageType.DeviceData, // Volume uses DeviceData type
    StatusType.CommandClient,
    new Uint8Array([0, 0, 0, 0]),
    sequenceNumber,
    componentId,
  );
  
  // Create volume payload
  const payload = new ArrayBuffer(8);
  const payloadView = new DataView(payload);
  
  // Volume path: [0, Input, 0, Gain] - master volume
  payloadView.setInt8(0, 1);  // 01
  payloadView.setInt8(1, 0);  // 00
  payloadView.setInt8(2, 3);  // 03
  payloadView.setInt8(3, 6);  // 06
  
  // Encoded gain value
  const encodedGain = encodeVolume(volume);
  payloadView.setUint32(4, encodedGain, true); // Little endian
  
  return createMessage(header, payload);
}

/**
 * Create a preset recall message
 */
export function createPresetRecallMessage(
  presetIndex: number,
  sequenceNumber: number,
  componentId: number = 0xFF,
): ArrayBuffer {
  const header = createBasicHeader(
    MessageType.PresetRecall,
    StatusType.CommandClient,
    new Uint8Array([0, 0, 0, 0]),
    sequenceNumber,
    componentId,
  );
  
  const presetRecall = createPresetRecall(
    RecallType.RecallByPresetIndex,
    0,
    presetIndex,
  );
  
  const payload = encodePresetRecall(presetRecall);
  
  return createMessage(header, payload);
}