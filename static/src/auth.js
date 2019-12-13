const getAuth = async () => {
  const res = await fetch('/auth/user')
  if (!res.ok) {
    console.log('error fetch', res.status)
    return
  }
  const data = await res.json()

  let avatar = ''
  if (data.picture) {
    avatar = 'data:image/png;base64,' + data.picture
  }
  if (data.pictureUrl) {
    avatar = data.pictureUrl
  }
  const token = getJwt()
  return [token, { ...data, avatar }]
}

function getJwt() {
  const cookies = document.cookie.split('; ').reduce((c, x) => {
    const splitted = x.split('=')
    c[splitted[0]] = splitted[1]
    return c
  }, {})
  return cookies['jwt']
}

export { getAuth }
