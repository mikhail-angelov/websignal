const id = Math.random()
  .toString(36)
  .substring(2, 15)
const protocol = location.protocol === 'https:'?"wss:":"ws:"
const socket = new WebSocket(`${protocol}://${location.host}/ws?id=${id}&token=1234`)
socket.binaryType = 'arraybuffer' //to support binary messages
const messages = document.getElementById('main')
const textInput = document.getElementById('text')
const send = document.getElementById('send')
const encoder = new TextEncoder()
const decoder = new TextDecoder('utf-8')

send.addEventListener('click', () => {
  const data = textInput.value
  if (data) {
    const msg = JSON.stringify({ data, from: id, type: 'text' })
    socket.send(encoder.encode(msg))
    textInput.value = ''
  }
})
console.log('Attempting Connection...')

socket.onopen = () => {
  console.log('Successfully Connected')
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
