import React from 'react'
import ossRoutes from './oss/routes'
import proRoutes from './pro/routes'

const baseRoutes = [{ path: '/', exact: true, name: 'Home' }]
let routes = [...baseRoutes]
if (import.meta.env.VITE_APP_VERSION === 'pro') {
  routes = [...baseRoutes, ...proRoutes]
} else {
  routes = [...baseRoutes, ...ossRoutes, ...proRoutes]
}
console.log(import.meta.env.VITE_APP_VERSION, routes)
export default routes
