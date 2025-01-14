/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import { logsMock } from '../mock'
function HighLightCode(props) {
  const { timestamp } = props
  const text = logsMock
  return (
    <pre className="h-full w-full p-2 bg-[#141414] rounded border border-solid border-[#343436] overflow-y-auto overflow-x-hidden">
      {/* <div className="flex-center">
        <TextArea
          onChange={(e) => {
            setSearchWords([e.currentTarget.value])
          }}
          placeholder="搜索"
        />
        <Button
          disabled={count === 0 || searchWords.length === 0}
          fill="text"
          icon="arrow-up"
          onClick={() => changeActiveIndex('sub')}
        ></Button>
        <Button
          icon="arrow-down"
          disabled={count === 0 || searchWords.length === 0}
          fill="text"
          onClick={() => changeActiveIndex('add')}
        ></Button>
        <Button icon="cloud-download" fill="text" onClick={exportCode}></Button>
      </div> */}
      <div
        style={{ height: '100%', overflowY: 'auto', textWrap: 'pretty' }}
        className="text-[#ccccdc] leading-5 overflow-x-hidden"
      >
        {text}
      </div>
    </pre>
  )
}
export default HighLightCode
