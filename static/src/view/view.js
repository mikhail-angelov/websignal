import { html } from '../../libs/lit-html/lit-html.js'
import { styleMap } from '../../libs/lit-html/directives/style-map.js'
import './login.js'
import { header } from './header.js'
import { users } from './users.js'
import { messages } from './messages.js'
import { audioBind } from './audioBind.js'

export const view = (data, store) => html`
  ${data.authenticated
    ? html`<div style=${styleMap(styles.container)}>
          ${header(data, store)}
          ${data.room ? html`<div style=${styleMap(styles.main)}>
            ${users(data, store)}
            ${messages(data, store)}
            ${store.webrtc.streams.map(stream => html`<audio autoplay test=${audioBind(stream.stream)}></audio>`)}
          </div>`: null}
        </div>`
    : html`<login-form />`}
`
const styles = {
  container: {
    display: 'flex',
    flexDirection: 'column',
    height: '100%',
    width: '100%',
  },
  main: {
    display: 'flex',
    flexDirection: 'column',
    height: '100%',
    width: '100%',
  }
}