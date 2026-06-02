/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

export const parentMenuNavigationTarget = (menu = {}) => (
  menu.children?.length > 0 && menu.router?.to ? menu.router.to : ''
)

export const parentMenuTitleClickProps = (menu = {}, navigate = () => {}) => {
  const target = parentMenuNavigationTarget(menu)
  if (!target) return {}

  return {
    onTitleClick: () => navigate(target),
  }
}
