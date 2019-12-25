import { html } from '../../libs/lit-html/lit-html.js'

export const broadcaster = (data, store) => html`
  <style>
    .container {
      display: flex;
      flex-direction: column;
    }
    .main {
      width: 640px;
      height: 480px;
    }
  </style>
  <div><a href=${data.conferenceLink} target="_blank">${data.conferenceLink}</a></div>
  <div class="container">
    <button @click=${store.onMute}>Mute</button>
    <input id="add-fake" type="file" @input=${store.onMp3FileLoad} accept=".mp3" multiple style="position:absolute;top:-100px"></input>
    <button @click=${()=>{document.getElementById("add-fake").click()}}>Add fake client</button>
    ${store.webrtc.clients.map((client) => html`<div><button @click=${()=>store.dropClient(client.id)}>X</button>${client.name}<div>`)}
    <video id="video" class="main" autoplay></video>
  </div>
`
