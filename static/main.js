const id = Math.random()
  .toString(36)
  .substring(2, 15)
const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
const noAuth = document.getElementById('no-auth')
const withAuth = document.getElementById('with-auth')
const user = document.getElementById('user')
const avatar = document.getElementById('avatar')
const messages = document.getElementById('main')
const textInput = document.getElementById('text')
const send = document.getElementById('send')
const yandex = document.getElementById('yandex')
const logout = document.getElementById('logout')
const encoder = new TextEncoder()
const decoder = new TextDecoder('utf-8')
const TEXT_TYPE = 0
let socket = null

const init = async () => {
  noAuth.style = 'display:block;'
  withAuth.style = 'display:none;'
  try {
    const res = await fetch('/auth/user')
    if (!res.ok) {
      console.log('error fetch', res.status)
      return
    }
    const data = await res.json()
    withAuth.style = 'display:block;'
    noAuth.style = 'display:none;'
    user.textContent = data.name
    if (data.picture) {
      avatar.src = 'data:image/png;base64,' + data.picture
    }
    if (data.pictureUrl) {
      avatar.src = data.pictureUrl
    }
    const token = getJwt()
    connect(token)
    console.log('user:', data)
  } catch (e) {
    console.log('invalid auth:', e)
  }
}

init()

send.addEventListener('click', () => {
  const data = textInput.value
  if (data) {
    const msg = JSON.stringify({ data, from: id, type: TEXT_TYPE })
    socket.send(encoder.encode(msg))
    textInput.value = ''
  }
})

function connect(token) {
  socket = new WebSocket(`${protocol}//${location.host}/ws?id=${id}&token=${token}`)
  socket.binaryType = 'arraybuffer' //to support binary messages

  console.log('Attempting Connection...')

  socket.onopen = () => {
    console.log('Successfully Connected')
    //get Rooms
    getRooms()
  }

  socket.onclose = event => {
    console.log('Socket Closed Connection: ', event)
  }

  socket.onerror = error => {
    console.log('Socket Error: ', error)
  }

  socket.onmessage = event => {
    try {
      console.log('Socket on message ', event)
      const message = JSON.parse(decoder.decode(new Uint8Array(event.data).buffer))
      console.log('new message', message)
      const newMessage = document.createElement('div')
      newMessage.textContent = message.data
      messages.appendChild(newMessage)
    } catch (e) {
      console.log('onmessage error', e)
    }
  }
}

async function getRooms(){
  const res = await fetch('/api/room')
    if (!res.ok) {
      console.log('error fetch', res.status)
      return
    }
    const data = await res.json()
    console.log('rooms:', data)
}

function getJwt() {
  const cookies = document.cookie.split('; ').reduce((c, x) => {
    const splitted = x.split('=')
    c[splitted[0]] = splitted[1]
    return c
  }, {})
  return cookies['jwt']
}
