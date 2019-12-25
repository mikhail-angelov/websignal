// const getScreenShareStream = () => navigator.getDisplayMedia({ video: true })

const servers = [{ "url": "stun:stun.l.google.com:19302" }]

export class WebRTC {
  constructor(getVideoElement, getRemoteVideoElement, updateStatus, refreshUI) {
    this.peers = {}
    this.localStream = null
    this.getVideoElement = getVideoElement
    this.getRemoteVideoElement = getRemoteVideoElement
    this.updateStatus = updateStatus
    this.refreshUI = refreshUI
    this.mic = null
    this.tracks = []
    this.trackId = 0
    this.streams = []
    this.clients = []
  }

  start = async () => {
    // this.localStream = await navigator.mediaDevices.getUserMedia({ video: true })
    // const video = this.getVideoElement()
    // video.srcObject = this.localStream
  }

  stop = () => {
    console.log('Ending call')
    if (this.localStream) {
      this.localStream.getTracks().forEach((track) => track.stop())
      this.localStream = null
    }
    this.streams = [] //todo investigate how to clear it correctly
    Object.keys(this.peers).forEach((id) => {
      const pc = this.peers[id]
      const localStream = pc.getLocalStreams()[0]
      if (localStream) {
        console.log('closing stream: ', localStream.id)
        pc.removeStream(localStream)
        localStream.getTracks().forEach(track => {
          track.stop()
          localStream.removeTrack(track)
        });
      }
      const remoteStream = pc.getRemoteStreams()[0]
      if (remoteStream) {
        console.log('closing stream: ', remoteStream.id)
        pc.removeStream(remoteStream)
        remoteStream.getTracks().forEach(track => {
          track.stop()
          remoteStream.removeTrack(track)
        });
      }
      pc.onicecandidate = null
      pc.ontrack = null
      pc.oniceconnectionstatechange = null
      pc.close()
      const video = this.getVideoElement()
      if (video) video.srcObject = null
      const remoteVideo = this.getRemoteVideoElement()
      if (remoteVideo) remoteVideo.srcObject = null
    })
    this.peers = {}
  }
  createPeer(remotePeerId, sendCandidate) {
    const pc = new RTCPeerConnection({ iceServers: servers })
    let track,
      tr = null
    pc.peerId = remotePeerId
    pc.onicecandidate = ({ candidate }) => sendCandidate(remotePeerId, candidate)
    pc.ontrack = (event) => {
      console.log('-str-', event.streams, event.track)
      if (event.track) {
        const stream = new MediaStream()
        stream.addTrack(event.track)
        this.streams.push(stream)
        console.log('Received remote track')

      } else if (event.streams && event.streams.length > 0) {
        this.streams = [...this.streams, ...event.streams]
        console.log('Received remote stream')
      }
      this.refreshUI()
    }
    pc.oniceconnectionstatechange = (event) => {
      console.log('oniceconnectionstatechange:', event)
      if (event.currentTarget.iceConnectionState === 'closed') {
        this.updateStatus(remotePeerId, 'disconnected', tr)
        track.stop() //need investigate it
      }
    }
    pc.onnegotiationnedded = (event) => {
      console.log('onnegotiationnedded:', event)
    }
    return pc
  }

  // connect peer for broadcaster
  connectPeer = async (id, remotePeerId, sendCandidate) => {
    await this.start()
    const pc = this.createPeer(remotePeerId, sendCandidate)
    this.peers[remotePeerId] = pc
    //add mic tracks
    const micTrack = await this.getMicrophoneTrack()
    pc.addTrack(micTrack)
    //and others
    if (this.tracks.length > 0) {
      this.tracks.forEach(track => pc.addTrack(track))
    }

    console.log('-local streams-', pc.getLocalStreams())
    //
    const offer = await pc.createOffer()
    console.log('Offer from pc \n' + offer.sdp)
    await pc.setLocalDescription(offer)
    offer.id = id
    return offer
  }

  onSDP = async (sdp, remotePeerId, sendCandidate) => {
    console.log('onSDP', sdp.type)
    if (sdp.type === 'answer') {
      //process answer
      const pc = this.peers[remotePeerId]
      if (pc) {
        pc.setRemoteDescription(sdp)
      } else {
        console.error('invalid sdp message peer does not exist: ', remotePeerId)
      }
    } else if (sdp.type === 'offer') {
      // connect peer for clients
      const pc = this.peers[remotePeerId] || this.createPeer(remotePeerId, sendCandidate)

      //add mic tracks
      const micTrack = await this.getMicrophoneTrack()
      pc.addTrack(micTrack)

      console.log('-local streams-', pc.getLocalStreams())
      //
      this.peers[remotePeerId] = pc
      await pc.setRemoteDescription(sdp)
      const answer = await pc.createAnswer()
      console.log('-answer-', answer.sdp)
      await pc.setLocalDescription(answer)
      return answer
    }
  }

  onCandidate = async (remotePeerId, candidate) => {
    const pc = this.peers[remotePeerId]
    if (pc) {
      if (candidate) {
        const c = new RTCIceCandidate(candidate)
        await pc.addIceCandidate(c)
        console.log('AddIceCandidate success: ', c.protocol, c.address, c.port)
      }
    } else {
      console.error('invalid sdp message peer does not exist: ', remotePeerId)
    }
  }

  addLocalAudioTracks = async (streams) => {
    const ids = Object.keys(this.peers)
    const pc = this.peers[ids[0]] //temp

    this.tracks = [] //todo some and remove prev tracks
    const str = new MediaStream()
    streams.forEach(([stream, name]) => {
      const newTrack = stream.getTracks()[0]
      str.addTrack(newTrack)
      pc.addTrack(newTrack)
      this.tracks.push(newTrack)
      this.clients.push({ name: `fake - ${name}`, id: newTrack.id })
      console.log('-add file tack-', newTrack)
    })

    console.log('-local streams-', pc.getLocalStreams())

    const offer = await pc.createOffer()
    console.log('Offer from pc \n' + offer.sdp)
    await pc.setLocalDescription(offer)
    this.refreshUI()
    return [pc.peerId, offer]
  }
  dropClient = async (id) => {
    const ids = Object.keys(this.peers)
    const pc = this.peers[ids[0]] //temp
    this.clients = this.clients.filter(c => c.id !== id)
    this.tracks = this.tracks.filter(t => t.id !== id)
    const sender = pc.getSenders().find(sn => sn.track.id == id);
    if (sender) {
      pc.removeTrack(sender)
      console.log('-track removed-', id)
    }

    const offer = await pc.createOffer()
    console.log('Offer from pc \n' + offer.sdp)
    await pc.setLocalDescription(offer)
    this.refreshUI()
    return [pc.peerId, offer]
  }
  getMicrophoneTrack = async () => {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    const track = stream.getTracks()[0]
    if (track) {
      //mute mic by default
      //https://developer.mozilla.org/en-US/docs/Web/API/MediaStreamTrack/enabled
      track.enabled = false
      this.mic = track
      return this.mic
    }
  }
  toggleMicrophoneMute() {
    if (this.mic) {
      this.mic.enabled = !this.mic.enabled
      const ids = Object.keys(this.peers)
      const pc = this.peers[ids[0]] //temp
      const sender = pc.getSenders().find((s) => {
        return s.track.kind == 'audio';
      });
      console.log('found sender:', sender);
      sender && sender.replaceTrack(this.mic);
    }
  }

  getBitrates = async () => {
    let bitrateTxt = 'No bitrate stats'
    if (!this.pc) {
      return bitrateTxt
    }
    const stats = await this.pc.getStats()
    const results = stats.result();
    for (const i in results) {
      const result = results[i];
      if (!result.local || result.local === result) {
        if (result.type === 'ssrc') {
          const bytesNow = result.stat('bytesReceived');
          if (this.timestampPrev > 0) {
            const bitrate = Math.round((bytesNow - this.bytesPrev) * 8 / (result.timestamp - this.timestampPrev));
            if (bitrate > 0) {
              bitrateTxt = 'Received in ' + bitrate + ' kbits/sec';
            }
          }
          this.timestampPrev = result.timestamp;
          this.bytesPrev = bytesNow;
        }
      }
    }
    return bitrateTxt
  }
}
