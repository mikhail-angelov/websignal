const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
const encoder = new TextEncoder()
const decoder = new TextDecoder('utf-8')

export const ONOPEN = 'ON_OPEN_CONNECTION'

export class Connection {
  socket
  listeners
  constructor(token, id) {
    this.token = token
    this.id = id
    this.listeners = {}
  }

  on(messageType, cb) {
    this.listeners[messageType] = cb
  }

  connect() {
    const socket = new WebSocket(`${protocol}//${location.host}/ws?token=${this.token}`)
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
    }

    socket.onerror = error => {
      console.log('Socket Error: ', error)
    }

    socket.onmessage = event => {
      try {
        console.log('Socket on message ', event)
        const message = JSON.parse(decoder.decode(new Uint8Array(event.data).buffer))
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
  }

  send(message) {
    if (this.socket) {
      this.socket.send(encoder.encode(JSON.stringify(message)))
    } else {
      console.log('sen message error, no open socket')
    }
  }
}
