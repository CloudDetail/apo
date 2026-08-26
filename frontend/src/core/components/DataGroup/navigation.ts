interface RouteLocationState {
  pathname: string
  hash: string
  state: unknown
}

interface SearchParamsValue {
  toString: () => string
}

export function buildDataGroupNavigation(
  location: RouteLocationState,
  searchParams: SearchParamsValue,
) {
  const search = searchParams.toString()
  return {
    to: {
      pathname: location.pathname,
      search: search ? `?${search}` : '',
      hash: location.hash,
    },
    options: {
      replace: true as const,
      state: location.state,
    },
  }
}
