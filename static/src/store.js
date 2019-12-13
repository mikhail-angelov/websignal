import { getAuth } from './auth.js'
import { getRooms, createRoom } from './rooms.js'
import { getId } from './utils.js'
import { Connection, ONOPEN } from './connection.js'
import * as webrtc from './webrtc.js'

const TEXT_TYPE = 0

export class Store {
  data = {
    authenticated: false,
    userId: '',
    username: '',
    avatar: '',
    room: null,
    connection: null,
    messages: [],
    message: '',
  }
  listeners = []

  async init() {
    try {
      const [token, user] = await getAuth()
      const connection = new Connection(token)
      connection.on(TEXT_TYPE, this.onTextMessage)
      connection.on(ONOPEN, this.onOpenConnection)
      connection.connect()
      this.set({ authenticated: true, username: user.name, avatar: user.avatar, userId: user.id, connection })
    } catch (e) {
      console.log('invalid auth:', e)
      this.set({ authenticated: false })
    }
  }
  onToggleConference = () => {
    const { room } = this.get()
    if (room) {
      this.stopConference()
    } else {
      this.startConference()
    }
  }
  onOpenConnection() {
    //get Rooms
    getRooms()
  }
  async startConference() {
    try {
      const room = await createRoom()
      const conferenceLink = `${location.href}?room=${room.id}`
      webrtc.start(room.id)
      this.set({ room, conferenceLink })
    } catch (e) {
      console.log('create room error', e)
    }
  }
  stopConference() {
    this.set({ room: null, conferenceLink: '' })
    webrtc.stop()
  }
  onMessageChange = e =>{
    this.set({ message: e.target.value})  
  }
  onTextMessage = message => {
    console.log('new message', message)
    this.set({ messages: [...this.get().messages, message] })
  }
  onSendMessage = () => {
    const { message, userId, connection } = this.get()
    if ((message, connection)) {
      connection.send({ data: message, from: userId, type: TEXT_TYPE })
      this.set({ message: '' })
    }
  }

  on(cb) {
    this.listeners = [...this.listeners, cb]
  }
  off(cb) {
    this.listeners = this.listeners.filter(item => item === cb)
  }
  get() {
    return this.data
  }
  set(value) {
    this.data = { ...this.data, ...value } //todo validate input param
    this.listeners.forEach(cb => cb(this.data))
  }
}
