const getScreenShareStream = () => navigator.getDisplayMedia({ video: true })

const servers = []

export class WebRTC {
  constructor(getVideoElement, getRemoteVideoElement, updateStatus) {
    this.peers = {}
    this.localStream = null
    this.getVideoElement = getVideoElement
    this.getRemoteVideoElement = getRemoteVideoElement
    this.updateStatus = updateStatus
  }

  static start = async () => {
    this.localStream = await navigator.mediaDevices.getUserMedia({ video: true })
    const video = this.getVideoElement()
    video.srcObject = this.localStream
  }

  stop = () => {
    console.log('Ending call')
    if (this.localStream) {
      this.localStream.getTracks().forEach((track) => track.stop())
      this.localStream = null
    }
    Object.keys(this.peers).forEach((id) => {
      const pc = this.peers[id]
      pc.removeTrack(pc.sender)
      pc.close()
      const video = this.getVideoElement()
      if (video) video.srcObject = null
      const remoteVideo = this.getRemoteVideoElement()
      if (remoteVideo) remoteVideo.srcObject = null
    })
    this.peers = {}
  }
  createPeer(remotePeerId, sendCandidate) {
    const pc = new RTCPeerConnection(servers)
    let track,
      tr = null
    pc.peerId = remotePeerId
    pc.onicecandidate = ({ candidate }) => sendCandidate(remotePeerId, candidate)
    pc.ontrack = (event) => {
      const remoteVideo = this.getRemoteVideoElement()
      if (remoteVideo && remoteVideo.srcObject !== event.streams[0]) {
        remoteVideo.srcObject = event.streams[0]
        tr = event.transceiver
        track = event.track
        console.log('Received remote stream')
      }
    }
    pc.oniceconnectionstatechange = (event) => {
      console.log('oniceconnectionstatechange:', event)
      if (event.currentTarget.iceConnectionState === 'closed') {
        this.updateStatus(remotePeerId, 'disconnected', tr)
        track.stop() //need investigate it
      }
    }
    this.localStream.getTracks().forEach((track) => {
      pc.sender = pc.addTrack(track, this.localStream)
    })
    return pc
  }

  connectPeer = async (id, remotePeerId, sendCandidate) => {
    await this.start()
    const pc = this.createPeer(remotePeerId, sendCandidate)
    this.peers[remotePeerId] = pc
    const offer = await pc.createOffer()
    console.log('Offer from pc \n' + offer.sdp)
    await pc.setLocalDescription(offer)
    offer.id = id
    return offer
  }

  onSDP = async (sdp, remotePeerId, sendCandidate) => {
    if (sdp.type === 'answer') {
      //process answer
      const pc = this.peers[remotePeerId]
      if (pc) {
        pc.setRemoteDescription(sdp)
      } else {
        console.error('invalid sdp message peer does not exist: ', remotePeerId)
      }
    } else if (sdp.type === 'offer') {
      const pc = this.createPeer(remotePeerId, sendCandidate)
      this.peers[remotePeerId] = pc
      await pc.setRemoteDescription(sdp)
      const answer = await pc.createAnswer()
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
}
