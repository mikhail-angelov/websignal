const noAuth = document.getElementById('no-auth')
const withAuth = document.getElementById('with-auth')
const user = document.getElementById('user')
const avatar = document.getElementById('avatar')

const getAuth = async ()=>{
    noAuth.style = 'display:block;'
    withAuth.style = 'display:none;'
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
    return [token, data]
}

function getJwt() {
  const cookies = document.cookie.split('; ').reduce((c, x) => {
    const splitted = x.split('=')
    c[splitted[0]] = splitted[1]
    return c
  }, {})
  return cookies['jwt']
}

export {getAuth}