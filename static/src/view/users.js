import { html } from '../../libs/lit-html/lit-html.js'
import { styleMap } from '../../libs/lit-html/directives/style-map.js'

const User = (user, fakeUsers, store) => html`
  <div style=${styleMap(styles.user)}>
    <img style=${styleMap(styles.avatar)} src=${user.picture} />
    <div style=${styleMap(styles.info)}>${user.name}</div>
    ${fakeUsers.includes(user.id) ?
    html`<button style=${styleMap(styles.closeUser)} @click=${() => store.dropClient(user.id)}>X</button>` :
    null}
  <div>
`
const You = (data) => html`
  <div style=${styleMap(styles.user)}>
    <img style=${styleMap(styles.avatar)} src=${data.avatar} />
    <div style=${styleMap(styles.info)}>${data.username + '(you)'}</div>
    <div><a href=${data.conferenceLink} target="_blank">${data.conferenceLink}</a></div>
  <div>
`
const AddFake = (store) => html`
  <div style=${styleMap(styles.user)}>
    <div style=${styleMap(styles.plus)} @click=${() => { document.getElementById("add-fake").click() }}>+</div>
    <input id="add-fake" type="file" @input=${store.onMp3FileLoad} accept=".mp3" style=${styleMap(styles.fileInput)} />
    <div style=${styleMap(styles.info)}>Add fake user</div>
  <div>
`

export const users = (data, store) => html`
  <div style=${styleMap(styles.container)}>
    ${data.broadcaster ? You(data) : null}
    ${data.users.map((user) => User(user, data.fakeUsers, store))}
    ${data.broadcaster ? AddFake(store) : null}
  </div>
`

const styles = {
  container: {
    display: 'flex',
    flexWrap: 'wrap',
    alignContent: 'flex-start',
    flex: 1,
    backgroundColor: 'lightblue',
    height: '100%',
  },
  fileInput: {
    position: 'absolute',
    top: '-100px',
    visibility: 'hidden'
  },
  user: {
    position: 'relative',
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'center',
    alignItems: 'center',
    width: '150px',
    height: '140px',
    border: '1px solid gray',
    borderRadius: '8px',
    background: 'aliceblue',
    margin: '10px',
    boxShadow: '2px 2px gray',
    textAlign: 'center',
  },
  avatar: {
    width: '24px',
    margin: '0 8px',
  },
  info: {
    marginTop: '8px',
  },
  plus: {
    fontSize: '36px',
  },
  closeUser: {
    position: 'absolute',
    top: 0,
    right: 0,
    borderRadius: '50%',
    height: '17px',
    width: '17px',
    padding: 0,
    fontSize: '10px'
  }
}
