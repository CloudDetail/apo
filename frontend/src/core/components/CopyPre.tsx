/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import CopyButton from 'src/core/components/CopyButton'

const CopyPre = ({ code,iconText="COPY" }) => {
  return (
    <div className="relative">
      <pre className="text-xs p-3 bg-[var(--ant-color-bg-elevated)]">{code}</pre>
      <div className="absolute right-2 top-2">
        <CopyButton value={code} iconText={iconText}></CopyButton>
      </div>
    </div>
  )
}
export default CopyPre
