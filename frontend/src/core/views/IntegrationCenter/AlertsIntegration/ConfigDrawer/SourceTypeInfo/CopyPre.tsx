/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import CopyButton from 'src/core/components/CopyButton'

const CopyPre = ({ code }) => {
  return (
    <div className="relative">
      <pre className="text-xs bg-[#161b22]">{code}</pre>
      <div className="absolute right-5 top-2">
        <CopyButton value={code} iconText="COPY"></CopyButton>
      </div>
    </div>
  )
}
export default CopyPre
