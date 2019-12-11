import { getAuth } from './auth.js'
import { getRooms } from './rooms.js'
import { getId } from './utils.js'
import { Connection, ONOPEN } from './connection.js'

const messages = document.getElementById('main')
const textInput = document.getElementById('text')
const send = document.getElementById('send')
const TEXT_TYPE = 0
let connection = null
let userId = null

async function init() {
  try {
    const [token, user] = await getAuth()
    userId = user.id
    connection = new Connection(token)
    connection.on(TEXT_TYPE, onTextMessage)
    connection.on(ONOPEN, onOpenConnection)
    connection.connect()
  } catch (e) {
    console.log('invalid auth:', e)
  }
}
init()

const onOpenConnection = () => {
  //get Rooms
  getRooms()
}

const onTextMessage = message => {
  console.log('new message', message)
  const newMessage = document.createElement('div')
  newMessage.textContent = message.data
  messages.appendChild(newMessage)
}

send.addEventListener('click', () => {
  const data = textInput.value
  if ((data, connection)) {
    connection.send({ data, from: userId, type: TEXT_TYPE })
    textInput.value = ''
  }
})
