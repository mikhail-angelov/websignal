async function getRooms() {
  const res = await fetch('/api/room')
  if (!res.ok) {
    console.log('error fetch', res.status)
    return
  }
  const data = await res.json()
  console.log('rooms:', data)
}

async function createRoom(id) {
  const res = await fetch('/api/room', { method: 'POST' })
  if (!res.ok) {
    console.log('error fetch', res.status)
    return
  }
  const data = await res.json()
  console.log('rooms:', data)
}

export { getRooms, createRoom }
