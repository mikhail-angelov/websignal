import { html} from '../../libs/lit-html/lit-html.js'
import { styleMap } from '../../libs/lit-html/directives/style-map.js'

export const video = (data, store) => html`
  <div style=${styleMap(styles.container)}>
    <video id="remote-0" style=${styleMap(styles.remoteVideo)}> autoplay></video>
    <button @click=${store.onRefresh}>Refresh peers</button>
  </div>
`

const styles = {
  container: {
    display: 'flex',
    flexDirection: 'column',
    height: '100%',
    flex: 1,
  },
  remoteVideo: {
    width: '240px',
    height: '180px',
  }
}