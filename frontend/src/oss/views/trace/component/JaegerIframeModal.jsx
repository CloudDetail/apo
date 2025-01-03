/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { CModal, CModalBody, CModalHeader, CModalTitle } from '@coreui/react'
import React from 'react'
function EndpointTableModal(props) {
  const { visible, traceId, closeModal } = props
  return (
    <CModal
      visible={visible}
      alignment="center"
      size="xl"
      className="absolute-modal"
      onClose={closeModal}
    >
      <CModalHeader>
        <CModalTitle>Jaeger Trace : {traceId}</CModalTitle>
      </CModalHeader>

      <CModalBody className="text-sm h-4/5">
        <iframe
          src={'/jaeger/trace/' + traceId + '?uiEmbed=v0'}
          width="100%"
          height="100%"
          frameBorder={0}
        ></iframe>
      </CModalBody>
    </CModal>
  )
}

export default EndpointTableModal
