import { html } from '../../libs/lit-html/lit-html.js'

export const login = () => {
  const from = '/'+location.search
  return html`
    <div id="no-auth">
      <form method="post" action="auth/local/login?from=${from}">
        <input name="email" />
        <input name="password" type="password" />
        <button>Login</button>
      </form>
      <form id="yandex" method="post" action="auth/yandex/login?from=${from}">
        <button>Yandex</button>
      </form>
      <form id="github" method="post" action="auth/github/login?from=${from}">
        <button>Github</button>
      </form>
      <form id="google" method="post" action="auth/google/login?from=${from}">
        <button>Google</button>
      </form>
    </div>
`
}
