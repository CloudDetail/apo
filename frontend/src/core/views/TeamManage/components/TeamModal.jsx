import React from 'react'
import { Modal } from 'antd'

const TeamModal = (props) => {
  const { open, closeModal } = props
  return <Modal open={open} onCancel={closeModal}></Modal>
}

export default TeamModal
