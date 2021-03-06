import { getId } from './utils.js'
const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
const encoder = new TextEncoder()
const decoder = new TextDecoder('utf-8')

const CONNECT_TIMEOUT = 1000
export const ONOPEN = 'ON_OPEN_CONNECTION'
export const ONCLOSE = 'ON_CLOSE_CONNECTION'

export class Connection {
  socket
  listeners
  constructor(token, id) {
    this.token = token
    this.id = id
    this.listeners = {}
    this.messageQueue = []
    this.connectTimer = null
  }

  on(messageType, cb) {
    this.listeners[messageType] = cb
  }

  connect = () => {
    const connectionId = getId()
    const socket = new WebSocket(`${protocol}//${location.host}/ws?token=${this.token}&id=${connectionId}`)
    socket.binaryType = 'arraybuffer' //to support binary messages
    this.socket = socket

    console.log('Attempting Connection...')

    socket.onopen = () => {
      console.log('Successfully Connected')
      const listener = this.listeners[ONOPEN]
      if (listener) {
        listener()
      }
    }

    socket.onclose = event => {
      console.log('Socket Closed Connection: ', event)
      const listener = this.listeners[ONCLOSE]
      if (listener) {
        listener()
      }
    }

    socket.onerror = error => {
      console.log('Socket Error: ', error)
    }

    socket.onmessage = event => {
      try {
        const message = JSON.parse(decoder.decode(new Uint8Array(event.data).buffer))
        console.log('Socket on message ', message)
        const listener = this.listeners[message.type]
        if (listener) {
          listener(message)
        } else {
          console.log('no listener for', message.type)
        }
      } catch (e) {
        console.log('onmessage error', e)
      }
    }
    return connectionId
  }

  push = (message) => {
    this.socket.send(encoder.encode(JSON.stringify(message)))
  }
  send = (message) => {
    if (this.socket && this.socket.readyState) {
      this.messageQueue.forEach(msg => {
        this.push(msg)
      })
      this.messageQueue = []
      if (message) {
        this.push(message)
      }
    } else {
      console.log('sen message queued, no open socket')
      this.messageQueue.push(message)
      if (this.connectTimer) {
        clearTimeout(this.connectTimer)
      }
      this.connectTimer = setTimeout(this.send, CONNECT_TIMEOUT)
    }
  }
}
