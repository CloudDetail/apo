import React from 'react'
import ossRoutes from './oss/routes'

const baseRoutes = [{ path: '/', exact: true, name: 'Home' }]
const routes = [...baseRoutes, ...ossRoutes]
export default routes
