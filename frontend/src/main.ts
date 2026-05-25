import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { VueQueryPlugin } from '@tanstack/vue-query'

import App from './App.vue'
import router from './router'
import './assets/main.css'

// Vue's bootstrap is like Angular's main.ts + AppModule bootstrapping combined.
// createApp()   → platformBrowserDynamic().bootstrapModule()
// use(router)   → RouterModule.forRoot()
// use(pinia)    → NgRx StoreModule.forRoot()
// use(VueQueryPlugin) → no Angular equivalent — replaces HttpClient + state management for server data

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(VueQueryPlugin)

app.mount('#app')
