import { html} from '../../libs/lit-html/lit-html.js'
import {audioBind} from './audioBind.js'

export const client = (data, store) => html`
  <style>
    .container {
      position: relative;
    }
    .remote-video {
      width: 640px;
      height: 480px;
    }
    .local-video {
      position: absolute;
      bottom: 0;
      right: 0;
      z-index: 1;
      width: 64px;
      height: 48px;
    }
  </style>
  <div><a href=${data.conferenceLink} target="_blank">${data.conferenceLink}</a></div>
  <div class="container">
    <video id="remote-0" class="remote-video" autoplay></video>
    <button @click=${store.onMute}>Mute</button>
    <button @click=${store.onRefresh}>Refresh peers</button>
    ${store.webrtc.streams.map(stream => html`<audio controls autoplay test=${audioBind(stream)}></audio>`)}
  </div>
`
