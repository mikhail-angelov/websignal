import { html } from '../../libs/lit-html/lit-html.js'
import { styleMap } from '../../libs/lit-html/directives/style-map.js'

const Message = (m) => html`
  <div>${m.text}</div>
`

export const messages = (data, store) => data.hideMessages ?
  html`
  <div style=${styleMap(styles.hiddenContainer)}>
    <div style=${styleMap(styles.toggleMessage)} @click=${() => store.toggleMessages()}>
      Show messages
    </div>
  </div>
  `:
  html`
    <div style=${styleMap(styles.container)}>
      <div style=${styleMap(styles.toggleMessage)} @click=${() => store.toggleMessages()}>
        Hide messages
      </div>
      <div style=${styleMap(styles.list)}>
        ${data.messages.map(m => Message(m))}
      </div>
      <br/>
      <div style=${styleMap(styles.inputBox)}>
        <input .value=${data.message} @input=${store.onMessageChange} />
        <button id="send" @click=${store.onSendMessage}>send</button>
      </div>
    </div>
    `
const styles = {
  hiddenContainer: {
    display: 'flex',
  },
  container: {
    display: 'flex',
    flexDirection: 'column',
    flex: 1,
  },
  toggleMessage: {
    backgroundColor: 'lightsteelblue',
    display: 'flex',
    width: '100%',
    justifyContent: 'center',
    alignItems: 'center',
    height: '30px',
  },
  list: {
    display: 'flex',
    height: '100%',
    flexDirection: 'column',
    overflow: 'auto',
  },
  inputBox: {
    display: 'flex',
  }
}