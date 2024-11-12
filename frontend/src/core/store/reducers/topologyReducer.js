const initialState = {
  allTopologyData: null,
  //前五
  displayData: null,
  //modaldaya
  modalService: null,
  modalEndpoint: null,
  // 点击路径
  modalDataUrl: [],
}

const topologyReducer = (state = initialState, action) => {
  switch (action.type) {
    case 'setAllTopologyData':
      return { ...state, allTopologyData: action.payload }

    case 'setDisplayData':
      return { ...state, displayData: action.payload }

    case 'clearTopology':
      return { ...initialState }

    case 'setModalData':
      console.log(action.payload)

      return {
        ...state,
        modalService: action.payload.modalService,
        modalEndpoint: action.payload.modalEndpoint,
        displayData: null,
        modalDataUrl: action.payload.modalDataUrl
          ? action.payload.modalDataUrl
          : [
              ...state.modalDataUrl,
              {
                modalService: action.payload.modalService,
                modalEndpoint: action.payload.modalEndpoint,
              },
            ],
      }

    case 'rollback':
      const modalDataUrl = state.modalDataUrl.slice(0, -1)
      console.log({
        ...state,
        ...modalDataUrl,
        modalDataUrl: modalDataUrl,
      })
      return {
        ...state,
        displayData: null,
        ...modalDataUrl[modalDataUrl.length - 1],
        modalDataUrl: modalDataUrl,
      }

    default:
      return state
  }
}

export default topologyReducer
