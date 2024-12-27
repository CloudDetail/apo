import React from 'react'
import { MdHexagon } from 'react-icons/md'
import { BsArrowRight } from 'react-icons/bs'
export default function SingleLinkTopology() {
  return (
    <div className="flex flex-row items-center justify-center">
      <div className="p-2 flex items-center justify-center hover:cursor-pointer">
        <MdHexagon color="red" size={20} /> node1
      </div>
      <BsArrowRight size={26} />
      <div className="p-2 flex items-center justify-center hover:cursor-pointer">
        <MdHexagon color="red" size={20} /> node2
      </div>
    </div>
  )
}
