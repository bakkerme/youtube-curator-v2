/**
 * Polyfills for MSW v2 compatibility in Jest environment
 */

// Import fetch polyfill
require('whatwg-fetch')

// Polyfill TextEncoder/TextDecoder using Node.js util
const { TextEncoder, TextDecoder } = require('util')

// Set up global objects that MSW v2 expects
Object.assign(global, {
  TextEncoder,
  TextDecoder,
  Request: global.Request || require('whatwg-fetch').Request,
  Response: global.Response || require('whatwg-fetch').Response,
  Headers: global.Headers || require('whatwg-fetch').Headers,
})

// Additional Node.js globals that might be needed
if (typeof global.ReadableStream === 'undefined') {
  global.ReadableStream = require('stream/web').ReadableStream
}

if (typeof global.WritableStream === 'undefined') {
  global.WritableStream = require('stream/web').WritableStream
}

if (typeof global.TransformStream === 'undefined') {
  global.TransformStream = require('stream/web').TransformStream
}

// BroadcastChannel polyfill for Jest environment
if (typeof global.BroadcastChannel === 'undefined') {
  global.BroadcastChannel = class BroadcastChannel {
    constructor(name) {
      this.name = name
    }
    
    postMessage(message) {
      // No-op in test environment
    }
    
    addEventListener(type, listener) {
      // No-op in test environment
    }
    
    removeEventListener(type, listener) {
      // No-op in test environment
    }
    
    close() {
      // No-op in test environment
    }
  }
} 