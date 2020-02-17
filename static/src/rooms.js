async function getRooms() {
  const res = await fetch('/api/room')
  if (!res.ok) {
    console.log('error fetch', res.status)
    return
  }
  const data = await res.json()
  console.log('rooms:', data)
  return data
}

async function joinRoom(id) {
  const res = await fetch(`/api/room/join/${id}`, { method: 'POST' })
  if (!res.ok) {
    console.log('error fetch', res.status)
    return
  }
  const data = await res.json()
  console.log('room:', data)
  return data
}


async function createRoom() {
  const res = await fetch('/api/room', { method: 'POST' })
  if (!res.ok) {
    console.log('error fetch', res.status)
    return
  }
  const data = await res.json()
  console.log('createRoom:', data)
  return data
}

export { getRooms, joinRoom, createRoom }
