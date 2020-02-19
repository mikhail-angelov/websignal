import { styleMap } from '../../libs/lit-html/directives/style-map.js'
import { LitElement, css, html } from '../../libs/lit-element/lit-element.js'

class Login extends LitElement {
  static get styles() {
    return css`
      .yat:hover { 
        opacity: 0.5;
        cursor: pointer; 
      }
      .logPass:focus {
        outline: none;
        colore: transparent;
      }
      .logPass:hover { 
        opacity: 0.5;
      }
      .logPass::-webkit-input-placeholder {
        color:#fff;
        opacity:1;
      }
      .logPass::-moz-placeholder {
        color:#fff; 
        opacity:1;
      }
      .logPass:-ms-input-placeholder {
        color:#fff; 
        opacity:1;
      }
      .logPass:focus::-webkit-input-placeholder {
        color:#fff;
        opacity:0;
      }
      .logPassfocus::-moz-placeholder {
        color:#fff; 
        opacity:0;
      }
      .logPass:focus:-ms-input-placeholder {
        color:#fff; 
        opacity:0;
      }
      .btnLogin:hover { 
        opacity: 0.8;
        cursor: pointer; 
      }
      .btnLogin:focus {
        outline: none;
      }
      .btnSite:hover { 
        opacity: 0.5;
        cursor: pointer; 
      }
      .btnSite:focus {
        outline: none;
      }
    `;
  }
  render() {
    const from = '/' + location.search
    return html`
      <div id="no-auth" style=${styleMap(styles.login)}>
        <div style=${styleMap(styles.registrationField)}>
        <div style=${styleMap(styles.yat)} class="yat">&#1123
        </div>
        <form style=${styleMap(styles.inputFields)} method="post" action="auth/local/login?from=${from}" >
          <input style=${styleMap(styles.logPass)} class="logPass" name="email" placeholder="Email" autocomplete="off" />
          <input style=${styleMap(styles.logPass)} class="logPass" name="password" type="password" placeholder="Password" autocomplete="off" />
          <button style=${styleMap(styles.btnLogin)} class="btnLogin">Login</button>
        </form>
        <div style=${styleMap(styles.links)}>
          <form id="yandex" method="post" action="auth/yandex/login?from=${from}">
            <button style=${styleMap(styles.btnSite)} class="btnSite"><img src="../../assets/yandex.svg" alt="Yandex" /></button>
          </form>
          <form id="github" method="post" action="auth/github/login?from=${from}">
            <button style=${styleMap(styles.btnSite)} class="btnSite"><img src="../../assets/github.svg" alt="Github" /></button>
          </form>
          <form id="google" method="post" action="auth/google/login?from=${from}">
            <button style=${styleMap(styles.btnSite)} class="btnSite"><img src="../../assets/google.svg" alt="Google" /></button>
          </form>
        </div>
        </div>
      </div>
  `
  }
}

customElements.define('login-form', Login)

const backColor1 = '#6e45e2'
const backColor2 = '#88d3ce'
const styles = {
  login: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    background: `linear-gradient(45deg, ${backColor1}, ${backColor2})`,
    height: '100%',
  },
  registrationField: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    padding: '2em 3em',
    background: 'rgba(242,242,242,0.1)',
    borderRadius: '1em',
  },
  yat: {
    fontSize: '3em',
    fontFamily: '-webkit-pictograph',
    borderRadius: '50%',
    width: '58px',
    textAlign: 'center',
    border: '1px solid white',
    color: 'white',
  },
  inputFields: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
  },
  logPass: {
    padding: '1em',
    marginTop: '1em',
    background: 'transparent',
    border: 'none',
    borderBottom: '1px solid white',
    fontSize: '25px',
    textAlign: 'center',
    webkitInputPlaceholder: '#c0392b',
    
  },
  btnLogin: {
    padding: '1em',
    marginTop: '2em',
    width: '10em',
    border: 'none',
    borderRadius: '2em',
    background: 'rgb(255,253,208)',
    fontSize: '1em',
  },
  links: {
    display: 'flex',
    margin: '2em',
  },
  btnSite: {
    margin: '1em',
    borderRadius: '50%',
    padding: '1em',
    border: '1px solid white',
    background: 'transparent',
  },
}