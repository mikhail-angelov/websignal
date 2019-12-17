import { getAuth } from './auth.js'
import { getRooms, joinRoom, createRoom } from './rooms.js'
import { getId } from './utils.js'
import { Connection, ONOPEN } from './connection.js'
import { WebRTC } from './webrtc.js'

const getVideoElement = () => document.getElementById('video')
const getRemoteVideoElement = () => document.getElementById('video-remote')

const TEXT_TYPE = 0
const CREATE_ROOM = 1
const JOIN_ROOM = 2
const LEAVE_ROOM = 3
const SDP = 4
const CANDIDATE = 5
const GET_ROOM = 6
const ROOM_IS_CREATED = 7

export class Store {
  data = {
    authenticated: false,
    userId: '',
    username: '',
    avatar: '',
    room: null,
    messages: [],
    message: '',
    broadcaster: false,
  }
  webrtc = null
  connection = null
  connectionId = null
  listeners = []

  async init() {
    try {
      const [token, user] = await getAuth()
      this.webrtc = new WebRTC(getVideoElement, getRemoteVideoElement, this.updatePeerStatus)
      const connection = new Connection(token)
      connection.on(TEXT_TYPE, this.onTextMessage)
      connection.on(ONOPEN, this.onOpenConnection)
      connection.on(SDP, this.onSDP)
      connection.on(CANDIDATE, this.onCandidate)
      connection.on(ROOM_IS_CREATED, this.onRoomIsCreated)
      this.connectionId = connection.connect()
      this.connection = connection
      this.set({ authenticated: true, username: user.name, avatar: user.avatar, userId: user.id })
      const roomId = this.getRoomId()
      if (roomId) {
        this.onJoinRoom(roomId)
      }
    } catch (e) {
      console.log('invalid auth:', e)
      this.set({ authenticated: false })
    }
  }
  getRoomId = () => {
    const queryString = location.search
    const pairs = (queryString[0] === '?' ? queryString.substr(1) : queryString).split('&')
    const roomPair = pairs.find(pair => pair.indexOf('room=') === 0)
    return roomPair ? roomPair.split('=')[1] : null
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
    this.connection.send({ data: {}, type: CREATE_ROOM })
  }
  stopConference() {
    this.set({ room: null, conferenceLink: '' })
    this.webrtc.stop()
  }
  onMessageChange = e => {
    this.set({ message: e.target.value })
  }
  onTextMessage = message => {
    console.log('new message', message)
    this.set({ messages: [...this.get().messages, message] })
  }
  onSendMessage = () => {
    const { message, userId } = this.get()
    if ((message, this.connection)) {
      this.connection.send({ data: message, from: userId, type: TEXT_TYPE })
      this.set({ message: '' })
    }
  }
  onRoomIsCreated = msg => {
    try {
      const { data: room } = msg
      const conferenceLink = `${location.origin}?room=${room.id}`
      this.webrtc.start(room.id)
      this.set({ room, conferenceLink, broadcaster: true })
    } catch (e) {
      console.log('create room error', e)
    }
  }
  onJoinRoom = async id => {
    try {
      const room = await joinRoom(id)
      this.set({ room })
      const peerId = room.owner
      const offer = await this.webrtc.connectPeer(this.connectionId, peerId, this.sendCandidate)
      this.connection.send({ data: offer, to: room.owner, type: SDP })
    } catch (e) {
      console.log('join room error', e)
      location.search = ''
    }
  }
  onSDP = async msg => {
    const { data, from } = msg
    if (data && data.sdp) {
      const answer = await this.webrtc.onSDP(data, from, this.sendCandidate)
      if (answer) {
        //do we need to send answer?
        this.connection.send({ data: answer, to: from, type: SDP })
      }
    } else {
      console.log('-invalid SDP message', msg)
    }
  }
  sendCandidate = (peerId, candidate) => {
    this.connection.send({ data: candidate, to: peerId, type: CANDIDATE })
  }
  onCandidate = msg => {
    const { data, from } = msg
    this.webrtc.onCandidate(from, data)
  }
  updatePeerStatus = (peerId, status)=>{

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
