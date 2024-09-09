// 初始化状态
const initialState = {
  startTime: null,
  endTime: null,
  service: '',
  instance: '',
  traceId: '',
  endpoint: '',
  instanceOption: {},
  namespace: '',
  faultTypeList: null,
  minDuration: null,
  maxDuration: null,
  // 待废弃
  duration: null,
  isSlow: null,
  isError: null,
}

const urlParamsReducer = (state = initialState, action) => {
  switch (action.type) {
    case 'setInstanceOption':
      return { ...state, instanceOption: action.payload }
    case 'setUrlParamsState':
      if (action.payload.service === '') {
        action.payload.instanceOption = {}
      }
      return { ...state, ...action.payload }
    case 'clearUrlParamsState':
      return { ...initialState }
    default:
      return state
  }
}
// const { modalDataUrl } = useSelector((state) => state.topologyReducer)
export default urlParamsReducer
