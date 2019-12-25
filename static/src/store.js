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
const ADD_PEER = 8

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
  streams = []

  async init() {
    try {
      const [token, user] = await getAuth()
      this.webrtc = new WebRTC(getVideoElement, getRemoteVideoElement, this.updatePeerStatus, () => this.set({}))
      const connection = new Connection(token)
      connection.on(TEXT_TYPE, this.onTextMessage)
      connection.on(ONOPEN, this.onOpenConnection)
      connection.on(ONCLOSE, this.onCloseConnection)
      connection.on(ADD_PEER, this.onAddPeer)
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
      this.connection.send({ data: { peerId: this.connectionId, id: room.id }, to: room.owner, type: JOIN_ROOM })
    } catch (e) {
      console.log('join room error', e)
      location.search = ''
    }
  }
  onAddPeer = async msg => {
    try {
      const { data } = msg
      const peerId = data.peerId
      const offer = await this.webrtc.connectPeer(this.connectionId, peerId, this.sendCandidate)
      this.connection.send({ data: offer, to: peerId, type: SDP })
    } catch (e) {
      console.log('add peer error', e)
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
  onCandidate = async msg => {
    const { data, from } = msg
    await this.webrtc.onCandidate(from, data)
  }
  updatePeerStatus = (peerId, status) => {

  }
  dropClient = async clientIndex=>{
    const [peerId, offer] = await this.webrtc.dropClient(clientIndex)
    this.connection.send({ data: offer, to: peerId, type: SDP })
  }
  onMp3FileLoad = async e => {
    const files = Array.from(e.target.files)
    const streams = await Promise.all(files.map(file => this.getAudioFromFileStream(file)))
    const [peerId, offer] = await this.webrtc.addLocalAudioTracks(streams)
    this.connection.send({ data: offer, to: peerId, type: SDP })
  }

  getAudioFromFileStream = file => {
    return new Promise((resolve, reject) => {
      const context = new AudioContext();
      const gainNode = context.createGain();
      gainNode.connect(context.destination);
      // don't play for self
      gainNode.gain.value = 0;

      const reader = new FileReader();
      reader.onload = ((e) => {
        // Import callback function that provides PCM audio data decoded as an audio buffer
        context.decodeAudioData(e.target.result, (buffer) => {
          // Create the sound source
          const soundSource = context.createBufferSource();

          soundSource.buffer = buffer;
          soundSource.start(0, 0 / 1000);
          soundSource.connect(gainNode);

          const destination = context.createMediaStreamDestination();
          soundSource.connect(destination);

          // destination.stream
          console.log('-', destination.stream)
          resolve([destination.stream, file.name])
        });
      });

      reader.readAsArrayBuffer(file);
    })
  }

  onMute = () => {
    this.webrtc.toggleMicrophoneMute()
  }
  onRefresh = () => {
    this.webrtc.stop()
    const { room } = this.get()
    this.connection.send({ data: { peerId: this.connectionId, id: room.id }, to: room.owner, type: JOIN_ROOM })
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
