import { getAuth } from './auth.js'
import { getRooms, joinRoom, createRoom } from './rooms.js'
import { getId } from './utils.js'
import { Connection, ONOPEN, ONCLOSE } from './connection.js'
import { WebRTC } from './webrtc.js'

const getVideoElement = () => document.getElementById('video')
const getRemoteVideoElement = () => document.getElementById('remote-0')

const TEXT_TYPE = 0
const CREATE_ROOM = 1
const JOIN_ROOM = 2
const LEAVE_ROOM = 3
const SDP = 4
const CANDIDATE = 5
const GET_ROOM = 6
const ROOM_IS_CREATED = 7
const UPDATE_ROOM = 8
const START_PEER_CONNECTION = 9
const ADD_FAKE_USER = 10
const REMOVE_FAKE_USER = 11

export class Store {
  data = {
    authenticated: false,
    hideMessages: true,
    userId: '',
    username: '',
    avatar: '',
    room: null,
    messages: [],
    users: [],
    message: '',
    broadcaster: false,
    muted: true,
    fakeUsers: [],
  }
  webrtc = null
  connection = null
  connectionId = null
  listeners = []
  streams = []

  async init() {
    try {
      const [token, user] = await getAuth()
      this.webrtc = new WebRTC(getVideoElement, getRemoteVideoElement, this.updatePeerStatus, () => this.set({}))
      const connection = new Connection(token)
      connection.on(TEXT_TYPE, this.onTextMessage)
      connection.on(ONOPEN, this.onOpenConnection)
      connection.on(ONCLOSE, this.onCloseConnection)
      connection.on(UPDATE_ROOM, this.onUpdateRoom)
      connection.on(START_PEER_CONNECTION, this.onStartPeerConnection)
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
      setInterval(() => {
        this.webrtc.getSoundIndicator(0)
        this.webrtc.getSoundIndicator(1)
      }, 10000)
    } catch (e) {
      console.log('invalid auth:', e)
      this.set({ authenticated: false })
    }
  }
  toggleMessages = () => {
    const { hideMessages } = this.get()
    this.set({ hideMessages: !hideMessages })
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
  onOpenConnection = () => {
    //get Rooms
    getRooms()
  }
  onCloseConnection = () => {
    this.set({ authenticated: false })
  }
  async startConference() {
    this.connection.send({ data: {}, type: CREATE_ROOM })
  }
  stopConference() {
    this.set({ room: null, conferenceLink: '' })
    this.webrtc.stop()
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
      this.connection.send({ data: { peerId: this.connectionId, id: id }, to: "me", type: JOIN_ROOM })
    } catch (e) {
      console.log('join room error', e)
      location.search = ''
    }
  }
  onUpdateRoom = async msg => {
    try {
      const { data } = msg
      const users = data.users
        .filter(user => user.peerId !== this.connectionId)
        .map(user => ({ ...user, picture: user.pictureUrl || `data:image/png;base64,${user.picture}` }))
      this.set({ room: data, messages: data.messages, users })
    } catch (e) {
      console.log('add peer error', e)
      location.search = ''
    }
  }
  onStartPeerConnection = async msg => {
    try {
      const { data } = msg
      const peerId = data.peerId
      const offer = await this.webrtc.connectPeer(this.connectionId, peerId, this.sendCandidate, this.getUsers)
      this.connection.send({ data: offer, to: peerId, type: SDP })
    } catch (e) {
      console.log('add peer error', e)
      location.search = ''
    }
  }
  onSDP = async msg => {
    const { data, from } = msg
    if (data && data.sdp) {
      const answer = await this.webrtc.onSDP(data, from, this.sendCandidate, this.getUsers)
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
  onCandidate = async msg => {
    const { data, from } = msg
    await this.webrtc.onCandidate(from, data)
  }
  getUsers = () => {
    const { room } = this.get()
    return room.users
  }
  updatePeerStatus = (peerId, status) => {
    console.log('-update peer:', peerId, status)
  }
  dropClient = async userId => {
    const { fakeUsers, room } = this.get()
    this.set({ fakeUsers: fakeUsers.filter(id => id != userId) })
    await this.webrtc.dropClient(userId, (peerId, offer) => {
      this.connection.send({ data: offer, to: peerId, type: SDP })
    })
    this.connection.send({ type: REMOVE_FAKE_USER, to: 'all', data: { roomId: room.id, id: userId } })
    this.set({ fakeUsers })
  }
  onMp3FileLoad = async e => {
    const { fakeUsers, room } = this.get()
    const files = Array.from(e.target.files)
    const tracks = await Promise.all(files.map(file => this.webrtc.getAudioFromFileStream(file)))
    tracks.forEach(([track, file]) => {
      fakeUsers.push(track.id)
      // set pending user id before track is added to peer
      // this is a hack, since I did not find the way how to pass extra information along with media track for the same peer
      this.connection.send({ type: ADD_FAKE_USER, to: 'all', data: { roomId: room.id, id: track.id, name: file, pictureUrl: '/assets/file.png' } })
    })
    this.webrtc.addLocalAudioTracks(tracks)
    await this.webrtc.refreshPeersSdp((peerId, offer) =>
      this.connection.send({ data: offer, to: peerId, type: SDP }))
    this.set({ fakeUsers })
  }

  onMute = () => {
    const { muted } = this.get()
    this.webrtc.toggleMicrophoneMute()
    this.set({ muted: !muted })
  }
  onRefresh = () => {
    this.webrtc.stop()
    const { room } = this.get()
    this.connection.send({ data: { peerId: this.connectionId, id: room.id }, to: room.owner, type: JOIN_ROOM })
  }
  setPendingUser = async msg => {
    const { data } = msg
    this.webrtc.setPendingUserId(data.id)
  }

  //messaging
  onTextMessage = msg => {
    const { data } = msg
    console.log('new message', data)
    this.set({ messages: [...this.get().messages, data] })
  }
  onMessageChange = e => {
    this.set({ message: e.target.value })
  }
  onSendMessage = () => {
    const { message, userId, room } = this.get()
    if ((message, this.connection)) {
      this.connection.send({ data: { text: message, id: room.id }, from: userId, type: TEXT_TYPE })
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
