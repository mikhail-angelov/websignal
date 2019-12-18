import { html } from '../../libs/lit-html/lit-html.js'

export const broadcaster = (data, store) => html`
  <style>
    .local-video {
      width: 640px;
      height: 480px;
    }
    .remote-video {
      width: 100px;
      height: 66px;
    }
  </style>
  <div><a href=${data.conferenceLink} target="_blank">${data.conferenceLink}</a></div>
  <div class="container">
    <video id="video-remote" class="remote-video" autoplay></video>
    <video id="video" class="local-video" autoplay></video>
  </div>
`
