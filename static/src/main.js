import { render } from '../libs/lit-html/lit-html.js'
import { view } from './view/view.js'
import { Store } from './store.js'

const store = new Store()
store.on(data => {
  render(view(data, store), document.body)
})
store.init()
