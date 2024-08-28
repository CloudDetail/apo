//skywalking traceid
import { Buffer } from 'buffer'
import { parse, stringify as stringifyUUID } from 'uuid'

export function swTraceIDToTraceID(traceID) {
  if (traceID.length <= 36) {
    try {
      let uuid = parseUUID(traceID)
      return uint8ArrayToHexString(uuid)
    } catch (err) {
      return null
    }
  }
  return uint8ArrayToHexString(swStringToUUID(traceID, 0))
}

function parseUUID(uuidStr) {
  let hexStr = uuidStr.replace(/-/g, '')
  if (hexStr.length !== 32) {
    throw new Error('Invalid UUID')
  }
  let bytes = new Uint8Array(16)
  for (let i = 0; i < 16; i++) {
    bytes[i] = parseInt(hexStr.slice(i * 2, i * 2 + 2), 16)
  }
  return bytes
}

function swStringToUUID(s, extra) {
  if (s.length < 32) {
    return new Uint8Array(16)
  }

  let uid = parseUUID(s.substring(0, 32))
  for (let i = 0; i < 4; i++) {
    uid[i] ^= extra & 0xff
    extra >>= 8
  }

  if (s.length === 32) {
    return uid
  }

  let parts = s.split('.')
  console.log(parts)
  if (parts.length !== 3) {
    return uid
  }

  let mid = parseInt(parts[1], 10)
  let last = BigInt(parts[2])
  console.log(last)
  mid = mid >>> 0

  for (let i = 4; i < 8; i++) {
    uid[i] ^= mid & 0xff
    mid >>>= 8
  }

  for (let i = 8; i < 16; i++) {
    if (i === 8) {
      console.log(last, last & 0xffn)
    }
    uid[i] ^= Number(last & 0xffn)
    last >>= 8n
  }

  return uid
}

function uint8ArrayToHexString(uint8Array) {
  return Array.from(uint8Array, (byte) => byte.toString(16).padStart(2, '0')).join('')
}
