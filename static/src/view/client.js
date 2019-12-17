import { html } from '../../libs/lit-html/lit-html.js'

export const client = (data, store) => html`
  <style>
    .local-video {
      width: 100px;
      height: 66px;
    }
    .remote-video {
      width: 640px;
      height: 480px;
    }
  </style>
  <div><a href=${data.conferenceLink} target="_blank">${data.conferenceLink}</a></div>
  <video id="video-remote" class="remote-video" autoplay></video>
  <video id="video" class="local-video" autoplay></video>
`
