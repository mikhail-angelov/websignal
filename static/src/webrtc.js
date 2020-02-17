// const getScreenShareStream = () => navigator.getDisplayMedia({ video: true })
const stun = { urls: ['stun:stun.l.google.com:19302', 'stun:stun3.l.google.com:19302'] }
const turns = [
  {
    urls: 'turn:167.99.235.114:3478',
    credential: 'crosstec',
    username: 'crosstec',
  }]
const servers = [stun, ...turns]

export class WebRTC {
  constructor(getVideoElement, getRemoteVideoElement, updateStatus, refreshUI) {
    this.peers = {}
    this.localStream = null
    this.getVideoElement = getVideoElement
    this.getRemoteVideoElement = getRemoteVideoElement
    this.updateStatus = updateStatus
    this.refreshUI = refreshUI
    this.mic = null
    this.blankTrack = this.getBlankAudioTrack()
    this.tracks = []
    this.analysers = []
    this.trackId = 0
    this.streams = []
    this.clients = []
    this.broadcaster = false
    this.pendingUser = null
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
    this.analysers = []
    this.tracks = []
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
  createPeer(remotePeerId, sendCandidate, getUsers) {
    const pc = new RTCPeerConnection({ iceServers: servers })
    pc.peerId = remotePeerId
    pc.onicecandidate = ({ candidate }) => sendCandidate(remotePeerId, candidate)
    pc.ontrack = (event) => {
      console.log('-on-track-', pc.peerId, event)
      if (event.track) {
        const users = getUsers()
        if (users.some(user => user.id === event.track.id)) {
          const stream = new MediaStream()
          stream.addTrack(event.track)
          this.streams.push({ stream, id: event.track.id })
          console.log('Received remote track:')
          console.log(event.track.id)
          console.log(event.track.label)
          // this.addAnalyser(stream)
          event.track.onended = () => {
            // for any streams in e.stream check if their active flag is false
            console.log('-Track is closed:', event.track.id)
          }
          event.track.onmute = () => {
            // for any streams in e.stream check if their active flag is false
            console.log('-Track is onmute (but probably it is closed, and should be removed from peer):', event.track.id)
            //clean it
            this.streams = this.streams.filter(s => s.id !== event.track.id)
          }
        }

      } else if (event.streams && event.streams.length > 0) {
        this.streams.push({ stream: event.streams[0], id: this.pendingUserId })
        this.pendingUserId = null
        // this.addAnalyser(event.streams[0])
        console.log('Received remote stream')
      }
      console.log('-stream summary:', this.streams.length)
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
  connectPeer = async (id, remotePeerId, sendCandidate, getUsers) => {
    await this.start()
    const pc = this.createPeer(remotePeerId, sendCandidate, getUsers)
    this.peers[remotePeerId] = pc
    this.broadcaster = true
    //add main track
    let mainTrack = this.blankTrack || this.mic
    pc.addTrack(mainTrack)

    //and other tracks
    if (this.tracks.length > 0) {
      this.tracks.forEach(track => pc.addTrack(track))
    }

    console.log('-local streams-', pc.getSenders())
    //
    const offer = await pc.createOffer()
    console.log('Offer from pc \n' + offer.sdp)
    await pc.setLocalDescription(offer)
    offer.id = id
    return offer
  }

  onSDP = async (sdp, remotePeerId, sendCandidate, getUsers) => {
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
      let newPeer = true
      if (this.peers[remotePeerId]) {
        console.log('-peer is already exist-', remotePeerId)
        newPeer = false
      }
      const pc = this.peers[remotePeerId] || this.createPeer(remotePeerId, sendCandidate, getUsers)

      // pc.getSenders().forEach(sender => {
      //   console.log('-track removed-', sender.track)
      //   pc.removeTrack(sender)
      // })
      // pc.getReceivers().forEach(receiver => {
      //   console.log('-track removed-', receiver.track)
      //   receiver.track && receiver.track.stop()
      // })

      if (newPeer) {
        //add mic tracks
        // const micTrack = await this.getMicrophoneTrack()
        let mainTrack = this.blankTrack || this.mic
        pc.addTrack(mainTrack)
      }
      console.log('-local streams-', pc.getReceivers())
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
      if (candidate && candidate.candidate) {
        const c = new RTCIceCandidate(candidate)
        await pc.addIceCandidate(c)
        console.log('AddIceCandidate success: ', c.protocol, c.address, c.port)
      }
    } else {
      console.error('invalid sdp message peer does not exist: ', remotePeerId)
    }
  }

  addLocalAudioTracks = (tracks) => {
    const ids = Object.keys(this.peers)
    this.tracks = [] //todo some and remove prev tracks   
    tracks.forEach(([track, name]) => {
      this.tracks.push(track)
      this.clients.push({ name: `fake - ${name}`, id: track.id })
    })
    for (let id of ids) {
      const pc = this.peers[id]

      tracks.forEach(([track]) => {
        pc.addTrack(track)
        console.log('-add file tack-', track, id)
      })
    }
    this.refreshUI()
  }
  refreshPeersSdp = async (sender) => {
    const ids = Object.keys(this.peers)
    for (let id of ids) {
      const pc = this.peers[id]
      const offer = await pc.createOffer()
      console.log(`Offer from pc ${id}\n` + offer.sdp)
      await pc.setLocalDescription(offer)
      sender(id, offer)
    }
  }

  dropClient = async (trackId, resendSDP) => {
    const ids = Object.keys(this.peers)
    //user id and track id is the same for fake users
    const track = this.tracks.find(t => t.id === trackId)
    if (track) {
      track.stop()
    }
    for (let id of ids) {
      const pc = this.peers[id]
      const sender = pc.getSenders().find(sn => sn.track && sn.track.id == trackId);
      if (sender) {
        sender.track && sender.track.stop()
        pc.removeTrack(sender)
        console.log('-track removed-', trackId)
      }
      console.log('-list all senders-', pc.getSenders())

      this.clients = this.clients.filter(c => c.id !== trackId)
      this.tracks = this.tracks.filter(t => t.id !== trackId)

      const offer = await pc.createOffer()
      console.log('Offer from pc \n' + offer.sdp)
      await pc.setLocalDescription(offer)
      resendSDP(pc.peerId, offer)
    }
  }
  getMicrophoneTrack = async () => {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    const track = stream.getAudioTracks()[0]
    if (track) {
      //mute mic by default
      //https://developer.mozilla.org/en-US/docs/Web/API/MediaStreamTrack/enabled
      // track.enabled = false
      console.log('get mic track', track.id, track.label)
    }
    this.mic = track
    return track
  }
  toggleMicrophoneMute = async () => {
    let currentTrack = null
    let newTrack = null
    if (this.mic) {
      this.blankTrack = this.getBlankAudioTrack()
      currentTrack = this.mic
      newTrack = this.blankTrack
      this.mic = null
    } else {
      this.mic = await this.getMicrophoneTrack()
      newTrack = this.mic
      currentTrack = this.blankTrack
      this.blankTrack = null
    }
    const ids = Object.keys(this.peers)
    ids.forEach(id => {
      const pc = this.peers[id]
      const sender = pc.getSenders().find((s) => {
        return s.track && s.track.id === currentTrack.id;
      });
      console.log('found local sender:', sender);
      sender && sender.replaceTrack(newTrack);
    })
    currentTrack.stop()
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
  addAnalyser = stream => {
    const audioCtx = new AudioContext();
    const analyser = audioCtx.createAnalyser();
    analyser.minDecibels = -90;
    analyser.maxDecibels = -10;
    analyser.smoothingTimeConstant = 0.85;
    analyser.fftSize = 256;

    const source = audioCtx.createMediaStreamSource(stream);
    const gainNode = audioCtx.createGain();
    source.connect(gainNode);
    gainNode.connect(analyser);
    // analyser.connect(audioCtx.destination);

    this.analysers.push(analyser)

  }
  getSoundIndicator = index => {
    const analyser = this.analysers[index]
    if (analyser) {
      const bufferLength = analyser.fftSize;
      // console.log(bufferLength);
      const dataArray = new Uint8Array(bufferLength);
      analyser.getByteTimeDomainData(dataArray);
      // console.log('-- gain: ', index, getAverageVolume(dataArray), dataArray)
    }
  }
  getAudioFromFileStream = file => {
    return new Promise((resolve) => {
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
          console.log('-create stream from file-', destination.stream)
          const track = destination.stream.getTracks()[0]
          track.contentHint = file.name
          resolve([track, file.name])
        });
      });

      reader.readAsArrayBuffer(file);
    })
  }
  getBlankAudioTrack = () => {
    const context = new AudioContext();
    const gainNode = context.createGain();
    gainNode.connect(context.destination);
    // don't play for self
    gainNode.gain.value = 0;
    // Create the sound source
    const soundSource = context.createBufferSource();
    // Create an empty three-second stereo buffer at the sample rate of the AudioContext
    const blankArrayBuffer = context.createBuffer(2, context.sampleRate * 3, context.sampleRate);


    soundSource.buffer = blankArrayBuffer;
    soundSource.start(0, 0 / 1000);
    soundSource.connect(gainNode);

    const destination = context.createMediaStreamDestination();
    soundSource.connect(destination);

    // destination.stream
    console.log('-create blank track-', destination.stream)
    const track = destination.stream.getTracks()[0]
    track.other = "test-" + this.broadcaster
    return track
  }
  setPendingUserId = id => {
    this.pendingUserId = id
  }
}

function getAverageVolume(array) {
  var values = 0;
  var average;

  var length = array.length;

  // get all the frequency amplitudes
  for (var i = 0; i < length; i++) {
    values += Math.abs(array[i] - 128);
  }

  average = values / length;
  return average;
}