import { html } from '../../libs/lit-html/lit-html.js'
import { styleMap } from '../../libs/lit-html/directives/style-map.js'

export const header = (data, store) => html`
  <div style=${styleMap(styles.header)}>
  <div style=${styleMap(styles.startBlock)}>
    <button @click=${store.onToggleConference}>
      ${data.room ? 'stop' : 'start'} conference
    </button>
    ${data.room ? html`
    <button @click=${store.onMute}>${data.muted ? 'unmute' : 'mute'}</button>` : null}
  </div>
    <div style=${styleMap(styles.name)} >${data.username}</div>
    <img style=${styleMap(styles.avatar)} src=${data.avatar} />
    <form style=${styleMap(styles.logout)} method="post" action="auth/local/logout">
      <button>Logout</button>
    </form>
  </div>
`
const styles = {
  header: {
    display: 'flex',
    height: '34px',
    backgroundColor: 'lightsteelblue',
    justifyContent: 'flex-end',
    alignItems: 'center',
  },
  startBlock: {
    display: 'flex',
    flex: 1,
    alignItems: 'center',
  },
  name: {
    margin: '0 8px',
  },
  avatar: {
    width: '32px',
    margin: '0 8px',
  },
  logout: {
    margin: 0,
  }
}