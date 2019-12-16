import { html } from '../libs/lit-html/lit-html.js'

export const view = (data, store) => html`
  ${data.authenticated
    ? html`
        <div id="with-auth">
          <div>${data.username}</div>
          <img style="width: 50px;" src=${data.avatar} />
          <form method="post" action="auth/local/logout">
            <button>Logout</button>
          </form>
          <button id="toggle-conference" @click=${store.onToggleConference}>
            ${data.room ? 'stop' : 'start'} conference
          </button>
          ${data.room
            ? html`
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
            : ''}
          <div>
            ${data.messages.map(
              m =>
                html`
                  <div>${m.data}</div>
                `
            )}
          </div>
          <input .value=${data.message} @input=${store.onMessageChange} />
          <button id="send" @click=${store.onSendMessage}>send</button>
        </div>
      `
    : html`
        <div id="no-auth">
          <form method="post" action="auth/local/login?from=/">
            <input name="email" />
            <input name="password" type="password" />
            <button>Login</button>
          </form>
          <form id="yandex" method="post" action="auth/yandex/login?from=/">
            <button>Yandex</button>
          </form>
          <form id="github" method="post" action="auth/github/login?from=/">
            <button>Github</button>
          </form>
          <form id="google" method="post" action="auth/google/login?from=/">
            <button>Google</button>
          </form>
        </div>
      `}
`