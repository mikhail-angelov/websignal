const getVideoElement = () => document.getElementById('video')
const getRemoteVideoElement = () => document.getElementById('video-remote')
const getScreenShareStream = () => navigator.getDisplayMedia({ video: true })

const servers = []
let pc = null //webrtc peer
let localStream = null //local stream
let remoteStream = null //remote stream

export const start = async () => {
  localStream = await navigator.mediaDevices.getUserMedia({ video: true })
  const video = getVideoElement()
  video.srcObject = localStream
}

export const stop = () => {
  console.log('Ending call')
  if (localStream) {
    localStream.getTracks().forEach(track => track.stop())
  }
  if (pc) {
    pc.close()
    pc = null
  }
}

export const connectPeer = async (id, remotePeerId, sendCandidate) => {
  await start()
  pc = new RTCPeerConnection(servers)
  pc.onicecandidate = async ({candidate}) => {
    sendCandidate(candidate, remotePeerId)
  }
  pc.ontrack = e => {
    const remoteVideo = getRemoteVideoElement()
    if (remoteVideo && remoteVideo.srcObject !== e.streams[0]) {
      remoteVideo.srcObject = e.streams[0]
      console.log('Received remote stream')
    }
  }
  localStream.getTracks().forEach(track => pc.addTrack(track, localStream))
  const offer = await pc.createOffer()
  console.log('Offer from pc \n' + offer.sdp)
  await pc.setLocalDescription(offer)

  offer.id = id
  return offer
}

export const onSDP = async (sdp, sendCandidate) => {
  if (sdp.type === 'answer') {
    //process answer
    pc.setRemoteDescription(sdp)
  } else  if(sdp.type === 'offer'){
    //process offer
    pc = new RTCPeerConnection(servers)
    pc.onicecandidate = async ({candidate}) => {
      sendCandidate(candidate, sdp.id)
    }
    pc.oniceconnectionstatechange = event => {
      console.log('oniceconnectionstatechange:',event)
    }
    pc.ontrack = e => {
      const remoteVideo = getRemoteVideoElement()
      if (remoteVideo && remoteVideo.srcObject !== e.streams[0]) {
        remoteVideo.srcObject = e.streams[0]
        console.log('Received remote stream')
      }
    }
    localStream.getTracks().forEach(track => pc.addTrack(track, localStream))
    await pc.setRemoteDescription(sdp)
    await pc.setLocalDescription(await pc.createAnswer())

    pc.setRemoteDescription(sdp)
    const answer = await pc.createAnswer()
    await pc.setLocalDescription(answer)
    return answer
  }
}

export const onCandidate = async msg => {
  if (pc && msg.data) {
    await pc.addIceCandidate(new RTCIceCandidate(msg.data))
    console.log('AddIceCandidate success.')
  }
}
